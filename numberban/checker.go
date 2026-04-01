// Package numberban จัดการ logic เลขอั้น
//
// เลขอั้นคือเลขที่เจ้ามือไม่อยากรับแทง (เสี่ยงเกินไป) มี 2 แบบ:
//   - full_ban:    ไม่รับแทงเลขนี้เลย → ลูกค้าแทงไม่ได้
//   - reduce_rate: รับแทงแต่ลด rate จ่าย → ลูกค้าแทงได้แต่ได้เงินน้อยลง
//
// ความสัมพันธ์:
// - ใช้ types.NumberBan struct
// - ถูกเรียกโดย: standalone-member-api (#3), provider-game-api (#7)
// - flow: ลูกค้าแทง → Validate() → Check() (เช็คอั้น) → CheckLimit() → บันทึก
//
// Provider mode (#7) มีเลขอั้น 2 ระดับ:
//   - Global: admin provider ตั้ง → อั้นทุก operator
//   - Per-operator: operator ตั้งเอง → อั้นเฉพาะ operator นั้น
//
// Standalone mode (#3) มีเลขอั้นระดับเดียว (global)
package numberban

import (
	"github.com/farritpcz/lotto-core/types"
)

// CheckResult ผลการตรวจสอบเลขอั้น
type CheckResult struct {
	IsBanned    bool          // true = เลขถูกอั้น (ไม่ว่าจะ full หรือ reduce)
	BanType     types.BanType // ประเภทการอั้น (full_ban / reduce_rate)
	ReducedRate float64       // rate ที่ลดลง (ใช้เมื่อ BanType = reduce_rate, 0 ถ้า full_ban)
	MaxAmount   float64       // จำนวนเงินสูงสุดที่รับ (0 = ตาม default)
}

// Check ตรวจสอบว่าเลขนี้ถูกอั้นหรือไม่
//
// Parameters:
//   - number:     เลขที่จะตรวจ เช่น "847"
//   - betType:    ประเภทการแทง เช่น BetType3Top
//   - bans:       รายการเลขอั้นทั้งหมด (ดึงจาก DB/Redis cache)
//
// NOTE: แต่ละ API ต้องดึง bans มาเองจาก DB โดย filter ตาม:
//   - standalone (#3): WHERE lottery_type_id = ? AND (round_id = ? OR round_id IS NULL)
//   - provider (#7):   WHERE lottery_type_id = ? AND (round_id = ? OR round_id IS NULL)
//     AND (operator_id = ? OR operator_id IS NULL)
//
// Redis cache key pattern: "bans:{lotteryTypeID}:{roundID}" → []NumberBan
//
// ตัวอย่าง:
//
//	bans := []NumberBan{{Number: "847", BanType: BanTypeFull, ...}}
//	result := Check("847", BetType3Top, bans)
//	result.IsBanned → true
//	result.BanType  → "full_ban"
func Check(number string, betType types.BetType, bans []types.NumberBan) CheckResult {
	for _, ban := range bans {
		// ตรวจว่าเลขตรงกัน และประเภทการแทงตรงกัน
		if ban.Number == number && ban.BetType == betType {
			return CheckResult{
				IsBanned:    true,
				BanType:     ban.BanType,
				ReducedRate: ban.ReducedRate,
				MaxAmount:   ban.MaxAmount,
			}
		}
	}

	// ไม่ถูกอั้น
	return CheckResult{
		IsBanned: false,
	}
}

// GetEffectiveRate คืนค่า rate ที่ใช้จริง หลังจากเช็คเลขอั้นแล้ว
//
// ถ้าเลขถูกอั้นแบบ reduce_rate → ใช้ rate ที่ลดลง
// ถ้าไม่ถูกอั้น → ใช้ rate ปกติ
// ถ้าถูกอั้นแบบ full_ban → return 0 (ไม่ควรมาถึงจุดนี้ เพราะ Check() จะ block ก่อน)
//
// ตัวอย่าง:
//
//	result := Check("847", BetType3Top, bans)
//	effectiveRate := GetEffectiveRate(result, 900)
//	// ถ้า reduce_rate=500 → return 500
//	// ถ้าไม่อั้น → return 900
func GetEffectiveRate(checkResult CheckResult, originalRate float64) float64 {
	if !checkResult.IsBanned {
		// ไม่ถูกอั้น → ใช้ rate ปกติ
		return originalRate
	}

	if checkResult.BanType == types.BanTypeFull {
		// อั้นเต็ม → rate = 0 (ไม่ควรแทงได้)
		return 0
	}

	if checkResult.BanType == types.BanTypeReduceRate && checkResult.ReducedRate > 0 {
		// ลด rate → ใช้ rate ที่ลดลง
		return checkResult.ReducedRate
	}

	// fallback → ใช้ rate ปกติ
	return originalRate
}

// FilterBansForOperator กรอง bans สำหรับ operator เฉพาะ (provider mode)
//
// เลขอั้นมี 2 ระดับ:
//   - Global (operator_id = nil): อั้นทุก operator
//   - Per-operator (operator_id = X): อั้นเฉพาะ operator X
//
// function นี้รวมทั้ง global + per-operator เข้าด้วยกัน
// ถ้าเลขเดียวกันถูกอั้นทั้ง 2 ระดับ → ใช้ per-operator (เฉพาะเจาะจงกว่า)
//
// Parameters:
//   - allBans:    bans ทั้งหมด (ทั้ง global + per-operator)
//   - operatorID: ID ของ operator ที่ต้องการกรอง (nil = standalone mode)
//
// ตัวอย่าง (provider mode):
//
//	allBans := [...global bans..., ...operator-specific bans...]
//	filtered := FilterBansForOperator(allBans, &operatorID)
//	// ได้ bans ที่ operator นี้ต้องใช้ (global + ของตัวเอง)
func FilterBansForOperator(allBans []types.NumberBan, operatorID *int64) []types.NumberBan {
	if operatorID == nil {
		// standalone mode: ใช้เฉพาะ global bans (operator_id = nil)
		var globalBans []types.NumberBan
		for _, ban := range allBans {
			if ban.OperatorID == nil {
				globalBans = append(globalBans, ban)
			}
		}
		return globalBans
	}

	// provider mode: เก็บ global + per-operator
	// ถ้ามี per-operator override สำหรับเลขเดียวกัน → ใช้ per-operator
	perOperatorMap := make(map[string]types.NumberBan) // key: "number:betType"
	var globalBans []types.NumberBan

	for _, ban := range allBans {
		if ban.OperatorID == nil {
			// global ban
			globalBans = append(globalBans, ban)
		} else if *ban.OperatorID == *operatorID {
			// per-operator ban
			key := ban.Number + ":" + string(ban.BetType)
			perOperatorMap[key] = ban
		}
	}

	// รวม: global + per-operator (per-operator override global)
	var result []types.NumberBan
	for _, ban := range globalBans {
		key := ban.Number + ":" + string(ban.BetType)
		if operatorBan, exists := perOperatorMap[key]; exists {
			// per-operator override global
			result = append(result, operatorBan)
			delete(perOperatorMap, key)
		} else {
			result = append(result, ban)
		}
	}

	// เพิ่ม per-operator bans ที่ไม่มีใน global
	for _, ban := range perOperatorMap {
		result = append(result, ban)
	}

	return result
}
