package lottery

import (
	"testing"

	"github.com/farritpcz/lotto-core/types"
)

// =============================================================================
// TestDefaultRules_AllLotteryTypes — ทุกประเภทหวยที่ define ใน enum ต้องมี rules
// ⚠️ ถ้าขาด → เปิดหวยใหม่แล้ว bet ไม่ได้
// =============================================================================

func TestDefaultRules_AllLotteryTypes(t *testing.T) {
	// ทุก LotteryType ที่ควรมี rules (ยกเว้น CUSTOM ที่ admin ตั้งเอง)
	requiredTypes := []types.LotteryType{
		types.LotteryTypeThai,
		types.LotteryTypeLao,
		types.LotteryTypeHanoi,
		types.LotteryTypeMalay,
		types.LotteryTypeLao9,
		types.LotteryTypeBAAC,
		types.LotteryTypeGSB,
		types.LotteryTypeStockTH,
		types.LotteryTypeStockForeign,
		types.LotteryTypeYeekee,
		types.LotteryTypeYeekee5,
		types.LotteryTypeYeekee15,
		types.LotteryTypeYeekeeVIP,
	}

	for _, lt := range requiredTypes {
		t.Run(string(lt), func(t *testing.T) {
			rule, ok := GetRule(lt)
			if !ok {
				t.Errorf("GetRule(%s) not found — lottery type has no default rules!", lt)
				return
			}

			// ต้องมี AllowedBetTypes
			if len(rule.AllowedBetTypes) == 0 {
				t.Errorf("%s has empty AllowedBetTypes", lt)
			}

			// ต้องมี DefaultRates สำหรับทุก allowed bet type
			for _, bt := range rule.AllowedBetTypes {
				rate, exists := rule.DefaultRates[bt]
				if !exists {
					t.Errorf("%s: missing DefaultRate for %s", lt, bt)
				}
				if rate <= 0 {
					t.Errorf("%s: rate for %s = %.2f, must be > 0", lt, bt, rate)
				}
			}

			// ต้องมี Name
			if rule.Name == "" {
				t.Errorf("%s has empty Name", lt)
			}
		})
	}
}

// =============================================================================
// TestThaiLottery_Has4DigitTypes — หวยไทยต้องรองรับ 4TOP/4TOD/3FRONT/3BOTTOM
// =============================================================================

func TestThaiLottery_Has4DigitTypes(t *testing.T) {
	// หวยไทยต้องมี bet types ใหม่ที่เพิ่ง add
	required := []types.BetType{
		types.BetType3Front,
		types.BetType3Bottom,
		types.BetType4Top,
		types.BetType4Tod,
	}

	for _, bt := range required {
		if !IsBetTypeAllowed(types.LotteryTypeThai, bt) {
			t.Errorf("Thai lottery should allow %s", bt)
		}

		rate, ok := GetDefaultRate(types.LotteryTypeThai, bt)
		if !ok {
			t.Errorf("Thai lottery missing default rate for %s", bt)
		}
		if rate <= 0 {
			t.Errorf("Thai lottery rate for %s = %.2f, should be > 0", bt, rate)
		}
	}
}

// =============================================================================
// TestYeekeeTypes_AreAutoResult — ยี่กีทุก variant ต้อง IsAutoResult = true
// =============================================================================

func TestYeekeeTypes_AreAutoResult(t *testing.T) {
	yeekeeTypes := []types.LotteryType{
		types.LotteryTypeYeekee,
		types.LotteryTypeYeekee5,
		types.LotteryTypeYeekee15,
		types.LotteryTypeYeekeeVIP,
	}

	for _, lt := range yeekeeTypes {
		t.Run(string(lt), func(t *testing.T) {
			rule, ok := GetRule(lt)
			if !ok {
				t.Fatalf("%s has no rule", lt)
			}
			if !rule.IsAutoResult {
				t.Errorf("%s.IsAutoResult = false, want true", lt)
			}
		})
	}
}

// =============================================================================
// TestNonYeekeeTypes_AreManualResult — หวยไม่ใช่ยี่กี ต้อง IsAutoResult = false
// =============================================================================

func TestNonYeekeeTypes_AreManualResult(t *testing.T) {
	manualTypes := []types.LotteryType{
		types.LotteryTypeThai,
		types.LotteryTypeLao,
		types.LotteryTypeHanoi,
		types.LotteryTypeMalay,
		types.LotteryTypeLao9,
		types.LotteryTypeBAAC,
		types.LotteryTypeGSB,
		types.LotteryTypeStockTH,
		types.LotteryTypeStockForeign,
	}

	for _, lt := range manualTypes {
		t.Run(string(lt), func(t *testing.T) {
			rule, ok := GetRule(lt)
			if !ok {
				t.Fatalf("%s has no rule", lt)
			}
			if rule.IsAutoResult {
				t.Errorf("%s.IsAutoResult = true, want false (admin should input result)", lt)
			}
		})
	}
}

// =============================================================================
// TestIsBetTypeAllowed_NonExistent — เช็คว่า lottery type ที่ไม่มี → return false
// =============================================================================

func TestIsBetTypeAllowed_NonExistent(t *testing.T) {
	if IsBetTypeAllowed("NONEXISTENT", types.BetType3Top) {
		t.Error("non-existent lottery type should return false")
	}
}

func TestGetDefaultRate_NonExistent(t *testing.T) {
	_, ok := GetDefaultRate("NONEXISTENT", types.BetType3Top)
	if ok {
		t.Error("non-existent lottery type should return ok=false")
	}
}
