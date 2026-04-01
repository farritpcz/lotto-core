package payout

import (
	"testing"

	"github.com/farritpcz/lotto-core/types"
)

func TestSettleRound(t *testing.T) {
	result := types.RoundResult{Top3: "847", Top2: "47", Bottom2: "56"}

	bets := []types.Bet{
		{ID: 1, MemberID: 100, Number: "847", BetType: types.BetType3Top, Amount: 100, Rate: 900, Status: types.BetStatusPending},
		{ID: 2, MemberID: 100, Number: "47", BetType: types.BetType2Top, Amount: 50, Rate: 90, Status: types.BetStatusPending},
		{ID: 3, MemberID: 200, Number: "123", BetType: types.BetType3Top, Amount: 100, Rate: 900, Status: types.BetStatusPending},
		{ID: 4, MemberID: 200, Number: "56", BetType: types.BetType2Bottom, Amount: 200, Rate: 90, Status: types.BetStatusPending},
		{ID: 5, MemberID: 300, Number: "999", BetType: types.BetType3Top, Amount: 50, Rate: 900, Status: types.BetStatusPending},
	}

	output := SettleRound(SettleRoundInput{RoundID: 1, Result: result, Bets: bets})

	// member 100: bet1 ชนะ (847 = 3top → 100×900 = 90000) + bet2 ชนะ (47 = 2top → 50×90 = 4500)
	// member 200: bet3 แพ้ + bet4 ชนะ (56 = 2bottom → 200×90 = 18000)
	// member 300: bet5 แพ้
	// Winners: 3, Losers: 2, TotalWin: 90000+4500+18000 = 112500

	if output.TotalWinners != 3 {
		t.Errorf("TotalWinners = %d, want 3", output.TotalWinners)
	}
	if output.TotalLosers != 2 {
		t.Errorf("TotalLosers = %d, want 2", output.TotalLosers)
	}
	if output.TotalWinAmount != 112500 {
		t.Errorf("TotalWinAmount = %.2f, want 112500", output.TotalWinAmount)
	}
	if output.TotalBetAmount != 500 {
		t.Errorf("TotalBetAmount = %.2f, want 500", output.TotalBetAmount)
	}
	if output.Profit != 500-112500 {
		t.Errorf("Profit = %.2f, want %.2f", output.Profit, 500-112500.0)
	}
}

func TestGroupWinnersByMember(t *testing.T) {
	bets := []types.Bet{
		{ID: 1, MemberID: 100, Number: "847", BetType: types.BetType3Top, Amount: 100, Rate: 900},
		{ID: 2, MemberID: 100, Number: "47", BetType: types.BetType2Top, Amount: 50, Rate: 90},
		{ID: 3, MemberID: 200, Number: "56", BetType: types.BetType2Bottom, Amount: 200, Rate: 90},
	}
	betResults := []types.BetResult{
		{BetID: 1, IsWin: true, WinAmount: 90000},
		{BetID: 2, IsWin: true, WinAmount: 4500},
		{BetID: 3, IsWin: true, WinAmount: 18000},
	}

	payouts := GroupWinnersByMember(bets, betResults)

	if payouts[100] != 94500 { // 90000 + 4500
		t.Errorf("member 100 payout = %.2f, want 94500", payouts[100])
	}
	if payouts[200] != 18000 {
		t.Errorf("member 200 payout = %.2f, want 18000", payouts[200])
	}
}
