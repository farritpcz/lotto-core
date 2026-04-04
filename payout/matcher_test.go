package payout

import (
	"math"
	"testing"

	"github.com/farritpcz/lotto-core/types"
)

// =============================================================================
// TestMatch — ครอบคลุมทุก BetType + edge cases
// ⚠️ ระบบเกี่ยวกับเงิน — ต้อง test ให้รอบคอบ
// =============================================================================

func TestMatch(t *testing.T) {
	// ─── ผลรางวัลตัวอย่าง ──────────────────────────────────────
	// รางวัลที่ 1: 491847 → Front3="491", Top3="847", Top2="47"
	// 2 ตัวล่าง: "56"
	// 3 ตัวล่าง: "123,456" (2 รางวัล)
	result := types.RoundResult{
		Top3:    "847",
		Top2:    "47",
		Bottom2: "56",
		Front3:  "491",
		Bottom3: "123,456",
	}

	tests := []struct {
		name      string
		bet       types.Bet
		wantWin   bool
		wantPay   float64 // expected WinAmount (0 = don't check exact)
	}{
		// ── 3 ตัวบน ─────────────────────────────────────────
		{"3top win exact", types.Bet{ID: 1, Number: "847", BetType: types.BetType3Top, Amount: 100, Rate: 900}, true, 90000},
		{"3top lose", types.Bet{ID: 2, Number: "123", BetType: types.BetType3Top, Amount: 100, Rate: 900}, false, 0},
		{"3top lose permuted", types.Bet{ID: 3, Number: "748", BetType: types.BetType3Top, Amount: 100, Rate: 900}, false, 0},

		// ── 3 ตัวโต๊ด ───────────────────────────────────────
		{"3tod win exact", types.Bet{ID: 10, Number: "847", BetType: types.BetType3Tod, Amount: 100, Rate: 150}, true, 15000},
		{"3tod win perm1", types.Bet{ID: 11, Number: "748", BetType: types.BetType3Tod, Amount: 100, Rate: 150}, true, 15000},
		{"3tod win perm2", types.Bet{ID: 12, Number: "478", BetType: types.BetType3Tod, Amount: 100, Rate: 150}, true, 15000},
		{"3tod win perm3", types.Bet{ID: 13, Number: "874", BetType: types.BetType3Tod, Amount: 100, Rate: 150}, true, 15000},
		{"3tod lose", types.Bet{ID: 14, Number: "123", BetType: types.BetType3Tod, Amount: 100, Rate: 150}, false, 0},
		{"3tod lose same digits diff count", types.Bet{ID: 15, Number: "884", BetType: types.BetType3Tod, Amount: 100, Rate: 150}, false, 0},

		// ── 3 ตัวหน้า ───────────────────────────────────────
		{"3front win", types.Bet{ID: 20, Number: "491", BetType: types.BetType3Front, Amount: 50, Rate: 450}, true, 22500},
		{"3front lose permuted", types.Bet{ID: 21, Number: "194", BetType: types.BetType3Front, Amount: 50, Rate: 450}, false, 0},
		{"3front lose top3", types.Bet{ID: 22, Number: "847", BetType: types.BetType3Front, Amount: 50, Rate: 450}, false, 0},
		{"3front lose random", types.Bet{ID: 23, Number: "999", BetType: types.BetType3Front, Amount: 50, Rate: 450}, false, 0},

		// ── 3 ตัวล่าง (multi-prize) ─────────────────────────
		{"3bottom win first prize", types.Bet{ID: 30, Number: "123", BetType: types.BetType3Bottom, Amount: 100, Rate: 450}, true, 45000},
		{"3bottom win second prize", types.Bet{ID: 31, Number: "456", BetType: types.BetType3Bottom, Amount: 100, Rate: 450}, true, 45000},
		{"3bottom lose", types.Bet{ID: 32, Number: "789", BetType: types.BetType3Bottom, Amount: 100, Rate: 450}, false, 0},
		{"3bottom lose permuted of prize", types.Bet{ID: 33, Number: "321", BetType: types.BetType3Bottom, Amount: 100, Rate: 450}, false, 0},
		{"3bottom lose permuted of prize2", types.Bet{ID: 34, Number: "654", BetType: types.BetType3Bottom, Amount: 100, Rate: 450}, false, 0},

		// ── 4 ตัวบน (Full=491847, last4=1847) ───────────────
		{"4top win", types.Bet{ID: 40, Number: "1847", BetType: types.BetType4Top, Amount: 10, Rate: 6000}, true, 60000},
		{"4top lose front4", types.Bet{ID: 41, Number: "4918", BetType: types.BetType4Top, Amount: 10, Rate: 6000}, false, 0},
		{"4top lose reversed", types.Bet{ID: 42, Number: "7481", BetType: types.BetType4Top, Amount: 10, Rate: 6000}, false, 0},
		{"4top lose random", types.Bet{ID: 43, Number: "1234", BetType: types.BetType4Top, Amount: 10, Rate: 6000}, false, 0},

		// ── 4 ตัวโต๊ด (last4=1847) ──────────────────────────
		{"4tod win exact", types.Bet{ID: 50, Number: "1847", BetType: types.BetType4Tod, Amount: 10, Rate: 250}, true, 2500},
		{"4tod win perm1", types.Bet{ID: 51, Number: "7184", BetType: types.BetType4Tod, Amount: 10, Rate: 250}, true, 2500},
		{"4tod win perm2", types.Bet{ID: 52, Number: "8174", BetType: types.BetType4Tod, Amount: 10, Rate: 250}, true, 2500},
		{"4tod win perm3", types.Bet{ID: 53, Number: "4718", BetType: types.BetType4Tod, Amount: 10, Rate: 250}, true, 2500},
		{"4tod lose", types.Bet{ID: 54, Number: "1234", BetType: types.BetType4Tod, Amount: 10, Rate: 250}, false, 0},
		{"4tod lose close numbers", types.Bet{ID: 55, Number: "1848", BetType: types.BetType4Tod, Amount: 10, Rate: 250}, false, 0},

		// ── 2 ตัวบน ─────────────────────────────────────────
		{"2top win", types.Bet{ID: 60, Number: "47", BetType: types.BetType2Top, Amount: 100, Rate: 90}, true, 9000},
		{"2top lose reversed", types.Bet{ID: 61, Number: "74", BetType: types.BetType2Top, Amount: 100, Rate: 90}, false, 0},
		{"2top lose first2", types.Bet{ID: 62, Number: "84", BetType: types.BetType2Top, Amount: 100, Rate: 90}, false, 0},

		// ── 2 ตัวล่าง ───────────────────────────────────────
		{"2bottom win", types.Bet{ID: 70, Number: "56", BetType: types.BetType2Bottom, Amount: 100, Rate: 90}, true, 9000},
		{"2bottom lose reversed", types.Bet{ID: 71, Number: "65", BetType: types.BetType2Bottom, Amount: 100, Rate: 90}, false, 0},

		// ── วิ่งบน (Top3=847 → digits 8,4,7) ───────────────
		{"run_top win 8", types.Bet{ID: 80, Number: "8", BetType: types.BetTypeRunTop, Amount: 100, Rate: 3.2}, true, 320},
		{"run_top win 4", types.Bet{ID: 81, Number: "4", BetType: types.BetTypeRunTop, Amount: 100, Rate: 3.2}, true, 320},
		{"run_top win 7", types.Bet{ID: 82, Number: "7", BetType: types.BetTypeRunTop, Amount: 100, Rate: 3.2}, true, 320},
		{"run_top lose 1", types.Bet{ID: 83, Number: "1", BetType: types.BetTypeRunTop, Amount: 100, Rate: 3.2}, false, 0},
		{"run_top lose 0", types.Bet{ID: 84, Number: "0", BetType: types.BetTypeRunTop, Amount: 100, Rate: 3.2}, false, 0},

		// ── วิ่งล่าง (Bottom2=56 → digits 5,6) ─────────────
		{"run_bot win 5", types.Bet{ID: 90, Number: "5", BetType: types.BetTypeRunBot, Amount: 100, Rate: 4.2}, true, 420},
		{"run_bot win 6", types.Bet{ID: 91, Number: "6", BetType: types.BetTypeRunBot, Amount: 100, Rate: 4.2}, true, 420},
		{"run_bot lose 1", types.Bet{ID: 92, Number: "1", BetType: types.BetTypeRunBot, Amount: 100, Rate: 4.2}, false, 0},
		{"run_bot lose 0", types.Bet{ID: 93, Number: "0", BetType: types.BetTypeRunBot, Amount: 100, Rate: 4.2}, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Match(tt.bet, result)

			// เช็ค win/lose
			if got.IsWin != tt.wantWin {
				t.Errorf("Match() IsWin = %v, want %v (number=%s, betType=%s)",
					got.IsWin, tt.wantWin, tt.bet.Number, tt.bet.BetType)
			}

			// เช็ค status
			if tt.wantWin && got.Status != types.BetStatusWon {
				t.Errorf("Match() Status = %v, want 'won'", got.Status)
			}
			if !tt.wantWin && got.Status != types.BetStatusLost {
				t.Errorf("Match() Status = %v, want 'lost'", got.Status)
			}

			// ⚠️ เช็คเงินรางวัลตรงเป๊ะ — ห้ามผิดแม้แต่สตางค์
			if tt.wantPay > 0 {
				if math.Abs(got.WinAmount-tt.wantPay) > 0.001 {
					t.Errorf("Match() WinAmount = %.2f, want %.2f (amount=%.0f × rate=%.1f)",
						got.WinAmount, tt.wantPay, tt.bet.Amount, tt.bet.Rate)
				}
			}

			// เช็คว่า lose ต้อง WinAmount = 0 เสมอ
			if !tt.wantWin && got.WinAmount != 0 {
				t.Errorf("Match() WinAmount = %.2f, want 0 for losing bet", got.WinAmount)
			}

			// เช็ค BetID ต้องตรงเสมอ
			if got.BetID != tt.bet.ID {
				t.Errorf("Match() BetID = %d, want %d", got.BetID, tt.bet.ID)
			}
		})
	}
}

// =============================================================================
// TestMatch_EdgeCases — edge cases ที่อาจทำให้ระบบจ่ายเงินผิด
// =============================================================================

func TestMatch_EdgeCases(t *testing.T) {
	t.Run("leading zeros in numbers", func(t *testing.T) {
		// ผล 3 ตัวบน = "007" → ต้อง match เลข "007" ไม่ใช่ "7"
		result := types.RoundResult{Top3: "007", Top2: "07", Bottom2: "03"}

		// "007" ต้องถูก 3 ตัวบน
		got := Match(types.Bet{ID: 1, Number: "007", BetType: types.BetType3Top, Amount: 100, Rate: 900}, result)
		if !got.IsWin {
			t.Error("007 should win 3TOP when result is 007")
		}

		// "7" ไม่ถูก 3 ตัวบน (ต้อง 3 หลักเท่ากัน)
		got = Match(types.Bet{ID: 2, Number: "7", BetType: types.BetType3Top, Amount: 100, Rate: 900}, result)
		if got.IsWin {
			t.Error("7 should NOT win 3TOP when result is 007")
		}

		// "0" ต้องถูกวิ่งบน (0 อยู่ใน "007")
		got = Match(types.Bet{ID: 3, Number: "0", BetType: types.BetTypeRunTop, Amount: 100, Rate: 3.2}, result)
		if !got.IsWin {
			t.Error("0 should win RUN_TOP when result is 007")
		}

		// "07" ต้องถูก 2 ตัวบน
		got = Match(types.Bet{ID: 4, Number: "07", BetType: types.BetType2Top, Amount: 100, Rate: 90}, result)
		if !got.IsWin {
			t.Error("07 should win 2TOP when result is 07")
		}
	})

	t.Run("3bottom single prize", func(t *testing.T) {
		// Bottom3 มีแค่รางวัลเดียว (ไม่มี comma)
		result := types.RoundResult{Top3: "111", Top2: "11", Bottom2: "22", Bottom3: "333"}

		got := Match(types.Bet{ID: 1, Number: "333", BetType: types.BetType3Bottom, Amount: 100, Rate: 450}, result)
		if !got.IsWin {
			t.Error("333 should win 3BOTTOM when Bottom3 is 333")
		}

		got = Match(types.Bet{ID: 2, Number: "222", BetType: types.BetType3Bottom, Amount: 100, Rate: 450}, result)
		if got.IsWin {
			t.Error("222 should NOT win 3BOTTOM when Bottom3 is 333")
		}
	})

	t.Run("3bottom empty", func(t *testing.T) {
		// Bottom3 ว่าง → ไม่มีใครถูก 3 ตัวล่าง (ยี่กี/หุ้น ไม่มีรางวัล 3 ตัวล่าง)
		result := types.RoundResult{Top3: "111", Top2: "11", Bottom2: "22", Bottom3: ""}

		got := Match(types.Bet{ID: 1, Number: "111", BetType: types.BetType3Bottom, Amount: 100, Rate: 450}, result)
		if got.IsWin {
			t.Error("should NOT win 3BOTTOM when Bottom3 is empty")
		}
	})

	t.Run("3front empty", func(t *testing.T) {
		// Front3 ว่าง → ไม่มีใครถูก 3 ตัวหน้า
		result := types.RoundResult{Top3: "111", Top2: "11", Bottom2: "22", Front3: ""}

		got := Match(types.Bet{ID: 1, Number: "111", BetType: types.BetType3Front, Amount: 100, Rate: 450}, result)
		if got.IsWin {
			t.Error("should NOT win 3FRONT when Front3 is empty")
		}
	})

	t.Run("4top without front3", func(t *testing.T) {
		// ไม่มี Front3 → full = Top3 เอง = "847" (3 หลัก < 4 หลัก → 4TOP ไม่ match)
		result := types.RoundResult{Top3: "847", Top2: "47", Bottom2: "56"}

		got := Match(types.Bet{ID: 1, Number: "0847", BetType: types.BetType4Top, Amount: 100, Rate: 6000}, result)
		if got.IsWin {
			t.Error("4TOP should NOT match when result has only 3 digits (no Front3)")
		}
	})

	t.Run("3bottom with spaces in comma list", func(t *testing.T) {
		// Bottom3 มี space หลัง comma → ต้อง trim แล้วยัง match ได้
		result := types.RoundResult{Top3: "111", Top2: "11", Bottom2: "22", Bottom3: "123, 456, 789"}

		got := Match(types.Bet{ID: 1, Number: "456", BetType: types.BetType3Bottom, Amount: 100, Rate: 450}, result)
		if !got.IsWin {
			t.Error("456 should win 3BOTTOM even with spaces in comma list")
		}

		got = Match(types.Bet{ID: 2, Number: "789", BetType: types.BetType3Bottom, Amount: 100, Rate: 450}, result)
		if !got.IsWin {
			t.Error("789 should win 3BOTTOM")
		}
	})

	t.Run("settled bets skipped in MatchAll", func(t *testing.T) {
		// bet ที่ cancelled/refunded ต้องถูกข้าม ไม่ settle ซ้ำ
		result := types.RoundResult{Top3: "847", Top2: "47", Bottom2: "56"}
		bets := []types.Bet{
			{ID: 1, Number: "847", BetType: types.BetType3Top, Amount: 100, Rate: 900, Status: types.BetStatusCancelled},
			{ID: 2, Number: "847", BetType: types.BetType3Top, Amount: 100, Rate: 900, Status: types.BetStatusRefunded},
			{ID: 3, Number: "847", BetType: types.BetType3Top, Amount: 100, Rate: 900, Status: types.BetStatusPending},
		}

		results := MatchAll(bets, result)
		// ต้องได้แค่ 1 result (bet 3 เท่านั้น)
		if len(results) != 1 {
			t.Errorf("MatchAll() returned %d results, want 1 (cancelled/refunded should be skipped)", len(results))
		}
		if results[0].BetID != 3 {
			t.Errorf("MatchAll() first result BetID = %d, want 3", results[0].BetID)
		}
	})
}

// =============================================================================
// TestMatch_PayoutAccuracy — ทดสอบความแม่นยำของการคำนวณเงิน
// ⚠️ สำคัญมาก — จ่ายผิดแม้แต่สตางค์ = ปัญหา
// =============================================================================

func TestMatch_PayoutAccuracy(t *testing.T) {
	result := types.RoundResult{Top3: "999", Top2: "99", Bottom2: "99"}

	tests := []struct {
		name    string
		amount  float64
		rate    float64
		wantPay float64
	}{
		// ─── ทดสอบ precision ─────────────────────────────────
		{"small bet", 1, 900, 900},
		{"normal bet", 100, 900, 90000},
		{"large bet", 50000, 900, 45000000},
		{"decimal rate", 100, 3.2, 320},
		{"decimal amount", 10.50, 90, 945},
		{"small rate", 1, 4.2, 4.2},

		// ─── ทดสอบ float precision edge cases ────────────────
		{"0.1 + 0.2 precision", 0.3, 900, 270},
		{"min bet", 1, 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Match(types.Bet{
				ID: 1, Number: "999", BetType: types.BetType3Top,
				Amount: tt.amount, Rate: tt.rate,
			}, result)

			if !got.IsWin {
				t.Fatal("expected win")
			}

			// ⚠️ ตรวจ precision — ต้องตรงภายใน 0.01 บาท
			diff := math.Abs(got.WinAmount - tt.wantPay)
			if diff > 0.01 {
				t.Errorf("WinAmount = %.4f, want %.4f (diff = %.4f) — amount=%.2f × rate=%.2f",
					got.WinAmount, tt.wantPay, diff, tt.amount, tt.rate)
			}
		})
	}
}

// =============================================================================
// TestSettleRound_FullFlow — ทดสอบ SettleRound ครบวงจร
// =============================================================================

func TestSettleRound_FullFlow(t *testing.T) {
	result := types.RoundResult{
		Top3: "847", Top2: "47", Bottom2: "56",
		Front3: "491", Bottom3: "123,456",
	}

	bets := []types.Bet{
		// ── สมาชิก A: แทง 3 ตัวบน 847 ถูก! ──────────────────
		{ID: 1, MemberID: 100, Number: "847", BetType: types.BetType3Top, Amount: 100, Rate: 900, Status: types.BetStatusPending},
		// ── สมาชิก A: แทง 2 ตัวล่าง 56 ถูก! ─────────────────
		{ID: 2, MemberID: 100, Number: "56", BetType: types.BetType2Bottom, Amount: 50, Rate: 90, Status: types.BetStatusPending},
		// ── สมาชิก B: แทง 3 ตัวหน้า 491 ถูก! ────────────────
		{ID: 3, MemberID: 200, Number: "491", BetType: types.BetType3Front, Amount: 100, Rate: 450, Status: types.BetStatusPending},
		// ── สมาชิก C: แทง 3 ตัวบน 123 แพ้ ────────────────────
		{ID: 4, MemberID: 300, Number: "123", BetType: types.BetType3Top, Amount: 100, Rate: 900, Status: types.BetStatusPending},
		// ── สมาชิก C: แทง 3 ตัวล่าง 123 ถูก! ────────────────
		{ID: 5, MemberID: 300, Number: "123", BetType: types.BetType3Bottom, Amount: 100, Rate: 450, Status: types.BetStatusPending},
		// ── สมาชิก D: bet ถูก cancel แล้ว → ต้องข้าม ─────────
		{ID: 6, MemberID: 400, Number: "847", BetType: types.BetType3Top, Amount: 1000, Rate: 900, Status: types.BetStatusCancelled},
	}

	output := SettleRound(SettleRoundInput{
		RoundID: 1,
		Result:  result,
		Bets:    bets,
	})

	// ─── เช็คจำนวน bet results ──────────────────────────────
	// 6 bets แต่ 1 ตัว cancelled → settle 5 ตัว
	if len(output.BetResults) != 5 {
		t.Errorf("BetResults count = %d, want 5", len(output.BetResults))
	}

	// ─── เช็ค winners/losers ────────────────────────────────
	// ถูก: bet 1 (3TOP), bet 2 (2BOT), bet 3 (3FRONT), bet 5 (3BOTTOM) = 4 winners
	// แพ้: bet 4 (3TOP 123) = 1 loser
	if output.TotalWinners != 4 {
		t.Errorf("TotalWinners = %d, want 4", output.TotalWinners)
	}
	if output.TotalLosers != 1 {
		t.Errorf("TotalLosers = %d, want 1", output.TotalLosers)
	}

	// ─── เช็คเงินรางวัลรวม ──────────────────────────────────
	// bet1: 100 × 900 = 90000
	// bet2: 50 × 90 = 4500
	// bet3: 100 × 450 = 45000
	// bet5: 100 × 450 = 45000
	// รวม = 184500
	expectedWin := 90000.0 + 4500.0 + 45000.0 + 45000.0
	if math.Abs(output.TotalWinAmount-expectedWin) > 0.01 {
		t.Errorf("TotalWinAmount = %.2f, want %.2f", output.TotalWinAmount, expectedWin)
	}

	// ─── เช็คยอดแทงรวม (exclude cancelled) ───────────────────
	// bet1: 100, bet2: 50, bet3: 100, bet4: 100, bet5: 100 = 450
	expectedBet := 100.0 + 50.0 + 100.0 + 100.0 + 100.0
	if math.Abs(output.TotalBetAmount-expectedBet) > 0.01 {
		t.Errorf("TotalBetAmount = %.2f, want %.2f", output.TotalBetAmount, expectedBet)
	}

	// ─── เช็คกำไร/ขาดทุน ────────────────────────────────────
	expectedProfit := expectedBet - expectedWin // 450 - 184500 = -184050 (ขาดทุน)
	if math.Abs(output.Profit-expectedProfit) > 0.01 {
		t.Errorf("Profit = %.2f, want %.2f", output.Profit, expectedProfit)
	}

	// ─── เช็ค GroupWinnersByMember ───────────────────────────
	memberPayouts := GroupWinnersByMember(bets, output.BetResults)

	// สมาชิก A (ID=100): bet1 (90000) + bet2 (4500) = 94500
	if math.Abs(memberPayouts[100]-94500) > 0.01 {
		t.Errorf("Member 100 payout = %.2f, want 94500", memberPayouts[100])
	}

	// สมาชิก B (ID=200): bet3 (45000)
	if math.Abs(memberPayouts[200]-45000) > 0.01 {
		t.Errorf("Member 200 payout = %.2f, want 45000", memberPayouts[200])
	}

	// สมาชิก C (ID=300): bet5 (45000) — bet4 แพ้ไม่นับ
	if math.Abs(memberPayouts[300]-45000) > 0.01 {
		t.Errorf("Member 300 payout = %.2f, want 45000", memberPayouts[300])
	}

	// สมาชิก D (ID=400): bet6 cancelled → ไม่ได้เงิน → ไม่มีใน map
	if _, exists := memberPayouts[400]; exists {
		t.Error("Member 400 (cancelled) should NOT be in memberPayouts")
	}
}

// =============================================================================
// TestIsPermutation — ครอบคลุม edge cases
// =============================================================================

func TestIsPermutation(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		// ─── 3 หลัก ─────────────────────────────────────────
		{"847", "847", true},  // ตรงเป๊ะ
		{"847", "748", true},  // สลับ
		{"847", "478", true},  // สลับ
		{"847", "874", true},  // สลับ
		{"847", "123", false}, // ตัวเลขต่าง
		{"847", "884", false}, // จำนวนแต่ละตัวไม่ตรง
		{"000", "000", true},  // เลขซ้ำ
		{"001", "010", true},  // leading zero
		{"001", "100", true},  // leading zero

		// ─── 2 หลัก ─────────────────────────────────────────
		{"12", "21", true},
		{"12", "12", true},
		{"12", "13", false},
		{"00", "00", true},

		// ─── 4 หลัก ─────────────────────────────────────────
		{"1847", "7184", true},
		{"1847", "1234", false},
		{"1111", "1111", true},
		{"1122", "2211", true},
		{"1122", "2112", true},
		{"1122", "1123", false},

		// ─── ความยาวต่างกัน ─────────────────────────────────
		{"12", "123", false},
		{"1", "11", false},
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

// =============================================================================
// TestBuildFullResult — ทดสอบ helper สร้างเลข 6 หลัก
// =============================================================================

func TestBuildFullResult(t *testing.T) {
	tests := []struct {
		name   string
		result types.RoundResult
		want   string
	}{
		{"with front3", types.RoundResult{Front3: "491", Top3: "847"}, "491847"},
		{"without front3", types.RoundResult{Front3: "", Top3: "847"}, "847"},
		{"front3 only zeros", types.RoundResult{Front3: "000", Top3: "001"}, "000001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildFullResult(tt.result)
			if got != tt.want {
				t.Errorf("buildFullResult() = %q, want %q", got, tt.want)
			}
		})
	}
}

// =============================================================================
// TestSummarizeResults — ทดสอบสรุปผลรวม
// =============================================================================

func TestSummarizeResults(t *testing.T) {
	results := []types.BetResult{
		{BetID: 1, IsWin: true, WinAmount: 90000},
		{BetID: 2, IsWin: true, WinAmount: 4500},
		{BetID: 3, IsWin: false, WinAmount: 0},
		{BetID: 4, IsWin: false, WinAmount: 0},
		{BetID: 5, IsWin: true, WinAmount: 320},
	}

	winners, winAmount, losers := SummarizeResults(results)

	if winners != 3 {
		t.Errorf("winners = %d, want 3", winners)
	}
	if losers != 2 {
		t.Errorf("losers = %d, want 2", losers)
	}
	if math.Abs(winAmount-94820) > 0.01 {
		t.Errorf("winAmount = %.2f, want 94820", winAmount)
	}
}
