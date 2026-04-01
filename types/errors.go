package types

import "errors"

// =============================================================================
// Errors กลาง — error ที่ใช้ร่วมกันทุก package
//
// แต่ละ repo (standalone/provider) จะจับ error เหล่านี้
// แล้วแปลงเป็น HTTP response ที่เหมาะสม
// เช่น ErrInvalidNumber → 400 Bad Request
//      ErrNumberBanned  → 403 Forbidden
// =============================================================================

// Betting errors — ข้อผิดพลาดเกี่ยวกับการแทง
var (
	ErrInvalidNumber    = errors.New("invalid number format")          // เลขไม่ถูก format (เช่น ใส่ตัวอักษร หรือจำนวนหลักไม่ตรง)
	ErrInvalidBetType   = errors.New("invalid bet type")              // ประเภทการแทงไม่ถูกต้อง
	ErrInvalidAmount    = errors.New("invalid bet amount")            // จำนวนเงินไม่ถูกต้อง (ติดลบ หรือ 0)
	ErrAmountTooLow     = errors.New("bet amount below minimum")     // จำนวนเงินต่ำกว่า minimum
	ErrAmountTooHigh    = errors.New("bet amount exceeds maximum")   // จำนวนเงินเกิน maximum
	ErrExceedBetLimit   = errors.New("bet limit exceeded for number") // เกิน limit ต่อเลข (รวมทุกคนแทงเลขนี้เกินแล้ว)
)

// Number ban errors — ข้อผิดพลาดเกี่ยวกับเลขอั้น
var (
	ErrNumberBanned     = errors.New("number is banned")             // เลขถูกอั้นเต็ม (full_ban)
	ErrNumberRateReduced = errors.New("number rate has been reduced") // เลขถูกลด rate (reduce_rate) — ไม่ใช่ error จริงๆ แต่ใช้แจ้งเตือน
)

// Round errors — ข้อผิดพลาดเกี่ยวกับรอบหวย
var (
	ErrRoundNotOpen   = errors.New("round is not open for betting") // รอบยังไม่เปิด หรือปิดแล้ว
	ErrRoundNotFound  = errors.New("round not found")               // ไม่พบรอบ
	ErrRoundResulted  = errors.New("round already has result")      // รอบนี้ออกผลแล้ว
)

// Yeekee errors — ข้อผิดพลาดเกี่ยวกับยี่กี
var (
	ErrYeekeeNotShooting = errors.New("yeekee round is not in shooting phase") // รอบยี่กียังไม่เปิดยิง
	ErrInvalidShootNumber = errors.New("invalid shoot number")                 // เลขยิงไม่ถูก format (ต้อง 5 หลัก)
	ErrNoShoots          = errors.New("no shoots in this round")              // ไม่มีเลขยิงในรอบนี้ (คำนวณผลไม่ได้)
)

// Payout errors — ข้อผิดพลาดเกี่ยวกับการจ่ายเงิน
var (
	ErrNoResult       = errors.New("round has no result yet")        // ยังไม่มีผลรางวัล
	ErrAlreadySettled = errors.New("bet already settled")            // bet นี้ตัดสินผลไปแล้ว
)
