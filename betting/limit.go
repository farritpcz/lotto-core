// Package betting — limit.go
// ตรวจสอบ bet limit ต่อเลข/ต่อรอบ
//
// ความสัมพันธ์:
// - ถูกเรียกหลัง Validate() ผ่าน และหลัง numberban.Check() ผ่าน
// - แต่ละ API (#3, #7) ต้องดึง currentTotal จาก DB หรือ Redis cache มาส่งให้
// - Redis ใช้เก็บ bet totals เพื่อ fast lookup (ไม่ต้อง query DB ทุกครั้ง)
package betting

import (
	"fmt"

	"github.com/farritpcz/lotto-core/types"
)

// CheckBetLimit ตรวจสอบว่าการแทงเลขนี้เกิน limit หรือไม่
//
// Parameters:
//   - number:       เลขที่จะแทง เช่น "847"
//   - amount:       จำนวนเงินที่จะแทง
//   - currentTotal: ยอดรวมที่เลขนี้ถูกแทงไปแล้วในรอบนี้ (ดึงจาก DB/Redis)
//   - maxPerNumber: จำนวนเงินสูงสุดที่รับต่อเลข (จาก pay_rates table, 0 = ไม่จำกัด)
//
// ตัวอย่าง:
//
//	CheckBetLimit("847", 100, 800, 1000)  → nil    (800+100=900, ยังไม่เกิน 1000)
//	CheckBetLimit("847", 300, 800, 1000)  → error  (800+300=1100, เกิน 1000)
//	CheckBetLimit("847", 100, 800, 0)     → nil    (maxPerNumber=0 ไม่จำกัด)
//
// NOTE: currentTotal ควรดึงจาก Redis เพื่อความเร็ว
// key pattern: "bet_total:{roundID}:{betType}:{number}" → value: float64
func CheckBetLimit(number string, amount float64, currentTotal float64, maxPerNumber float64) error {
	// maxPerNumber = 0 หมายถึงไม่จำกัด
	if maxPerNumber <= 0 {
		return nil
	}

	newTotal := currentTotal + amount
	if newTotal > maxPerNumber {
		remaining := maxPerNumber - currentTotal
		if remaining < 0 {
			remaining = 0
		}
		return fmt.Errorf("%w: max %.2f per number, current %.2f, tried %.2f, remaining %.2f",
			types.ErrExceedBetLimit, maxPerNumber, currentTotal, amount, remaining)
	}

	return nil
}

// CalculateRemainingLimit คำนวณจำนวนเงินที่ยังแทงได้อีก สำหรับเลขนี้
//
// ใช้แสดงให้ลูกค้าเห็นว่า "เลขนี้แทงได้อีก X บาท"
//
// ตัวอย่าง:
//
//	CalculateRemainingLimit(800, 1000) → 200   (แทงได้อีก 200)
//	CalculateRemainingLimit(1000, 1000) → 0    (เต็มแล้ว)
//	CalculateRemainingLimit(500, 0) → -1       (ไม่จำกัด, return -1)
func CalculateRemainingLimit(currentTotal float64, maxPerNumber float64) float64 {
	// ไม่จำกัด
	if maxPerNumber <= 0 {
		return -1
	}

	remaining := maxPerNumber - currentTotal
	if remaining < 0 {
		return 0
	}
	return remaining
}
