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

	"github.com/farritpcz/lotto-core/betting"
	"github.com/farritpcz/lotto-core/types"
)

// =============================================================================
// Matcher — เทียบเลขที่แทง กับ ผลที่ออก
// =============================================================================

// Match ตรวจสอบว่า bet นี้ถูกรางวัลหรือไม่
//
// เทียบ bet.Number กับ result ตาม BetType:
//   - 3TOP:    bet.Number == result.Top3 (ตรงตำแหน่ง)
//   - 3BOTTOM: bet.Number == result.Bottom3 (ถ้ามี)
//   - 3TOD:    bet.Number เป็น permutation ของ result.Top3 (สลับตำแหน่งได้)
//   - 2TOP:    bet.Number == result.Top2 (2 ตัวท้ายของ 3 ตัวบน)
//   - 2BOTTOM: bet.Number == result.Bottom2
//   - RUN_TOP: bet.Number อยู่ใน result.Top3 (ตัวเดียว ถ้าอยู่ใน 3 ตัวบน ถือว่าถูก)
//   - RUN_BOT: bet.Number อยู่ใน result.Bottom2 (ตัวเดียว ถ้าอยู่ใน 2 ตัวล่าง ถือว่าถูก)
//
// ตัวอย่าง:
//
//	result := RoundResult{Top3: "847", Top2: "47", Bottom2: "56"}
//	Match(Bet{Number: "847", BetType: BetType3Top}, result)   → BetResult{IsWin: true, WinAmount: 84700}
//	Match(Bet{Number: "748", BetType: BetType3Tod}, result)   → BetResult{IsWin: true, ...} (สลับได้)
//	Match(Bet{Number: "123", BetType: BetType3Top}, result)   → BetResult{IsWin: false}
//	Match(Bet{Number: "4", BetType: BetTypeRunTop}, result)   → BetResult{IsWin: true} (4 อยู่ใน 847)
func Match(bet types.Bet, result types.RoundResult) types.BetResult {
	isWin := false

	switch bet.BetType {
	case types.BetType3Top:
		// 3 ตัวบน: ตรงตำแหน่งเป๊ะ
		isWin = bet.Number == result.Top3

	case types.BetType3Tod:
		// 3 ตัวโต๊ด: สลับตำแหน่งได้
		isWin = isPermutation(bet.Number, result.Top3)

	case types.BetType2Top:
		// 2 ตัวบน: ตรงกับ 2 ตัวท้ายของ 3 ตัวบน
		isWin = bet.Number == result.Top2

	case types.BetType2Bottom:
		// 2 ตัวล่าง: ตรงกับ 2 ตัวล่าง
		isWin = bet.Number == result.Bottom2

	case types.BetTypeRunTop:
		// วิ่งบน: เลข 1 ตัว ถ้าอยู่ใน 3 ตัวบน
		isWin = strings.Contains(result.Top3, bet.Number)

	case types.BetTypeRunBot:
		// วิ่งล่าง: เลข 1 ตัว ถ้าอยู่ใน 2 ตัวล่าง
		isWin = strings.Contains(result.Bottom2, bet.Number)

	case types.BetType3Bottom:
		// 3 ตัวล่าง (ถ้ารองรับ): TODO — ขึ้นกับ requirement
		isWin = false
	}

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
	for _, r := range results {
		if r.IsWin {
			totalWinners++
			totalWinAmount += r.WinAmount
		} else {
			totalLosers++
		}
	}
	return
}
