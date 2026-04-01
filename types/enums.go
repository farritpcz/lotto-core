// Package types กำหนด shared types, enums, และ constants กลาง
// ที่ใช้ร่วมกันทุก package ใน lotto-core
//
// ความสัมพันธ์:
// - ใช้โดย: betting/, payout/, yeekee/, numberban/, lottery/
// - import โดย: standalone-member-api (#3), standalone-admin-api (#5),
//               provider-game-api (#7), provider-backoffice-api (#9)
package types

// =============================================================================
// Lottery Types — ประเภทหวยที่รองรับ
// =============================================================================

// LotteryType กำหนดประเภทของหวย
// ใช้ใน: lottery_types table, betting validation, result matching
type LotteryType string

const (
	LotteryTypeThai          LotteryType = "THAI"           // หวยไทย (ใต้ดิน) — ออกผล 1, 16 ของเดือน
	LotteryTypeLao           LotteryType = "LAO"            // หวยลาว — ออกผลตามรอบลาว
	LotteryTypeHanoi         LotteryType = "HANOI"          // หวยฮานอย — ออกผลทุกวัน 18:30
	LotteryTypeMalay         LotteryType = "MALAY"          // หวยมาเลย์ — ออกผลตามรอบมาเลเซีย
	LotteryTypeLao9          LotteryType = "LAO_STAR"       // หวยลาว Star (9+) — ออกผลหลายรอบ/วัน
	LotteryTypeBAAC          LotteryType = "BAAC"           // หวย ธกส. — ออกผลตามรอบ ธกส.
	LotteryTypeGSB           LotteryType = "GSB"            // หวยออมสิน — ออกผลตามรอบออมสิน
	LotteryTypeStockTH       LotteryType = "STOCK_TH"       // หวยหุ้นไทย — ออกผลตามตลาดหุ้น จ-ศ
	LotteryTypeStockForeign  LotteryType = "STOCK_FOREIGN"  // หวยหุ้นต่างประเทศ — ออกผลตามตลาดแต่ละประเทศ
	LotteryTypeYeekee        LotteryType = "YEEKEE"         // หวยยี่กี — ออกผลทุก 15 นาที (88 รอบ/วัน)
	LotteryTypeYeekee5       LotteryType = "YEEKEE_5"       // ยี่กี 5 นาที — ออกผลทุก 5 นาที
	LotteryTypeYeekee15      LotteryType = "YEEKEE_15"      // ยี่กี 15 นาที (มาตรฐาน)
	LotteryTypeYeekeeVIP     LotteryType = "YEEKEE_VIP"     // ยี่กี VIP — rate สูงกว่าปกติ
	LotteryTypeCustom        LotteryType = "CUSTOM"         // หวยอื่นๆ — configurable
)

// IsValid ตรวจสอบว่า LotteryType ถูกต้องหรือไม่
func (lt LotteryType) IsValid() bool {
	switch lt {
	case LotteryTypeThai, LotteryTypeLao, LotteryTypeHanoi, LotteryTypeMalay,
		LotteryTypeLao9, LotteryTypeBAAC, LotteryTypeGSB,
		LotteryTypeStockTH, LotteryTypeStockForeign,
		LotteryTypeYeekee, LotteryTypeYeekee5, LotteryTypeYeekee15, LotteryTypeYeekeeVIP,
		LotteryTypeCustom:
		return true
	}
	return false
}

// IsAutoResult ตรวจสอบว่าประเภทหวยนี้ออกผลอัตโนมัติหรือไม่
// true = ระบบออกผลเอง (Yeekee), false = admin กรอกผล (Thai, Lao, Stock)
func (lt LotteryType) IsAutoResult() bool {
	switch lt {
	case LotteryTypeYeekee, LotteryTypeYeekee5, LotteryTypeYeekee15, LotteryTypeYeekeeVIP:
		return true
	}
	return false
}

// =============================================================================
// Bet Types — ประเภทการแทง
// =============================================================================

// BetType กำหนดประเภทการแทงหวย
// ใช้ใน: bet_types table, betting validation, payout calculation
type BetType string

const (
	BetType3Top    BetType = "3TOP"    // 3 ตัวบน — ตรงตำแหน่ง เช่น 847
	BetType3Bottom BetType = "3BOTTOM" // 3 ตัวล่าง — ตรงตำแหน่ง (บางระบบไม่มี)
	BetType3Tod    BetType = "3TOD"    // 3 ตัวโต๊ด — สลับตำแหน่งได้ เช่น 847 = 478 = 748 ...
	BetType3Front  BetType = "3FRONT"  // 3 ตัวหน้า — 3 ตัวแรกของเลขท้าย 6 ตำแหน่ง
	BetType2Top    BetType = "2TOP"    // 2 ตัวบน — 2 ตัวท้ายของ 3 ตัวบน เช่น 47
	BetType2Bottom BetType = "2BOTTOM" // 2 ตัวล่าง — 2 ตัวล่าง เช่น 56
	BetType4Top    BetType = "4TOP"    // 4 ตัวบน — ตรงตำแหน่ง เช่น 8471
	BetType4Tod    BetType = "4TOD"    // 4 ตัวโต๊ด — สลับตำแหน่งได้ เช่น 8471 = 1748 ...
	BetTypeRunTop  BetType = "RUN_TOP" // วิ่งบน — เลขตัวเดียว ถ้าอยู่ใน 3 ตัวบน ถือว่าถูก
	BetTypeRunBot  BetType = "RUN_BOT" // วิ่งล่าง — เลขตัวเดียว ถ้าอยู่ใน 2 ตัวล่าง ถือว่าถูก
)

// IsValid ตรวจสอบว่า BetType ถูกต้องหรือไม่
func (bt BetType) IsValid() bool {
	switch bt {
	case BetType3Top, BetType3Bottom, BetType3Tod, BetType3Front,
		BetType2Top, BetType2Bottom, BetType4Top, BetType4Tod,
		BetTypeRunTop, BetTypeRunBot:
		return true
	}
	return false
}

// DigitCount จำนวนหลักของเลขที่ต้องกรอกสำหรับ BetType นี้
// เช่น 3TOP → 3 หลัก, 2TOP → 2 หลัก, RUN_TOP → 1 หลัก
func (bt BetType) DigitCount() int {
	switch bt {
	case BetType4Top, BetType4Tod:
		return 4
	case BetType3Top, BetType3Bottom, BetType3Tod, BetType3Front:
		return 3
	case BetType2Top, BetType2Bottom:
		return 2
	case BetTypeRunTop, BetTypeRunBot:
		return 1
	}
	return 0
}

// =============================================================================
// Round Status — สถานะของรอบหวย
// =============================================================================

// RoundStatus สถานะของรอบหวย
// flow: upcoming → open → closed → resulted
type RoundStatus string

const (
	RoundStatusUpcoming RoundStatus = "upcoming" // รอเปิดรับ — ยังไม่ถึงเวลา
	RoundStatusOpen     RoundStatus = "open"     // เปิดรับแทง — สมาชิกแทงได้
	RoundStatusClosed   RoundStatus = "closed"   // ปิดรับแทง — รอผล
	RoundStatusResulted RoundStatus = "resulted" // ออกผลแล้ว — คำนวณแพ้ชนะเสร็จ
)

// IsValid ตรวจสอบว่า RoundStatus ถูกต้องหรือไม่
func (rs RoundStatus) IsValid() bool {
	switch rs {
	case RoundStatusUpcoming, RoundStatusOpen, RoundStatusClosed, RoundStatusResulted:
		return true
	}
	return false
}

// CanBet ตรวจสอบว่ารอบนี้ยังแทงได้หรือไม่
// แทงได้เฉพาะตอน status = open
func (rs RoundStatus) CanBet() bool {
	return rs == RoundStatusOpen
}

// =============================================================================
// Bet Status — สถานะของการเดิมพัน
// =============================================================================

// BetStatus สถานะของ bet
// flow: pending → (won | lost | cancelled | refunded)
type BetStatus string

const (
	BetStatusPending   BetStatus = "pending"   // รอผล — ยังไม่ออกผล
	BetStatusWon       BetStatus = "won"       // ชนะ — ถูกรางวัล จ่ายเงินแล้ว
	BetStatusLost      BetStatus = "lost"      // แพ้ — ไม่ถูกรางวัล
	BetStatusCancelled BetStatus = "cancelled" // ยกเลิก — ยกเลิกก่อนปิดรับ
	BetStatusRefunded  BetStatus = "refunded"  // คืนเงิน — คืนเงินจากเหตุพิเศษ
)

// IsSettled ตรวจสอบว่า bet นี้ตัดสินผลแล้วหรือยัง
// true = จบแล้ว (won/lost/cancelled/refunded), false = ยังรอผล
func (bs BetStatus) IsSettled() bool {
	return bs != BetStatusPending
}

// =============================================================================
// Yeekee Round Status — สถานะรอบยี่กี
// =============================================================================

// YeekeeStatus สถานะของรอบยี่กี
// flow: waiting → shooting → calculating → resulted
type YeekeeStatus string

const (
	YeekeeStatusWaiting     YeekeeStatus = "waiting"     // รอเริ่ม — ยังไม่ถึงเวลา
	YeekeeStatusShooting    YeekeeStatus = "shooting"    // กำลังยิงเลข — สมาชิกส่งเลขได้ (real-time)
	YeekeeStatusCalculating YeekeeStatus = "calculating" // กำลังคำนวณ — หมดเวลายิง กำลังคำนวณผล
	YeekeeStatusResulted    YeekeeStatus = "resulted"    // ออกผลแล้ว — ประกาศผลเสร็จ
)

// CanShoot ตรวจสอบว่ารอบนี้ยังยิงเลขได้หรือไม่
func (ys YeekeeStatus) CanShoot() bool {
	return ys == YeekeeStatusShooting
}

// =============================================================================
// Number Ban Type — ประเภทการอั้นเลข
// =============================================================================

// BanType ประเภทของการอั้นเลข
type BanType string

const (
	BanTypeFull       BanType = "full_ban"    // อั้นเต็ม — ไม่รับแทงเลขนี้เลย
	BanTypeReduceRate BanType = "reduce_rate" // ลด rate — รับแทงแต่จ่ายน้อยลง
)

// =============================================================================
// Transaction Type — ประเภทธุรกรรม
// =============================================================================

// TransactionType ประเภทของธุรกรรมในระบบ wallet
type TransactionType string

const (
	TxTypeDeposit  TransactionType = "deposit"  // ฝากเงิน
	TxTypeWithdraw TransactionType = "withdraw" // ถอนเงิน
	TxTypeBet      TransactionType = "bet"      // วางเดิมพัน (หักเงิน)
	TxTypeWin      TransactionType = "win"      // ชนะรางวัล (เพิ่มเงิน)
	TxTypeRefund   TransactionType = "refund"   // คืนเงิน
)

// =============================================================================
// Wallet Type — ประเภท wallet (provider mode)
// =============================================================================

// WalletType ประเภทของ wallet integration สำหรับ provider mode
// ใช้ใน: operators table, provider-game-api (#7)
type WalletType string

const (
	WalletTypeSeamless WalletType = "seamless" // Seamless — เรียก API operator ทุกครั้ง (balance/debit/credit)
	WalletTypeTransfer WalletType = "transfer" // Transfer — โอนเงินเข้า provider ก่อน แล้วเล่นจาก balance ใน provider
)
