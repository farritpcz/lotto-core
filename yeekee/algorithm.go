// Package yeekee จัดการ logic ยี่กี — ออกผลอัตโนมัติจากเลขที่สมาชิกยิงมา
//
// ยี่กีทำงานอย่างไร:
//  1. ระบบสร้างรอบอัตโนมัติทุก 15 นาที (88 รอบ/วัน) — cron job ใน API layer
//  2. รอบเปิดรับยิงเลข (shooting phase) — สมาชิกส่งเลข 5 หลัก ผ่าน WebSocket
//  3. หมดเวลา → ระบบ sum เลขทั้งหมด → mod ให้เหลือ 5 หลัก → ตัด 2-3 ตัวท้ายเป็นผล
//  4. เทียบผลกับ bets → จ่ายเงิน
//
// ความสัมพันธ์:
// - ใช้ types.YeekeeShoot, types.RoundResult
// - ถูกเรียกโดย:
//   - standalone-member-api (#3): cron job หมดเวลายี่กี → CalculateResult()
//   - provider-game-api (#7): cron job หมดเวลายี่กี → CalculateResult()
// - ผลที่ได้ส่งต่อไปยัง payout.MatchAll() เพื่อเทียบ bets
//
// ป้องกันการโกง:
// - เลขยิงต้องผ่าน betting.ValidateYeekeeShoot() (5 หลัก, ตัวเลขล้วน)
// - server-side validation เท่านั้น (ไม่เชื่อ client)
// - ทุกเลขบันทึก timestamp ลง DB ก่อนคำนวณ
package yeekee

import (
	"fmt"
	"strconv"

	"github.com/farritpcz/lotto-core/types"
)

// =============================================================================
// Algorithm — คำนวณผลยี่กีจากเลขที่ยิง
// =============================================================================

// CalculateResult คำนวณผลยี่กีจากเลขที่สมาชิกยิงมาทั้งหมด
//
// อัลกอริทึม:
//  1. sum เลขทั้งหมดที่ยิงมา (แต่ละเลข 5 หลัก เช่น 12345 + 67890 + ...)
//  2. ผลลัพธ์ mod 100000 ให้เหลือ 5 หลัก
//  3. จาก 5 หลัก ตัดเป็นผลรางวัล:
//     - 3 ตัวบน = 3 ตัวท้าย (หลักที่ 3-5)
//     - 2 ตัวบน = 2 ตัวท้าย (หลักที่ 4-5)
//     - 2 ตัวล่าง = หลักที่ 2-3
//
// Parameters:
//   - shoots: เลขที่สมาชิกยิงมาทั้งหมดในรอบนี้
//
// Returns:
//   - resultNumber: เลข 5 หลักผลลัพธ์ (เช่น "83456")
//   - roundResult:  ผลรางวัลที่ตัดแล้ว (Top3, Top2, Bottom2)
//   - error:        ถ้าไม่มีเลขยิง (ไม่สามารถคำนวณได้)
//
// ตัวอย่าง:
//
//	shoots := []YeekeeShoot{
//	    {Number: "12345"},
//	    {Number: "67890"},
//	    {Number: "11111"},
//	}
//	// sum = 12345 + 67890 + 11111 = 91346
//	// mod 100000 = 91346
//	// resultNumber = "91346"
//	// Top3 = "346", Top2 = "46", Bottom2 = "13"
func CalculateResult(shoots []types.YeekeeShoot) (resultNumber string, roundResult types.RoundResult, err error) {
	if len(shoots) == 0 {
		return "", types.RoundResult{}, types.ErrNoShoots
	}

	// Step 1: sum เลขทั้งหมด
	var totalSum int64
	for _, shoot := range shoots {
		num, parseErr := strconv.ParseInt(shoot.Number, 10, 64)
		if parseErr != nil {
			// ข้ามเลขที่ parse ไม่ได้ (ไม่ควรเกิดถ้า validate ก่อน)
			continue
		}
		totalSum += num
	}

	// Step 2: mod 100000 ให้เหลือ 5 หลัก
	result := totalSum % 100000

	// Step 3: format เป็น 5 หลัก (pad ด้วย 0 ข้างหน้า)
	resultNumber = fmt.Sprintf("%05d", result)

	// Step 4: ตัดเป็นผลรางวัล
	roundResult = ExtractResult(resultNumber)

	return resultNumber, roundResult, nil
}

// ExtractResult ตัดเลข 5 หลักเป็นผลรางวัล
//
// จาก 5 หลัก เช่น "83456":
//   - 3 ตัวบน (Top3)   = 3 ตัวท้าย    = "456" (index 2-4)
//   - 2 ตัวบน (Top2)   = 2 ตัวท้าย    = "56"  (index 3-4)
//   - 2 ตัวล่าง (Bottom2) = หลักที่ 2-3 = "34"  (index 1-2)
//
// ตัวอย่าง:
//
//	ExtractResult("83456") → RoundResult{Top3: "456", Top2: "56", Bottom2: "34"}
//	ExtractResult("00123") → RoundResult{Top3: "123", Top2: "23", Bottom2: "01"}
func ExtractResult(fiveDigit string) types.RoundResult {
	// ป้องกัน panic ถ้าเลขไม่ครบ 5 หลัก
	if len(fiveDigit) != 5 {
		return types.RoundResult{}
	}

	return types.RoundResult{
		Top3:    fiveDigit[2:5], // 3 ตัวท้าย
		Top2:    fiveDigit[3:5], // 2 ตัวท้าย
		Bottom2: fiveDigit[1:3], // หลักที่ 2-3
	}
}

// GetShootCount นับจำนวนเลขที่ยิงในรอบ
func GetShootCount(shoots []types.YeekeeShoot) int {
	return len(shoots)
}

// GetShootSum คำนวณผลรวมของเลขที่ยิง (สำหรับแสดงผลใน UI)
//
// ใช้ใน: WebSocket broadcast — แสดงผลรวมสะสม real-time ให้ลูกค้าเห็น
func GetShootSum(shoots []types.YeekeeShoot) int64 {
	var sum int64
	for _, shoot := range shoots {
		num, err := strconv.ParseInt(shoot.Number, 10, 64)
		if err != nil {
			continue
		}
		sum += num
	}
	return sum
}
