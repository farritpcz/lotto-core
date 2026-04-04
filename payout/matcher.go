// Package payout จัดการ logic การเทียบผลรางวัล และคำนวณเงินรางวัล
//
// ความสัมพันธ์:
// - ใช้ types.Bet, types.RoundResult, types.BetResult
// - ใช้ betting.CalculatePayout() สำหรับคำนวณเงินรางวัล
// - ถูกเรียกโดย:
//   - standalone-admin-api (#5): เมื่อ admin กรอกผล → trigger job → payout.SettleRound()
//   - provider-backoffice-api (#9): เมื่อ admin กรอกผล → trigger job → payout.SettleRound()
//   - standalone-member-api (#3): เมื่อยี่กีออกผลอัตโนมัติ
//   - provider-game-api (#7): เมื่อยี่กีออกผลอัตโนมัติ
//
// flow: ผลออก → ดึง bets ทั้งหมดของรอบ → Match() ทีละ bet → จ่ายเงินคนชนะ
package payout

import (
	"sort"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/farritpcz/lotto-core/betting"
	"github.com/farritpcz/lotto-core/types"
)

// =============================================================================
// Matcher — เทียบเลขที่แทง กับ ผลที่ออก
// =============================================================================

// Match ตรวจสอบว่า bet นี้ถูกรางวัลหรือไม่
//
// เทียบ bet.Number กับ result ตาม BetType:
//
//	BetType     | เทียบกับ          | วิธีเทียบ
//	----------- | ----------------- | -----------------------------------------
//	3TOP        | result.Top3       | ตรงตำแหน่งเป๊ะ
//	3TOD        | result.Top3       | สลับตำแหน่งได้ (permutation)
//	3FRONT      | result.Front3     | ตรงตำแหน่งเป๊ะ (3 ตัวหน้ารางวัลที่ 1)
//	3BOTTOM     | result.Bottom3    | ตรงตำแหน่ง กับรางวัลใดรางวัลหนึ่ง (comma-separated)
//	4TOP        | result.Top3       | ตรงตำแหน่ง (ต้องมีผล 4+ หลัก — เทียบ 4 ตัวสุดท้าย)
//	4TOD        | result.Top3       | สลับตำแหน่งได้ (ต้องมีผล 4+ หลัก)
//	2TOP        | result.Top2       | ตรงตำแหน่งเป๊ะ
//	2BOTTOM     | result.Bottom2    | ตรงตำแหน่งเป๊ะ
//	RUN_TOP     | result.Top3       | เลข 1 ตัว ถ้าอยู่ใน 3 ตัวบน
//	RUN_BOT     | result.Bottom2    | เลข 1 ตัว ถ้าอยู่ใน 2 ตัวล่าง
//
// ตัวอย่าง:
//
//	result := RoundResult{Top3: "847", Top2: "47", Bottom2: "56", Front3: "491", Bottom3: "123,456"}
//	Match(Bet{Number: "847", BetType: 3TOP}, result)   → IsWin: true
//	Match(Bet{Number: "748", BetType: 3TOD}, result)   → IsWin: true  (สลับได้)
//	Match(Bet{Number: "491", BetType: 3FRONT}, result) → IsWin: true
//	Match(Bet{Number: "123", BetType: 3BOTTOM}, result)→ IsWin: true  (ตรงรางวัลแรก)
//	Match(Bet{Number: "456", BetType: 3BOTTOM}, result)→ IsWin: true  (ตรงรางวัลที่ 2)
//	Match(Bet{Number: "789", BetType: 3BOTTOM}, result)→ IsWin: false
//	Match(Bet{Number: "4", BetType: RUN_TOP}, result)  → IsWin: true  (4 อยู่ใน 847)
func Match(bet types.Bet, result types.RoundResult) types.BetResult {
	isWin := false

	switch bet.BetType {

	// ─── 3 ตัวบน: ตรงตำแหน่งเป๊ะ กับ 3 ตัวท้ายรางวัลที่ 1 ────────
	case types.BetType3Top:
		isWin = bet.Number == result.Top3

	// ─── 3 ตัวโต๊ด: สลับตำแหน่งได้ กับ 3 ตัวท้ายรางวัลที่ 1 ──────
	case types.BetType3Tod:
		isWin = isPermutation(bet.Number, result.Top3)

	// ─── 3 ตัวหน้า: ตรงตำแหน่งเป๊ะ กับ 3 ตัวแรกของรางวัลที่ 1 ────
	// ใช้ result.Front3 — admin ต้องกรอกมาด้วย (ไม่ใช่ทุกหวยจะมี)
	case types.BetType3Front:
		if result.Front3 != "" {
			isWin = bet.Number == result.Front3
		}

	// ─── 3 ตัวล่าง: ตรงกับผลเลขท้าย 3 ตัว (อาจมีหลายรางวัล) ─────
	// result.Bottom3 อาจเป็น "123" (รางวัลเดียว) หรือ "123,456" (หลายรางวัล)
	// ตรงกับรางวัลใดรางวัลหนึ่งก็ถือว่าถูก
	case types.BetType3Bottom:
		if result.Bottom3 != "" {
			// แยกหลายรางวัลด้วย comma แล้วเทียบทีละตัว
			prizes := strings.Split(result.Bottom3, ",")
			for _, prize := range prizes {
				trimmed := strings.TrimSpace(prize)
				if trimmed != "" && bet.Number == trimmed {
					isWin = true
					break
				}
			}
		}

	// ─── 4 ตัวบน: ตรงตำแหน่งเป๊ะ กับ 4 ตัวสุดท้ายของผลเลข 6 หลัก ─
	// หวยบางประเภทผลเลข 6 หลัก เช่น "491847" → 4 ตัวท้าย = "1847"
	// ถ้า Top3 มี 3 หลัก → ใช้ Front3[2:] + Top3 สร้าง 4 หลักสุดท้าย
	// ถ้าผลมี 4+ หลักอยู่แล้ว → ตัด 4 ตัวท้าย
	case types.BetType4Top:
		full := buildFullResult(result)
		if len(full) >= 4 {
			last4 := full[len(full)-4:]
			isWin = bet.Number == last4
		}

	// ─── 4 ตัวโต๊ด: สลับตำแหน่งได้ กับ 4 ตัวสุดท้าย ──────────────
	case types.BetType4Tod:
		full := buildFullResult(result)
		if len(full) >= 4 {
			last4 := full[len(full)-4:]
			isWin = isPermutation(bet.Number, last4)
		}

	// ─── 2 ตัวบน: ตรงกับ 2 ตัวท้ายของ 3 ตัวบน ────────────────────
	case types.BetType2Top:
		isWin = bet.Number == result.Top2

	// ─── 2 ตัวล่าง: ตรงกับ 2 ตัวล่าง ──────────────────────────────
	case types.BetType2Bottom:
		isWin = bet.Number == result.Bottom2

	// ─── วิ่งบน: เลข 1 ตัว ถ้าอยู่ใน 3 ตัวบน ─────────────────────
	case types.BetTypeRunTop:
		isWin = strings.Contains(result.Top3, bet.Number)

	// ─── วิ่งล่าง: เลข 1 ตัว ถ้าอยู่ใน 2 ตัวล่าง ─────────────────
	case types.BetTypeRunBot:
		isWin = strings.Contains(result.Bottom2, bet.Number)
	}

	// ─── สร้าง BetResult ────────────────────────────────────────────
	if isWin {
		winAmount := betting.CalculatePayout(bet.Amount, bet.Rate)
		return types.BetResult{
			BetID:     bet.ID,
			IsWin:     true,
			WinAmount: winAmount,
			Status:    types.BetStatusWon,
		}
	}

	return types.BetResult{
		BetID:     bet.ID,
		IsWin:     false,
		WinAmount: 0,
		Status:    types.BetStatusLost,
	}
}

// buildFullResult สร้างเลขผลรวม 6 หลัก จาก Front3 + Top3
//
// หวยไทย: รางวัลที่ 1 = 6 หลัก เช่น "491847"
// → Front3 = "491", Top3 = "847"
// → full = "491847", 4 ตัวท้าย = "1847"
//
// ถ้าไม่มี Front3 → ใช้ Top3 ตรงๆ (ผลจะสั้นกว่า 4 หลัก → 4TOP ไม่ match)
func buildFullResult(result types.RoundResult) string {
	if result.Front3 != "" {
		return result.Front3 + result.Top3
	}
	return result.Top3
}

// MatchAll เทียบผล bets ทั้งหมดของรอบ
//
// เรียกครั้งเดียว ได้ผลทุก bet ในรอบ
// ใช้ใน background job ตอนออกผลรางวัล
//
// ตัวอย่าง:
//
//	bets := []Bet{...ทุก bet ในรอบนี้...}
//	results := MatchAll(bets, roundResult)
//	for _, r := range results {
//	    if r.IsWin { /* จ่ายเงิน */ }
//	}
func MatchAll(bets []types.Bet, result types.RoundResult) []types.BetResult {
	results := make([]types.BetResult, 0, len(bets))
	for _, bet := range bets {
		// ข้าม bet ที่ตัดสินผลไปแล้ว (cancelled, refunded)
		if bet.Status.IsSettled() {
			continue
		}
		results = append(results, Match(bet, result))
	}
	return results
}

// =============================================================================
// Helper functions
// =============================================================================

// isPermutation ตรวจสอบว่า a เป็น permutation ของ b หรือไม่
// ใช้สำหรับ 3 ตัวโต๊ด — สลับตำแหน่งแล้วตรงกัน
//
// ตัวอย่าง:
//
//	isPermutation("847", "847") → true  (ตรง)
//	isPermutation("748", "847") → true  (สลับ)
//	isPermutation("478", "847") → true  (สลับ)
//	isPermutation("123", "847") → false (ไม่ตรง)
//	isPermutation("884", "847") → false (ตัวเลขไม่ตรง)
func isPermutation(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	// sort ตัวอักษรแล้วเทียบ
	aChars := strings.Split(a, "")
	bChars := strings.Split(b, "")
	sort.Strings(aChars)
	sort.Strings(bChars)

	return strings.Join(aChars, "") == strings.Join(bChars, "")
}

// SummarizeResults สรุปผลรวมของรอบ
//
// ใช้ใน: admin dashboard แสดงสรุป "รอบนี้ถูกกี่คน จ่ายเท่าไหร่"
//
// Returns:
//   - totalWinners: จำนวนคนชนะ
//   - totalWinAmount: เงินรางวัลรวม
//   - totalLosers: จำนวนคนแพ้
func SummarizeResults(results []types.BetResult) (totalWinners int, totalWinAmount float64, totalLosers int) {
	// ⚠️ ใช้ decimal สำหรับ summation ป้องกัน precision error
	totalDec := decimal.Zero
	for _, r := range results {
		if r.IsWin {
			totalWinners++
			totalDec = totalDec.Add(decimal.NewFromFloat(r.WinAmount))
		} else {
			totalLosers++
		}
	}
	totalWinAmount, _ = totalDec.Float64()
	return
}
