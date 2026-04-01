// Package lottery — schedule.go
// กำหนด schedule การออกผลของหวยแต่ละประเภท
//
// ความสัมพันธ์:
// - ถูกเรียกโดย cron job ใน API layer เพื่อสร้างรอบอัตโนมัติ
// - standalone-member-api (#3): สร้างรอบยี่กีทุก 15 นาที
// - standalone-admin-api (#5): สร้างรอบหวยไทย/ลาว/หุ้น
// - provider-game-api (#7) + backoffice-api (#9): เหมือนกัน
package lottery

import (
	"fmt"
	"time"

	"github.com/farritpcz/lotto-core/types"
)

// =============================================================================
// Yeekee Schedule — สร้างรอบยี่กีอัตโนมัติ
// =============================================================================

// YeekeeConfig ตั้งค่ายี่กี
type YeekeeConfig struct {
	IntervalMinutes int // ระยะเวลาแต่ละรอบ (default: 15 นาที)
	RoundsPerDay    int // จำนวนรอบต่อวัน (default: 88)
	ShootDigits     int // จำนวนหลักที่ยิง (default: 5)
}

// DefaultYeekeeConfig ค่า default ยี่กี
var DefaultYeekeeConfig = YeekeeConfig{
	IntervalMinutes: 15,
	RoundsPerDay:    88,
	ShootDigits:     5,
}

// YeekeeRoundSchedule ข้อมูลรอบยี่กีที่ต้องสร้าง
type YeekeeRoundSchedule struct {
	RoundNo   int       // ลำดับรอบ (1-88)
	StartTime time.Time // เวลาเริ่มรับยิง
	EndTime   time.Time // เวลาหยุดรับยิง
}

// GenerateYeekeeSchedule สร้าง schedule รอบยี่กีทั้งวัน
//
// ยี่กีเริ่มรอบแรก 06:00 ทุกวัน (configurable)
// แต่ละรอบ 15 นาที → 88 รอบ/วัน = 06:00 - 28:00 (04:00 วันถัดไป)
//
// ตัวอย่าง:
//
//	schedule := GenerateYeekeeSchedule(time.Now(), DefaultYeekeeConfig)
//	// schedule[0] = {RoundNo: 1, StartTime: 06:00, EndTime: 06:15}
//	// schedule[1] = {RoundNo: 2, StartTime: 06:15, EndTime: 06:30}
//	// ...
//	// schedule[87] = {RoundNo: 88, StartTime: 27:45, EndTime: 28:00}
func GenerateYeekeeSchedule(date time.Time, config YeekeeConfig) []YeekeeRoundSchedule {
	// เริ่มรอบแรก 06:00
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 6, 0, 0, 0, date.Location())

	rounds := make([]YeekeeRoundSchedule, 0, config.RoundsPerDay)
	for i := 0; i < config.RoundsPerDay; i++ {
		roundStart := startOfDay.Add(time.Duration(i*config.IntervalMinutes) * time.Minute)
		roundEnd := roundStart.Add(time.Duration(config.IntervalMinutes) * time.Minute)

		rounds = append(rounds, YeekeeRoundSchedule{
			RoundNo:   i + 1,
			StartTime: roundStart,
			EndTime:   roundEnd,
		})
	}

	return rounds
}

// GetCurrentYeekeeRound หารอบยี่กีปัจจุบัน (ที่กำลัง shooting)
//
// ใช้ใน: cron job ที่ตรวจสอบว่าต้องเปิด/ปิดรอบไหน
func GetCurrentYeekeeRound(now time.Time, config YeekeeConfig) *YeekeeRoundSchedule {
	schedule := GenerateYeekeeSchedule(now, config)
	for _, round := range schedule {
		if now.After(round.StartTime) && now.Before(round.EndTime) {
			return &round
		}
	}
	return nil
}

// =============================================================================
// Round Number Generation — สร้างหมายเลขรอบ
// =============================================================================

// GenerateRoundNumber สร้างหมายเลขรอบสำหรับหวยแต่ละประเภท
//
// Format:
//   - หวยไทย/ลาว: "YYYYMMDD" เช่น "20260401"
//   - หวยหุ้น: "YYYYMMDD-AM" หรือ "YYYYMMDD-PM"
//   - ยี่กี: "YYYYMMDD-RR" เช่น "20260401-01", "20260401-88"
//
// ตัวอย่าง:
//
//	GenerateRoundNumber(LotteryTypeThai, date, 0)     → "20260401"
//	GenerateRoundNumber(LotteryTypeStockTH, date, 1)  → "20260401-AM"
//	GenerateRoundNumber(LotteryTypeYeekee, date, 5)   → "20260401-05"
func GenerateRoundNumber(lotteryType types.LotteryType, date time.Time, roundNo int) string {
	dateStr := date.Format("20060102")

	switch lotteryType {
	case types.LotteryTypeThai, types.LotteryTypeLao:
		return dateStr

	case types.LotteryTypeStockTH, types.LotteryTypeStockForeign:
		if roundNo <= 1 {
			return dateStr + "-AM"
		}
		return dateStr + "-PM"

	case types.LotteryTypeYeekee:
		return fmt.Sprintf("%s-%02d", dateStr, roundNo)

	default:
		return fmt.Sprintf("%s-%02d", dateStr, roundNo)
	}
}

// =============================================================================
// Schedule Helpers
// =============================================================================

// GetThaiLotteryDates คืนวันที่ออกผลหวยไทยในเดือนนั้น (1 และ 16)
//
// ตัวอย่าง:
//
//	dates := GetThaiLotteryDates(2026, 4)
//	// dates = [2026-04-01, 2026-04-16]
func GetThaiLotteryDates(year int, month time.Month) []time.Time {
	loc := time.Now().Location()
	return []time.Time{
		time.Date(year, month, 1, 0, 0, 0, 0, loc),
		time.Date(year, month, 16, 0, 0, 0, 0, loc),
	}
}

// GetStockMarketTimes คืนเวลาเปิด/ปิดตลาดหุ้น
//
// ตลาดหุ้นไทย:
//   - เช้า: 10:00 - 12:30 (ออกผลจาก SET index ตอน 12:30)
//   - บ่าย: 14:30 - 16:30 (ออกผลจาก SET index ตอน 16:30)
type StockMarketTime struct {
	Session   string    // "AM" or "PM"
	OpenTime  time.Time // เวลาเปิดรับแทง
	CloseTime time.Time // เวลาปิดรับแทง (ผลออกหลังนี้)
}

// GetThaiStockTimes คืนเวลาเปิด/ปิดหวยหุ้นไทยสำหรับวันที่กำหนด
func GetThaiStockTimes(date time.Time) []StockMarketTime {
	loc := date.Location()
	y, m, d := date.Year(), date.Month(), date.Day()

	return []StockMarketTime{
		{
			Session:   "AM",
			OpenTime:  time.Date(y, m, d, 9, 0, 0, 0, loc),  // เปิดรับ 09:00
			CloseTime: time.Date(y, m, d, 12, 0, 0, 0, loc),  // ปิดรับ 12:00 (ผลออก ~12:30)
		},
		{
			Session:   "PM",
			OpenTime:  time.Date(y, m, d, 13, 0, 0, 0, loc),  // เปิดรับ 13:00
			CloseTime: time.Date(y, m, d, 16, 0, 0, 0, loc),  // ปิดรับ 16:00 (ผลออก ~16:30)
		},
	}
}

// IsWeekday ตรวจว่าวันนี้เป็นวันทำการหรือไม่ (จ-ศ)
// ใช้สำหรับหวยหุ้น — ไม่ออกผลวัน เสาร์-อาทิตย์
func IsWeekday(date time.Time) bool {
	day := date.Weekday()
	return day != time.Saturday && day != time.Sunday
}
