// Package lottery กำหนดกฎของหวยแต่ละประเภท
//
// ความสัมพันธ์:
// - ใช้ types.LotteryType, types.BetType
// - ถูกเรียกโดยทุก API เพื่อเช็คว่า:
//   - ประเภทหวยนี้รองรับ BetType ไหนบ้าง
//   - Default rate จ่ายเท่าไหร่
//   - schedule ออกผลอย่างไร
package lottery

import (
	"github.com/farritpcz/lotto-core/types"
)

// =============================================================================
// Rules — กฎหวยแต่ละประเภท
// =============================================================================

// LotteryRule กฎของหวย 1 ประเภท
type LotteryRule struct {
	Type            types.LotteryType // ประเภทหวย
	Name            string            // ชื่อแสดงผล
	AllowedBetTypes []types.BetType   // ประเภทการแทงที่รองรับ
	DefaultRates    map[types.BetType]float64 // rate จ่ายเริ่มต้น
	IsAutoResult    bool              // ระบบออกผลเองหรือไม่ (true = ยี่กี)
	Description     string            // คำอธิบาย
}

// DefaultRules กฎเริ่มต้นของหวยทุกประเภท
//
// NOTE: rate เหล่านี้เป็นค่า default — admin สามารถปรับได้ผ่าน admin panel
// operator (provider mode) ก็ตั้ง rate ของตัวเองได้ผ่าน dashboard
//
// rate หมายถึง: แทง 1 บาท ถ้าถูกได้กี่บาท
// เช่น rate 900 = แทง 1 บาท ถ้าถูกได้ 900 บาท
var DefaultRules = map[types.LotteryType]LotteryRule{
	types.LotteryTypeThai: {
		Type: types.LotteryTypeThai,
		Name: "หวยไทย (ใต้ดิน)",
		AllowedBetTypes: []types.BetType{
			types.BetType3Top,
			types.BetType3Tod,
			types.BetType2Top,
			types.BetType2Bottom,
			types.BetTypeRunTop,
			types.BetTypeRunBot,
		},
		DefaultRates: map[types.BetType]float64{
			types.BetType3Top:    900,  // 3 ตัวบน: จ่าย 900 เท่า
			types.BetType3Tod:    150,  // 3 ตัวโต๊ด: จ่าย 150 เท่า
			types.BetType2Top:    90,   // 2 ตัวบน: จ่าย 90 เท่า
			types.BetType2Bottom: 90,   // 2 ตัวล่าง: จ่าย 90 เท่า
			types.BetTypeRunTop:  3.2,  // วิ่งบน: จ่าย 3.2 เท่า
			types.BetTypeRunBot:  4.2,  // วิ่งล่าง: จ่าย 4.2 เท่า
		},
		IsAutoResult: false, // admin กรอกผลเอง
		Description:  "ออกผลวันที่ 1 และ 16 ของทุกเดือน",
	},

	types.LotteryTypeLao: {
		Type: types.LotteryTypeLao,
		Name: "หวยลาว",
		AllowedBetTypes: []types.BetType{
			types.BetType3Top,
			types.BetType3Tod,
			types.BetType2Top,
			types.BetType2Bottom,
			types.BetTypeRunTop,
			types.BetTypeRunBot,
		},
		DefaultRates: map[types.BetType]float64{
			types.BetType3Top:    900,
			types.BetType3Tod:    150,
			types.BetType2Top:    90,
			types.BetType2Bottom: 90,
			types.BetTypeRunTop:  3.2,
			types.BetTypeRunBot:  4.2,
		},
		IsAutoResult: false,
		Description:  "ออกผลตามรอบหวยลาว",
	},

	types.LotteryTypeStockTH: {
		Type: types.LotteryTypeStockTH,
		Name: "หวยหุ้นไทย",
		AllowedBetTypes: []types.BetType{
			types.BetType3Top,
			types.BetType3Tod,
			types.BetType2Top,
			types.BetType2Bottom,
			types.BetTypeRunTop,
			types.BetTypeRunBot,
		},
		DefaultRates: map[types.BetType]float64{
			types.BetType3Top:    850,
			types.BetType3Tod:    120,
			types.BetType2Top:    90,
			types.BetType2Bottom: 90,
			types.BetTypeRunTop:  3.2,
			types.BetTypeRunBot:  4.2,
		},
		IsAutoResult: false, // admin กรอกผล หรือดึงจาก API ตลาดหุ้น
		Description:  "ออกผลตามตลาดหุ้นไทย จันทร์-ศุกร์ (เปิด/ปิดตลาด)",
	},

	types.LotteryTypeStockForeign: {
		Type: types.LotteryTypeStockForeign,
		Name: "หวยหุ้นต่างประเทศ",
		AllowedBetTypes: []types.BetType{
			types.BetType3Top,
			types.BetType3Tod,
			types.BetType2Top,
			types.BetType2Bottom,
			types.BetTypeRunTop,
			types.BetTypeRunBot,
		},
		DefaultRates: map[types.BetType]float64{
			types.BetType3Top:    850,
			types.BetType3Tod:    120,
			types.BetType2Top:    90,
			types.BetType2Bottom: 90,
			types.BetTypeRunTop:  3.2,
			types.BetTypeRunBot:  4.2,
		},
		IsAutoResult: false,
		Description:  "ออกผลตามตลาดหุ้นต่างประเทศ",
	},

	types.LotteryTypeYeekee: {
		Type: types.LotteryTypeYeekee,
		Name: "หวยยี่กี",
		AllowedBetTypes: []types.BetType{
			types.BetType3Top,
			types.BetType3Tod,
			types.BetType2Top,
			types.BetType2Bottom,
			types.BetTypeRunTop,
			types.BetTypeRunBot,
		},
		DefaultRates: map[types.BetType]float64{
			types.BetType3Top:    850,
			types.BetType3Tod:    120,
			types.BetType2Top:    90,
			types.BetType2Bottom: 90,
			types.BetTypeRunTop:  3.2,
			types.BetTypeRunBot:  4.2,
		},
		IsAutoResult: true, // ระบบออกผลเอง (yeekee algorithm)
		Description:  "ออกผลทุก 15 นาที (88 รอบ/วัน) — สมาชิกยิงเลข real-time",
	},
}

// GetRule ดึงกฎของหวยประเภทที่ต้องการ
//
// ตัวอย่าง:
//
//	rule, ok := GetRule(LotteryTypeThai)
//	if ok {
//	    fmt.Println(rule.Name)           // "หวยไทย (ใต้ดิน)"
//	    fmt.Println(rule.DefaultRates)   // map[3TOP:900 3TOD:150 ...]
//	}
func GetRule(lotteryType types.LotteryType) (LotteryRule, bool) {
	rule, ok := DefaultRules[lotteryType]
	return rule, ok
}

// IsBetTypeAllowed ตรวจสอบว่าประเภทหวยนี้รองรับ BetType นี้หรือไม่
//
// ตัวอย่าง:
//
//	IsBetTypeAllowed(LotteryTypeThai, BetType3Top)  → true
//	IsBetTypeAllowed(LotteryTypeThai, BetType3Bottom) → false (หวยไทยไม่มี 3 ตัวล่าง)
func IsBetTypeAllowed(lotteryType types.LotteryType, betType types.BetType) bool {
	rule, ok := GetRule(lotteryType)
	if !ok {
		return false
	}

	for _, allowed := range rule.AllowedBetTypes {
		if allowed == betType {
			return true
		}
	}
	return false
}

// GetDefaultRate ดึง rate จ่ายเริ่มต้นของประเภทหวย + ประเภทการแทง
//
// ตัวอย่าง:
//
//	rate, ok := GetDefaultRate(LotteryTypeThai, BetType3Top)
//	// rate = 900, ok = true
func GetDefaultRate(lotteryType types.LotteryType, betType types.BetType) (float64, bool) {
	rule, ok := GetRule(lotteryType)
	if !ok {
		return 0, false
	}

	rate, ok := rule.DefaultRates[betType]
	return rate, ok
}
