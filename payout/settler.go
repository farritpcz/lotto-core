// Package payout — settler.go
// Logic จ่ายเงินรางวัล — ประมวลผลทั้งรอบ
//
// ความสัมพันธ์:
// - ใช้ Match() / MatchAll() จาก matcher.go
// - ถูกเรียกโดย:
//   - standalone-admin-api (#5): admin กรอกผล → SettleRound()
//   - provider-backoffice-api (#9): admin กรอกผล → SettleRound()
//   - standalone-member-api (#3): ยี่กีออกผลอัตโนมัติ → SettleRound()
//   - provider-game-api (#7): ยี่กีออกผลอัตโนมัติ → SettleRound()
//
// NOTE: function เหล่านี้ไม่จัดการ DB โดยตรง — คืนผลลัพธ์กลับไปให้ API layer
// จัดการ DB transaction + wallet (internal หรือ operator API)
package payout

import (
	"github.com/farritpcz/lotto-core/types"
)

// SettleRoundInput ข้อมูลที่ต้องส่งมาเพื่อตัดสินผลรอบ
type SettleRoundInput struct {
	RoundID int64              // ID ของรอบ
	Result  types.RoundResult  // ผลรางวัล (Top3, Top2, Bottom2)
	Bets    []types.Bet        // bets ทั้งหมดของรอบ (status = pending เท่านั้น)
}

// SettleRoundOutput ผลลัพธ์หลังตัดสินผลรอบ
type SettleRoundOutput struct {
	RoundID        int64              // ID ของรอบ
	BetResults     []types.BetResult  // ผลแต่ละ bet (won/lost + winAmount)
	TotalWinners   int                // จำนวนคนชนะ
	TotalWinAmount float64            // เงินรางวัลรวม
	TotalLosers    int                // จำนวนคนแพ้
	TotalBetAmount float64            // ยอดแทงรวม
	Profit         float64            // กำไร/ขาดทุน = ยอดแทงรวม - เงินรางวัลรวม
}

// SettleRound ตัดสินผลรอบหวย — เทียบ bets ทั้งหมดกับผลรางวัล
//
// Flow:
//  1. เทียบ bets ทั้งหมดกับ result → ได้ผล won/lost ทุก bet
//  2. สรุปผลรวม (winners, losers, amounts)
//  3. คำนวณกำไร/ขาดทุน
//  4. return SettleRoundOutput ให้ API layer ไปอัพเดท DB + จ่ายเงิน
//
// ตัวอย่าง:
//
//	input := SettleRoundInput{
//	    RoundID: 123,
//	    Result:  RoundResult{Top3: "847", Top2: "47", Bottom2: "56"},
//	    Bets:    []Bet{...ทุก bet ในรอบ...},
//	}
//	output := SettleRound(input)
//	// output.TotalWinners = 5
//	// output.TotalWinAmount = 45000
//	// output.Profit = 100000 - 45000 = 55000
//
// API layer (#5/#9) ต้องทำหลังจากได้ output:
//  1. อัพเดท bets table: status → won/lost, win_amount
//  2. จ่ายเงินคนชนะ:
//     - standalone: UPDATE members SET balance = balance + win_amount
//     - provider: POST {operator}/wallet/credit
//  3. สร้าง transactions สำหรับคนชนะ
//  4. อัพเดท lottery_rounds: status → resulted
func SettleRound(input SettleRoundInput) SettleRoundOutput {
	// 1. เทียบ bets ทั้งหมด
	betResults := MatchAll(input.Bets, input.Result)

	// 2. สรุปผล
	totalWinners, totalWinAmount, totalLosers := SummarizeResults(betResults)

	// 3. คำนวณยอดแทงรวม
	var totalBetAmount float64
	for _, bet := range input.Bets {
		if !bet.Status.IsSettled() { // นับเฉพาะ pending
			totalBetAmount += bet.Amount
		}
	}

	// 4. กำไร/ขาดทุน
	profit := totalBetAmount - totalWinAmount

	return SettleRoundOutput{
		RoundID:        input.RoundID,
		BetResults:     betResults,
		TotalWinners:   totalWinners,
		TotalWinAmount: totalWinAmount,
		TotalLosers:    totalLosers,
		TotalBetAmount: totalBetAmount,
		Profit:         profit,
	}
}

// GroupWinnersByMember จัดกลุ่มผลชนะตาม member
//
// ใช้สำหรับจ่ายเงิน — รวมรางวัลทั้งหมดของ member คนเดียวกัน
// แทนที่จะจ่ายทีละ bet → จ่ายรวมครั้งเดียวต่อ member (ลด DB writes)
//
// Returns: map[memberID] → totalWinAmount
//
// ตัวอย่าง:
//
//	memberPayouts := GroupWinnersByMember(bets, betResults)
//	// memberPayouts[101] = 90000  (member 101 ชนะรวม 90,000)
//	// memberPayouts[205] = 4500   (member 205 ชนะรวม 4,500)
func GroupWinnersByMember(bets []types.Bet, betResults []types.BetResult) map[int64]float64 {
	memberPayouts := make(map[int64]float64)

	// สร้าง map betID → bet เพื่อ lookup memberID
	betMap := make(map[int64]types.Bet)
	for _, bet := range bets {
		betMap[bet.ID] = bet
	}

	for _, result := range betResults {
		if result.IsWin {
			bet, ok := betMap[result.BetID]
			if ok {
				memberPayouts[bet.MemberID] += result.WinAmount
			}
		}
	}

	return memberPayouts
}

// GroupWinnersByOperator จัดกลุ่มผลชนะตาม operator (provider mode)
//
// ใช้สำหรับ: provider-backoffice-api (#9) → callback แจ้ง operator ว่าต้องจ่ายเท่าไหร่
// standalone ไม่ใช้ function นี้
//
// Returns: map[operatorID] → map[memberID] → totalWinAmount
func GroupWinnersByOperator(bets []types.Bet, betResults []types.BetResult) map[int64]map[int64]float64 {
	operatorPayouts := make(map[int64]map[int64]float64)

	betMap := make(map[int64]types.Bet)
	for _, bet := range bets {
		betMap[bet.ID] = bet
	}

	for _, result := range betResults {
		if result.IsWin {
			bet, ok := betMap[result.BetID]
			if !ok || bet.OperatorID == nil {
				continue // standalone bet — ข้าม
			}

			opID := *bet.OperatorID
			if operatorPayouts[opID] == nil {
				operatorPayouts[opID] = make(map[int64]float64)
			}
			operatorPayouts[opID][bet.MemberID] += result.WinAmount
		}
	}

	return operatorPayouts
}
