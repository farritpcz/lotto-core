// Package betting — calculator.go
// คำนวณเงินจ่าย (payout) จากจำนวนเงินที่แทง × rate
//
// ความสัมพันธ์:
// - ใช้หลังจาก Validate() ผ่านแล้ว
// - ถูกเรียกตอนลูกค้าแทง: เพื่อแสดงว่า "ถ้าถูกจะได้เท่าไหร่"
// - ถูกเรียกตอนออกผล: เพื่อคำนวณเงินรางวัลจริง (ใน payout/)
package betting

// CalculatePayout คำนวณเงินที่จะได้รับถ้าถูกรางวัล
//
// สูตร: payout = amount × rate
//
// ตัวอย่าง:
//
//	CalculatePayout(100, 900) → 90000  (แทง 100 บาท, rate 900, ได้ 90,000 บาท)
//	CalculatePayout(50, 90)   → 4500   (แทง 50 บาท, rate 90, ได้ 4,500 บาท)
//
// NOTE: rate จ่ายปกติ:
//   - 3 ตัวบน: 800-900
//   - 3 ตัวโต๊ด: 100-150
//   - 2 ตัวบน/ล่าง: 90-95
//   - วิ่ง: 3-4
func CalculatePayout(amount float64, rate float64) float64 {
	if amount <= 0 || rate <= 0 {
		return 0
	}
	return amount * rate
}

// CalculateNetProfit คำนวณกำไรสุทธิของลูกค้า (ไม่รวมเงินต้น)
//
// สูตร: profit = payout - amount = (amount × rate) - amount
//
// ตัวอย่าง:
//
//	CalculateNetProfit(100, 900) → 89900  (ได้ 90,000 - ต้นทุน 100 = กำไร 89,900)
func CalculateNetProfit(amount float64, rate float64) float64 {
	payout := CalculatePayout(amount, rate)
	if payout <= 0 {
		return 0
	}
	return payout - amount
}
