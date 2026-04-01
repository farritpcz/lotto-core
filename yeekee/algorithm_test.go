package yeekee

import (
	"testing"

	"github.com/farritpcz/lotto-core/types"
)

func TestCalculateResult(t *testing.T) {
	tests := []struct {
		name       string
		shoots     []types.YeekeeShoot
		wantResult string
		wantTop3   string
		wantTop2   string
		wantBot2   string
		wantErr    bool
	}{
		{
			name: "basic calculation",
			shoots: []types.YeekeeShoot{
				{Number: "12345"},
				{Number: "67890"},
				{Number: "11111"},
			},
			// sum = 12345 + 67890 + 11111 = 91346
			wantResult: "91346",
			wantTop3:   "346",
			wantTop2:   "46",
			wantBot2:   "13",
			wantErr:    false,
		},
		{
			name: "result with mod",
			shoots: []types.YeekeeShoot{
				{Number: "99999"},
				{Number: "99999"},
			},
			// sum = 99999 + 99999 = 199998
			// mod 100000 = 99998
			wantResult: "99998",
			wantTop3:   "998",
			wantTop2:   "98",
			wantBot2:   "99",
			wantErr:    false,
		},
		{
			name: "small numbers with leading zeros",
			shoots: []types.YeekeeShoot{
				{Number: "00001"},
				{Number: "00002"},
			},
			// sum = 1 + 2 = 3
			wantResult: "00003",
			wantTop3:   "003",
			wantTop2:   "03",
			wantBot2:   "00",
			wantErr:    false,
		},
		{
			name:    "no shoots — error",
			shoots:  []types.YeekeeShoot{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultNum, roundResult, err := CalculateResult(tt.shoots)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CalculateResult() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if resultNum != tt.wantResult {
				t.Errorf("resultNumber = %q, want %q", resultNum, tt.wantResult)
			}
			if roundResult.Top3 != tt.wantTop3 {
				t.Errorf("Top3 = %q, want %q", roundResult.Top3, tt.wantTop3)
			}
			if roundResult.Top2 != tt.wantTop2 {
				t.Errorf("Top2 = %q, want %q", roundResult.Top2, tt.wantTop2)
			}
			if roundResult.Bottom2 != tt.wantBot2 {
				t.Errorf("Bottom2 = %q, want %q", roundResult.Bottom2, tt.wantBot2)
			}
		})
	}
}

func TestExtractResult(t *testing.T) {
	tests := []struct {
		input    string
		wantTop3 string
		wantTop2 string
		wantBot2 string
	}{
		{"83456", "456", "56", "34"},
		{"00123", "123", "23", "01"},
		{"99999", "999", "99", "99"},
		{"10000", "000", "00", "00"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ExtractResult(tt.input)
			if got.Top3 != tt.wantTop3 {
				t.Errorf("Top3 = %q, want %q", got.Top3, tt.wantTop3)
			}
			if got.Top2 != tt.wantTop2 {
				t.Errorf("Top2 = %q, want %q", got.Top2, tt.wantTop2)
			}
			if got.Bottom2 != tt.wantBot2 {
				t.Errorf("Bottom2 = %q, want %q", got.Bottom2, tt.wantBot2)
			}
		})
	}
}
