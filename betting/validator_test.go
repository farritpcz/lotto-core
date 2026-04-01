package betting

import (
	"testing"

	"github.com/farritpcz/lotto-core/types"
)

func TestValidateNumber(t *testing.T) {
	tests := []struct {
		name    string
		number  string
		betType types.BetType
		wantErr bool
	}{
		// 3 ตัวบน — ต้อง 3 หลัก
		{"3top valid", "847", types.BetType3Top, false},
		{"3top with leading zero", "007", types.BetType3Top, false},
		{"3top too short", "47", types.BetType3Top, true},
		{"3top too long", "1234", types.BetType3Top, true},
		{"3top has letters", "abc", types.BetType3Top, true},

		// 2 ตัวบน — ต้อง 2 หลัก
		{"2top valid", "47", types.BetType2Top, false},
		{"2top with leading zero", "05", types.BetType2Top, false},
		{"2top wrong digits", "847", types.BetType2Top, true},

		// 2 ตัวล่าง — ต้อง 2 หลัก
		{"2bottom valid", "56", types.BetType2Bottom, false},

		// วิ่งบน — ต้อง 1 หลัก
		{"run_top valid", "4", types.BetTypeRunTop, false},
		{"run_top zero", "0", types.BetTypeRunTop, false},
		{"run_top too long", "42", types.BetTypeRunTop, true},

		// 3 ตัวโต๊ด — ต้อง 3 หลัก
		{"3tod valid", "748", types.BetType3Tod, false},

		// invalid bet type
		{"invalid bet type", "123", types.BetType("INVALID"), true},

		// empty / special
		{"empty string", "", types.BetType3Top, true},
		{"spaces", "8 7", types.BetType3Top, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNumber(tt.number, tt.betType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNumber(%q, %q) error = %v, wantErr %v", tt.number, tt.betType, err, tt.wantErr)
			}
		})
	}
}

func TestValidateAmount(t *testing.T) {
	tests := []struct {
		name    string
		amount  float64
		minBet  float64
		maxBet  float64
		wantErr bool
	}{
		{"valid amount", 100, 1, 1000, false},
		{"exact min", 1, 1, 1000, false},
		{"exact max", 1000, 1, 1000, false},
		{"zero amount", 0, 1, 1000, true},
		{"negative", -100, 1, 1000, true},
		{"below min", 0.5, 1, 1000, true},
		{"above max", 1001, 1, 1000, true},
		{"unlimited max", 99999, 1, 0, false},  // maxBet=0 ไม่จำกัด
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAmount(tt.amount, tt.minBet, tt.maxBet)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAmount(%.2f, %.2f, %.2f) error = %v, wantErr %v", tt.amount, tt.minBet, tt.maxBet, err, tt.wantErr)
			}
		})
	}
}

func TestValidateYeekeeShoot(t *testing.T) {
	tests := []struct {
		name    string
		number  string
		wantErr bool
	}{
		{"valid 5 digits", "12345", false},
		{"all zeros", "00000", false},
		{"max value", "99999", false},
		{"too short", "1234", true},
		{"too long", "123456", true},
		{"has letters", "abcde", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateYeekeeShoot(tt.number)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateYeekeeShoot(%q) error = %v, wantErr %v", tt.number, err, tt.wantErr)
			}
		})
	}
}
