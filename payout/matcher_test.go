package payout

import (
	"testing"

	"github.com/farritpcz/lotto-core/types"
)

func TestMatch(t *testing.T) {
	result := types.RoundResult{
		Top3:    "847",
		Top2:    "47",
		Bottom2: "56",
	}

	tests := []struct {
		name    string
		bet     types.Bet
		wantWin bool
	}{
		// 3 ตัวบน
		{"3top win", types.Bet{ID: 1, Number: "847", BetType: types.BetType3Top, Amount: 100, Rate: 900}, true},
		{"3top lose", types.Bet{ID: 2, Number: "123", BetType: types.BetType3Top, Amount: 100, Rate: 900}, false},

		// 3 ตัวโต๊ด (สลับตำแหน่งได้)
		{"3tod win exact", types.Bet{ID: 3, Number: "847", BetType: types.BetType3Tod, Amount: 100, Rate: 150}, true},
		{"3tod win permuted", types.Bet{ID: 4, Number: "748", BetType: types.BetType3Tod, Amount: 100, Rate: 150}, true},
		{"3tod win permuted2", types.Bet{ID: 5, Number: "478", BetType: types.BetType3Tod, Amount: 100, Rate: 150}, true},
		{"3tod lose", types.Bet{ID: 6, Number: "123", BetType: types.BetType3Tod, Amount: 100, Rate: 150}, false},

		// 2 ตัวบน
		{"2top win", types.Bet{ID: 7, Number: "47", BetType: types.BetType2Top, Amount: 100, Rate: 90}, true},
		{"2top lose", types.Bet{ID: 8, Number: "84", BetType: types.BetType2Top, Amount: 100, Rate: 90}, false},

		// 2 ตัวล่าง
		{"2bottom win", types.Bet{ID: 9, Number: "56", BetType: types.BetType2Bottom, Amount: 100, Rate: 90}, true},
		{"2bottom lose", types.Bet{ID: 10, Number: "65", BetType: types.BetType2Bottom, Amount: 100, Rate: 90}, false},

		// วิ่งบน (เลข 1 ตัว อยู่ใน Top3 "847")
		{"run_top win 8", types.Bet{ID: 11, Number: "8", BetType: types.BetTypeRunTop, Amount: 100, Rate: 3.2}, true},
		{"run_top win 4", types.Bet{ID: 12, Number: "4", BetType: types.BetTypeRunTop, Amount: 100, Rate: 3.2}, true},
		{"run_top win 7", types.Bet{ID: 13, Number: "7", BetType: types.BetTypeRunTop, Amount: 100, Rate: 3.2}, true},
		{"run_top lose", types.Bet{ID: 14, Number: "1", BetType: types.BetTypeRunTop, Amount: 100, Rate: 3.2}, false},

		// วิ่งล่าง (เลข 1 ตัว อยู่ใน Bottom2 "56")
		{"run_bot win 5", types.Bet{ID: 15, Number: "5", BetType: types.BetTypeRunBot, Amount: 100, Rate: 4.2}, true},
		{"run_bot win 6", types.Bet{ID: 16, Number: "6", BetType: types.BetTypeRunBot, Amount: 100, Rate: 4.2}, true},
		{"run_bot lose", types.Bet{ID: 17, Number: "1", BetType: types.BetTypeRunBot, Amount: 100, Rate: 4.2}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Match(tt.bet, result)
			if got.IsWin != tt.wantWin {
				t.Errorf("Match() IsWin = %v, want %v", got.IsWin, tt.wantWin)
			}
			if tt.wantWin && got.WinAmount <= 0 {
				t.Errorf("Match() WinAmount = %.2f, want > 0", got.WinAmount)
			}
			if !tt.wantWin && got.WinAmount != 0 {
				t.Errorf("Match() WinAmount = %.2f, want 0", got.WinAmount)
			}
		})
	}
}

func TestIsPermutation(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"847", "847", true},
		{"847", "748", true},
		{"847", "478", true},
		{"847", "874", true},
		{"847", "123", false},
		{"847", "884", false},
		{"12", "21", true},
		{"12", "12", true},
		{"12", "13", false},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			got := isPermutation(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("isPermutation(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
