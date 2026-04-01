// Package betting จัดการ logic การตรวจสอบและคำนวณการเดิมพัน
//
// ความสัมพันธ์:
// - ใช้ types/ สำหรับ struct และ enum
// - ถูกเรียกโดย: standalone-member-api (#3), provider-game-api (#7)
// - flow: HTTP request → API handler → betting.Validate() → ถ้าผ่าน → บันทึก DB
//
// ทั้ง standalone (#3) และ provider (#7) เรียก function เหล่านี้เหมือนกันเป๊ะ
// ต่างกันแค่ "ก่อน" และ "หลัง" การเรียก (auth, wallet, callback)
package betting

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/farritpcz/lotto-core/types"
)

// =============================================================================
// Validator — ตรวจสอบข้อมูลการเดิมพัน
// =============================================================================

// regex สำหรับตรวจสอบว่าเป็นตัวเลขล้วนๆ
var numberRegex = regexp.MustCompile(`^\d+$`)

// ValidateNumber ตรวจสอบเลขที่แทงว่าถูก format หรือไม่
//
// กฎ:
//   - ต้องเป็นตัวเลขล้วนๆ (ไม่มีตัวอักษร, ไม่มีช่องว่าง)
//   - จำนวนหลักต้องตรงกับ BetType (3TOP=3หลัก, 2TOP=2หลัก, RUN=1หลัก)
//   - อนุญาตเลข 0 นำหน้า (เช่น "007", "05")
//
// ตัวอย่าง:
//
//	ValidateNumber("847", BetType3Top)   → nil       (ถูกต้อง)
//	ValidateNumber("47", BetType3Top)    → error     (ต้อง 3 หลัก)
//	ValidateNumber("abc", BetType2Top)   → error     (ไม่ใช่ตัวเลข)
//	ValidateNumber("05", BetType2Bottom) → nil       (ถูกต้อง, 0 นำหน้าได้)
func ValidateNumber(number string, betType types.BetType) error {
	// ตรวจว่า betType ถูกต้อง
	if !betType.IsValid() {
		return types.ErrInvalidBetType
	}

	// ตรวจว่าเป็นตัวเลขล้วน
	if !numberRegex.MatchString(number) {
		return fmt.Errorf("%w: must contain only digits, got %q", types.ErrInvalidNumber, number)
	}

	// ตรวจจำนวนหลัก
	expectedDigits := betType.DigitCount()
	if len(number) != expectedDigits {
		return fmt.Errorf("%w: %s requires %d digits, got %d", types.ErrInvalidNumber, betType, expectedDigits, len(number))
	}

	return nil
}

// ValidateAmount ตรวจสอบจำนวนเงินที่แทง
//
// กฎ:
//   - ต้องมากกว่า 0
//   - ต้องไม่น้อยกว่า minBet (ค่าต่ำสุดที่ตั้งไว้)
//   - ต้องไม่มากกว่า maxBet (ค่าสูงสุดที่ตั้งไว้, 0 = ไม่จำกัด)
//
// ตัวอย่าง:
//
//	ValidateAmount(100, 1, 1000)  → nil      (ผ่าน)
//	ValidateAmount(0, 1, 1000)    → error    (ต้อง > 0)
//	ValidateAmount(5000, 1, 1000) → error    (เกิน max)
func ValidateAmount(amount float64, minBet float64, maxBet float64) error {
	if amount <= 0 {
		return fmt.Errorf("%w: amount must be positive, got %.2f", types.ErrInvalidAmount, amount)
	}

	if amount < minBet {
		return fmt.Errorf("%w: minimum is %.2f, got %.2f", types.ErrAmountTooLow, minBet, amount)
	}

	// maxBet = 0 หมายถึงไม่จำกัด
	if maxBet > 0 && amount > maxBet {
		return fmt.Errorf("%w: maximum is %.2f, got %.2f", types.ErrAmountTooHigh, maxBet, amount)
	}

	return nil
}

// Validate ตรวจสอบ BetRequest ทั้งหมด (number + amount + round status)
//
// นี่คือ function หลักที่ standalone-member-api (#3) และ provider-game-api (#7)
// เรียกใช้เมื่อลูกค้าแทงหวย
//
// ขั้นตอน:
//  1. ตรวจ BetType ว่า valid
//  2. ตรวจ LotteryType ว่า valid
//  3. ตรวจ Number ว่าถูก format + จำนวนหลัก
//  4. ตรวจ Amount ว่าอยู่ในช่วงที่กำหนด
//
// NOTE: ไม่ได้เช็คเลขอั้น (numberban) และ limit ต่อเลข ในนี้
// เพราะต้องดึงข้อมูลจาก DB ซึ่งเป็นหน้าที่ของ API layer
// ใช้ numberban.Check() และ CheckBetLimit() แยกต่างหาก
func Validate(req types.BetRequest, minBet float64, maxBet float64) error {
	// ตรวจ BetType
	if !req.BetType.IsValid() {
		return types.ErrInvalidBetType
	}

	// ตรวจ LotteryType
	if !req.LotteryType.IsValid() {
		return fmt.Errorf("invalid lottery type: %s", req.LotteryType)
	}

	// ตรวจ Number
	if err := ValidateNumber(req.Number, req.BetType); err != nil {
		return err
	}

	// ตรวจ Amount
	if err := ValidateAmount(req.Amount, minBet, maxBet); err != nil {
		return err
	}

	return nil
}

// ValidateYeekeeShoot ตรวจสอบเลขที่ยิงในยี่กี
//
// กฎยี่กี:
//   - เลขต้องเป็นตัวเลข 5 หลัก (00000-99999)
//   - ใช้สำหรับ WebSocket ยิงเลข real-time
//
// ตัวอย่าง:
//
//	ValidateYeekeeShoot("12345") → nil     (ถูกต้อง)
//	ValidateYeekeeShoot("1234")  → error   (ต้อง 5 หลัก)
//	ValidateYeekeeShoot("abcde") → error   (ไม่ใช่ตัวเลข)
func ValidateYeekeeShoot(number string) error {
	if !numberRegex.MatchString(number) {
		return fmt.Errorf("%w: must contain only digits", types.ErrInvalidShootNumber)
	}

	if len(number) != 5 {
		return fmt.Errorf("%w: must be exactly 5 digits, got %d", types.ErrInvalidShootNumber, len(number))
	}

	// ตรวจว่าเลขอยู่ในช่วง 00000-99999 (เป็นตัวเลขได้จริง)
	if _, err := strconv.Atoi(number); err != nil {
		return fmt.Errorf("%w: %s", types.ErrInvalidShootNumber, err.Error())
	}

	return nil
}
