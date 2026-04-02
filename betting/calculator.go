// Package betting — calculator.go
// คำนวณเงินจ่าย (payout) จากจำนวนเงินที่แทง × rate
//
// ⚠️ SECURITY: ใช้ shopspring/decimal ป้องกัน float64 precision errors
// float64: 0.1 + 0.2 ≠ 0.3 → ยอดเงินผิดพลาดสะสม
// decimal: คำนวณแม่นยำเหมือนเครื่องคิดเลข
//
// ความสัมพันธ์:
// - ใช้หลังจาก Validate() ผ่านแล้ว
// - ถูกเรียกตอนลูกค้าแทง: เพื่อแสดงว่า "ถ้าถูกจะได้เท่าไหร่"
// - ถูกเรียกตอนออกผล: เพื่อคำนวณเงินรางวัลจริง (ใน payout/)
package betting

import "github.com/shopspring/decimal"

// MaxPayoutAmount จำนวนเงินจ่ายสูงสุดต่อ bet (ป้องกัน overflow)
var MaxPayoutAmount = decimal.NewFromFloat(100_000_000) // 100 ล้าน

// CalculatePayoutDecimal คำนวณเงินที่จะได้รับถ้าถูกรางวัล (decimal — แม่นยำ)
//
// สูตร: payout = amount × rate
// ถ้าผลเกิน MaxPayoutAmount → จำกัดไว้ที่ MaxPayoutAmount
func CalculatePayoutDecimal(amount, rate decimal.Decimal) decimal.Decimal {
	if amount.LessThanOrEqual(decimal.Zero) || rate.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero
	}
	result := amount.Mul(rate)
	if result.GreaterThan(MaxPayoutAmount) {
		return MaxPayoutAmount
	}
	return result
}

// CalculateNetProfitDecimal คำนวณกำไรสุทธิ (ไม่รวมเงินต้น)
//
// สูตร: profit = (amount × rate) - amount
func CalculateNetProfitDecimal(amount, rate decimal.Decimal) decimal.Decimal {
	payout := CalculatePayoutDecimal(amount, rate)
	if payout.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero
	}
	return payout.Sub(amount)
}

// CalculatePayout คำนวณเงินที่จะได้รับถ้าถูกรางวัล (float64 wrapper)
//
// ⚠️ ภายในแปลงเป็น decimal คำนวณ แล้วแปลงกลับ float64
// ใช้ได้ทุกจุดที่ยังรับ float64 — ความแม่นยำเท่ากับ decimal
func CalculatePayout(amount float64, rate float64) float64 {
	result := CalculatePayoutDecimal(
		decimal.NewFromFloat(amount),
		decimal.NewFromFloat(rate),
	)
	f, _ := result.Float64()
	return f
}

// CalculateNetProfit คำนวณกำไรสุทธิ (float64 wrapper)
func CalculateNetProfit(amount float64, rate float64) float64 {
	result := CalculateNetProfitDecimal(
		decimal.NewFromFloat(amount),
		decimal.NewFromFloat(rate),
	)
	f, _ := result.Float64()
	return f
}
