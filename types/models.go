package types

import "time"

// =============================================================================
// Models กลาง — struct ที่ใช้ร่วมกันทุก package ใน lotto-core
//
// NOTE: นี่ไม่ใช่ DB models โดยตรง — แต่ละ repo (standalone/provider)
// จะมี DB models ของตัวเอง แล้ว map มาเป็น struct เหล่านี้
// เพื่อส่งเข้า lotto-core ให้คำนวณ
//
// ความสัมพันธ์:
// - ใช้โดย: betting/, payout/, yeekee/, numberban/
// - standalone-member-api (#3) → map DB row → Bet struct → ส่งเข้า payout.Match()
// - provider-game-api (#7) → map DB row → Bet struct → ส่งเข้า payout.Match()
// =============================================================================

// Bet แทนการเดิมพัน 1 รายการ
// ใช้ใน: betting.Validate(), payout.Match(), payout.Calculate()
type Bet struct {
	ID            int64      // ID ของ bet (จาก DB)
	MemberID      int64      // ID สมาชิกที่แทง
	OperatorID    *int64     // ID operator (nil = standalone mode, มีค่า = provider mode)
	LotteryTypeID int64      // ID ประเภทหวย
	LotteryType   LotteryType // ประเภทหวย (THAI, LAO, YEEKEE, etc.)
	RoundID       int64      // ID รอบหวย
	BetType       BetType    // ประเภทการแทง (3TOP, 2BOTTOM, etc.)
	Number        string     // เลขที่แทง เช่น "847", "56", "5"
	Amount        float64    // จำนวนเงินที่แทง
	Rate          float64    // rate จ่าย ณ ตอนแทง (เก็บไว้เพราะ rate อาจเปลี่ยนทีหลัง)
	Status        BetStatus  // สถานะ (pending, won, lost, etc.)
	WinAmount     float64    // เงินรางวัลที่ได้ (0 ถ้าแพ้)
	CreatedAt     time.Time  // เวลาที่แทง
}

// LotteryRound แทนรอบหวย 1 รอบ
// ใช้ใน: betting (เช็คว่ารอบเปิดรับ), payout (ดึงผลมาเทียบ)
type LotteryRound struct {
	ID            int64       // ID ของรอบ
	LotteryTypeID int64       // ID ประเภทหวย
	LotteryType   LotteryType // ประเภทหวย
	RoundNumber   string      // หมายเลขรอบ เช่น "20260401-01"
	RoundDate     time.Time   // วันที่ของรอบ
	OpenTime      time.Time   // เวลาเปิดรับแทง
	CloseTime     time.Time   // เวลาปิดรับแทง
	Status        RoundStatus // สถานะรอบ
	Result        *RoundResult // ผลรางวัล (nil ถ้ายังไม่ออก)
}

// RoundResult ผลรางวัลของรอบ
// ใช้ใน: payout.Match() — เทียบเลขที่แทง กับ ผลที่ออก
type RoundResult struct {
	Top3    string // 3 ตัวบน เช่น "847"
	Top2    string // 2 ตัวบน (2 ตัวท้ายของ Top3) เช่น "47"
	Bottom2 string // 2 ตัวล่าง เช่น "56"
}

// PayRate อัตราจ่ายสำหรับประเภทหวย + ประเภทการแทง
// ใช้ใน: betting.CalculatePayout()
type PayRate struct {
	LotteryTypeID   int64    // ID ประเภทหวย
	BetType         BetType  // ประเภทการแทง
	OperatorID      *int64   // ID operator (nil = rate กลาง, มีค่า = rate ของ operator)
	Rate            float64  // อัตราจ่าย เช่น 900 (แทง 1 บาท ได้ 900 บาท)
	MaxBetPerNumber float64  // จำนวนเงินสูงสุดที่แทงต่อเลข
}

// NumberBan เลขที่ถูกอั้น
// ใช้ใน: numberban.Check() — เช็คก่อนรับ bet
type NumberBan struct {
	LotteryTypeID int64    // ID ประเภทหวย
	RoundID       *int64   // ID รอบ (nil = อั้นทุกรอบ)
	OperatorID    *int64   // ID operator (nil = อั้น global, มีค่า = อั้นเฉพาะ operator)
	BetType       BetType  // ประเภทการแทง
	Number        string   // เลขที่อั้น
	BanType       BanType  // ประเภทการอั้น (full_ban / reduce_rate)
	ReducedRate   float64  // rate ที่ลดลง (ใช้เมื่อ BanType = reduce_rate)
	MaxAmount     float64  // จำนวนเงินสูงสุดที่รับ (0 = ไม่จำกัด)
}

// YeekeeRound แทนรอบยี่กี 1 รอบ
// ใช้ใน: yeekee.CalculateResult()
type YeekeeRound struct {
	ID           int64        // ID ของรอบยี่กี
	RoundID      int64        // ID ของ lottery_round ที่เชื่อมกับ
	RoundNo      int          // ลำดับรอบในวัน (1-88)
	StartTime    time.Time    // เวลาเริ่มรับยิง
	EndTime      time.Time    // เวลาหยุดรับยิง
	Status       YeekeeStatus // สถานะ
	ResultNumber string       // ผลที่ออก (หลังคำนวณ)
}

// YeekeeShoot แทนเลขที่สมาชิกยิงมา 1 ครั้ง
// ใช้ใน: yeekee.CalculateResult() — นำเลขทั้งหมดมาคำนวณผล
type YeekeeShoot struct {
	ID        int64     // ID
	RoundID   int64     // ID รอบยี่กี
	MemberID  int64     // ID สมาชิกที่ยิง
	Number    string    // เลขที่ยิง (5 หลัก เช่น "12345")
	ShotAt    time.Time // เวลาที่ยิง
}

// BetRequest ข้อมูลที่ต้องใช้ในการวาง bet
// ใช้ใน: betting.Validate(), betting.CheckLimit()
// standalone-member-api (#3) และ provider-game-api (#7) จะสร้าง struct นี้จาก HTTP request
type BetRequest struct {
	MemberID      int64       // ID สมาชิก
	OperatorID    *int64      // ID operator (nil สำหรับ standalone)
	LotteryTypeID int64       // ID ประเภทหวย
	LotteryType   LotteryType // ประเภทหวย
	RoundID       int64       // ID รอบ
	BetType       BetType     // ประเภทการแทง
	Number        string      // เลขที่แทง
	Amount        float64     // จำนวนเงิน
}

// BetResult ผลการตรวจ bet หลังออกผลรางวัล
// ใช้ใน: payout.Match() → return BetResult
type BetResult struct {
	BetID     int64     // ID ของ bet
	IsWin     bool      // ถูกรางวัลหรือไม่
	WinAmount float64   // เงินรางวัล (0 ถ้าแพ้)
	Status    BetStatus // สถานะใหม่ (won/lost)
}
