// Package lottery กำหนดกฎของหวยแต่ละประเภท
//
// ความสัมพันธ์:
// - ใช้ types.LotteryType, types.BetType
// - ถูกเรียกโดยทุก API เพื่อเช็คว่า:
//   - ประเภทหวยนี้รองรับ BetType ไหนบ้าง
//   - Default rate จ่ายเท่าไหร่
//   - schedule ออกผลอย่างไร
package lottery

import (
	"github.com/farritpcz/lotto-core/types"
)

// =============================================================================
// Rules — กฎหวยแต่ละประเภท
// =============================================================================

// LotteryRule กฎของหวย 1 ประเภท
type LotteryRule struct {
	Type            types.LotteryType // ประเภทหวย
	Name            string            // ชื่อแสดงผล
	AllowedBetTypes []types.BetType   // ประเภทการแทงที่รองรับ
	DefaultRates    map[types.BetType]float64 // rate จ่ายเริ่มต้น
	IsAutoResult    bool              // ระบบออกผลเองหรือไม่ (true = ยี่กี)
	Description     string            // คำอธิบาย
}

// DefaultRules กฎเริ่มต้นของหวยทุกประเภท
//
// NOTE: rate เหล่านี้เป็นค่า default — admin สามารถปรับได้ผ่าน admin panel
// operator (provider mode) ก็ตั้ง rate ของตัวเองได้ผ่าน dashboard
//
// rate หมายถึง: แทง 1 บาท ถ้าถูกได้กี่บาท
// เช่น rate 900 = แทง 1 บาท ถ้าถูกได้ 900 บาท
// ── Rate presets ────────────────────────────────────────────────────────
// ใช้สำหรับสร้าง rules โดยไม่ต้องก๊อปปี้ rates ซ้ำทุกประเภท
var (
	// หวยไทย (รัฐบาล) — มี 3ตัวหน้า, 3ตัวล่าง, 4ตัว
	thaiFullBetTypes = []types.BetType{
		types.BetType3Top, types.BetType3Tod, types.BetType3Front, types.BetType3Bottom,
		types.BetType4Top, types.BetType4Tod,
		types.BetType2Top, types.BetType2Bottom, types.BetTypeRunTop, types.BetTypeRunBot,
	}
	thaiFullRates = map[types.BetType]float64{
		types.BetType3Top: 900, types.BetType3Tod: 150,
		types.BetType3Front: 450, types.BetType3Bottom: 450,
		types.BetType4Top: 6000, types.BetType4Tod: 250,
		types.BetType2Top: 90, types.BetType2Bottom: 90,
		types.BetTypeRunTop: 3.2, types.BetTypeRunBot: 4.2,
	}

	// หวยทั่วไป (ลาว, ฮานอย, มาเลย์, หุ้น) — 6 bet types
	standardBetTypes = []types.BetType{
		types.BetType3Top, types.BetType3Tod,
		types.BetType2Top, types.BetType2Bottom,
		types.BetTypeRunTop, types.BetTypeRunBot,
	}
	standardRates = map[types.BetType]float64{
		types.BetType3Top: 850, types.BetType3Tod: 120,
		types.BetType2Top: 90, types.BetType2Bottom: 90,
		types.BetTypeRunTop: 3.2, types.BetTypeRunBot: 4.2,
	}
)

// newRule สร้าง LotteryRule แบบย่อ
func newRule(lt types.LotteryType, name string, betTypes []types.BetType, rates map[types.BetType]float64, auto bool, desc string) LotteryRule {
	return LotteryRule{Type: lt, Name: name, AllowedBetTypes: betTypes, DefaultRates: rates, IsAutoResult: auto, Description: desc}
}

// DefaultRules กฎเริ่มต้นของหวยทุกประเภท (39 types)
var DefaultRules = map[types.LotteryType]LotteryRule{
	// ── หวยไทย ────────────────────────────────────────────────
	types.LotteryTypeThaiGov: newRule(types.LotteryTypeThaiGov, "หวยรัฐบาลไทย", thaiFullBetTypes, thaiFullRates, false, "ออกผลวันที่ 1 และ 16 ของทุกเดือน"),
	types.LotteryTypeBAAC:    newRule(types.LotteryTypeBAAC, "หวย ธกส", thaiFullBetTypes, thaiFullRates, false, "หวยธนาคารเพื่อการเกษตร"),
	types.LotteryTypeGSB:     newRule(types.LotteryTypeGSB, "หวย ออมสิน", thaiFullBetTypes, thaiFullRates, false, "หวยธนาคารออมสิน"),

	// ── ยี่กี ─────────────────────────────────────────────────
	types.LotteryTypeYeekee: newRule(types.LotteryTypeYeekee, "หวยยี่กี", standardBetTypes, standardRates, true, "ออกผลทุก 15 นาที (88 รอบ/วัน)"),

	// ── หวยลาว ────────────────────────────────────────────────
	types.LotteryTypeLaoVIP:      newRule(types.LotteryTypeLaoVIP, "หวยลาว VIP", standardBetTypes, standardRates, false, "หวยลาว VIP"),
	types.LotteryTypeLaoPattana:  newRule(types.LotteryTypeLaoPattana, "หวยลาวพัฒนา", standardBetTypes, standardRates, false, "หวยลาวพัฒนา"),
	types.LotteryTypeLaoStar:     newRule(types.LotteryTypeLaoStar, "หวยลาวสตาร์", standardBetTypes, standardRates, false, "หวยลาวสตาร์"),
	types.LotteryTypeLaoSamakkee: newRule(types.LotteryTypeLaoSamakkee, "หวยลาวสามัคคี", standardBetTypes, standardRates, false, "หวยลาวสามัคคี"),
	types.LotteryTypeLaoThakhek:  newRule(types.LotteryTypeLaoThakhek, "หวยลาวท่าแขก VIP", standardBetTypes, standardRates, false, "หวยลาวท่าแขก VIP"),

	// ── หวยฮานอย ──────────────────────────────────────────────
	types.LotteryTypeHanoi:        newRule(types.LotteryTypeHanoi, "หวยฮานอย", standardBetTypes, standardRates, false, "หวยฮานอย"),
	types.LotteryTypeHanoiVIP:     newRule(types.LotteryTypeHanoiVIP, "หวยฮานอย VIP", standardBetTypes, standardRates, false, "หวยฮานอย VIP"),
	types.LotteryTypeHanoiPattana: newRule(types.LotteryTypeHanoiPattana, "หวยฮานอยพัฒนา", standardBetTypes, standardRates, false, "หวยฮานอยพัฒนา"),

	// ── มาเลย์ ────────────────────────────────────────────────
	types.LotteryTypeMalay: newRule(types.LotteryTypeMalay, "หวยมาเลย์", standardBetTypes, standardRates, false, "หวยมาเลเซีย"),

	// ── หวยหุ้น (26 ตัว) ──────────────────────────────────────
	types.LotteryTypeStockRussiaVIP:   newRule(types.LotteryTypeStockRussiaVIP, "หวยหุ้นรัสเซีย VIP", standardBetTypes, standardRates, false, "หวยหุ้นรัสเซีย VIP"),
	types.LotteryTypeStockDJVIP:       newRule(types.LotteryTypeStockDJVIP, "หวยหุ้นดาวโจนส์ VIP", standardBetTypes, standardRates, false, "หวยหุ้นดาวโจนส์ VIP"),
	types.LotteryTypeStockHSIVIPAM:    newRule(types.LotteryTypeStockHSIVIPAM, "หวยหุ้นฮั่งเส็ง VIP รอบเช้า", standardBetTypes, standardRates, false, "หวยหุ้นฮั่งเส็ง VIP รอบเช้า"),
	types.LotteryTypeStockTaiwanVIP:   newRule(types.LotteryTypeStockTaiwanVIP, "หวยหุ้นไต้หวัน VIP", standardBetTypes, standardRates, false, "หวยหุ้นไต้หวัน VIP"),
	types.LotteryTypeStockKoreaVIP:    newRule(types.LotteryTypeStockKoreaVIP, "หวยหุ้นเกาหลี VIP", standardBetTypes, standardRates, false, "หวยหุ้นเกาหลี VIP"),
	types.LotteryTypeStockHSIVIPPM:    newRule(types.LotteryTypeStockHSIVIPPM, "หวยหุ้นฮั่งเส็ง VIP รอบบ่าย", standardBetTypes, standardRates, false, "หวยหุ้นฮั่งเส็ง VIP รอบบ่าย"),
	types.LotteryTypeStockNikkeiAM:    newRule(types.LotteryTypeStockNikkeiAM, "หวยหุ้นนิเคอิ รอบเช้า", standardBetTypes, standardRates, false, "หวยหุ้นนิเคอิ รอบเช้า"),
	types.LotteryTypeStockChinaAM:     newRule(types.LotteryTypeStockChinaAM, "หวยหุ้นจีน รอบเช้า", standardBetTypes, standardRates, false, "หวยหุ้นจีน รอบเช้า"),
	types.LotteryTypeStockHSIAM:       newRule(types.LotteryTypeStockHSIAM, "หวยหุ้นฮั่งเส็ง รอบเช้า", standardBetTypes, standardRates, false, "หวยหุ้นฮั่งเส็ง รอบเช้า"),
	types.LotteryTypeStockTaiwan:      newRule(types.LotteryTypeStockTaiwan, "หวยหุ้นไต้หวัน", standardBetTypes, standardRates, false, "หวยหุ้นไต้หวัน"),
	types.LotteryTypeStockNikkeiPM:    newRule(types.LotteryTypeStockNikkeiPM, "หวยหุ้นนิเคอิ รอบบ่าย", standardBetTypes, standardRates, false, "หวยหุ้นนิเคอิ รอบบ่าย"),
	types.LotteryTypeStockKorea:       newRule(types.LotteryTypeStockKorea, "หวยหุ้นเกาหลี", standardBetTypes, standardRates, false, "หวยหุ้นเกาหลี"),
	types.LotteryTypeStockChinaPM:     newRule(types.LotteryTypeStockChinaPM, "หวยหุ้นจีน รอบบ่าย", standardBetTypes, standardRates, false, "หวยหุ้นจีน รอบบ่าย"),
	types.LotteryTypeStockHSIPM:       newRule(types.LotteryTypeStockHSIPM, "หวยหุ้นฮั่งเส็ง รอบบ่าย", standardBetTypes, standardRates, false, "หวยหุ้นฮั่งเส็ง รอบบ่าย"),
	types.LotteryTypeStockTHPM:        newRule(types.LotteryTypeStockTHPM, "หวยหุ้นไทย รอบเย็น", standardBetTypes, standardRates, false, "หวยหุ้นไทย รอบเย็น"),
	types.LotteryTypeStockSingapore:   newRule(types.LotteryTypeStockSingapore, "หวยหุ้นสิงคโปร์", standardBetTypes, standardRates, false, "หวยหุ้นสิงคโปร์"),
	types.LotteryTypeStockIndia:       newRule(types.LotteryTypeStockIndia, "หวยหุ้นอินเดีย", standardBetTypes, standardRates, false, "หวยหุ้นอินเดีย"),
	types.LotteryTypeStockUK:          newRule(types.LotteryTypeStockUK, "หวยหุ้นอังกฤษ", standardBetTypes, standardRates, false, "หวยหุ้นอังกฤษ"),
	types.LotteryTypeStockGermany:     newRule(types.LotteryTypeStockGermany, "หวยหุ้นเยอรมัน", standardBetTypes, standardRates, false, "หวยหุ้นเยอรมัน"),
	types.LotteryTypeStockRussia:      newRule(types.LotteryTypeStockRussia, "หวยหุ้นรัสเซีย", standardBetTypes, standardRates, false, "หวยหุ้นรัสเซีย"),
	types.LotteryTypeStockDJ:          newRule(types.LotteryTypeStockDJ, "หวยหุ้นดาวโจนส์", standardBetTypes, standardRates, false, "หวยหุ้นดาวโจนส์"),
	types.LotteryTypeStockGermanyVIP:  newRule(types.LotteryTypeStockGermanyVIP, "หวยหุ้นเยอรมัน VIP", standardBetTypes, standardRates, false, "หวยหุ้นเยอรมัน VIP"),
	types.LotteryTypeStockUKVIP:       newRule(types.LotteryTypeStockUKVIP, "หวยหุ้นอังกฤษ VIP", standardBetTypes, standardRates, false, "หวยหุ้นอังกฤษ VIP"),
	types.LotteryTypeStockNikkeiVIPPM: newRule(types.LotteryTypeStockNikkeiVIPPM, "หวยหุ้นนิเคอิ VIP รอบบ่าย", standardBetTypes, standardRates, false, "หวยหุ้นนิเคอิ VIP รอบบ่าย"),
	types.LotteryTypeStockNikkeiVIPAM: newRule(types.LotteryTypeStockNikkeiVIPAM, "หวยหุ้นนิเคอิ VIP รอบเช้า", standardBetTypes, standardRates, false, "หวยหุ้นนิเคอิ VIP รอบเช้า"),
	types.LotteryTypeStockChinaVIPPM:  newRule(types.LotteryTypeStockChinaVIPPM, "หวยหุ้นจีน VIP รอบบ่าย", standardBetTypes, standardRates, false, "หวยหุ้นจีน VIP รอบบ่าย"),
	types.LotteryTypeStockChinaVIPAM:  newRule(types.LotteryTypeStockChinaVIPAM, "หวยหุ้นจีน VIP รอบเช้า", standardBetTypes, standardRates, false, "หวยหุ้นจีน VIP รอบเช้า"),
}

// GetRule ดึงกฎของหวยประเภทที่ต้องการ
//
// ตัวอย่าง:
//
//	rule, ok := GetRule(LotteryTypeThai)
//	if ok {
//	    fmt.Println(rule.Name)           // "หวยไทย (ใต้ดิน)"
//	    fmt.Println(rule.DefaultRates)   // map[3TOP:900 3TOD:150 ...]
//	}
func GetRule(lotteryType types.LotteryType) (LotteryRule, bool) {
	rule, ok := DefaultRules[lotteryType]
	return rule, ok
}

// IsBetTypeAllowed ตรวจสอบว่าประเภทหวยนี้รองรับ BetType นี้หรือไม่
//
// ตัวอย่าง:
//
//	IsBetTypeAllowed(LotteryTypeThai, BetType3Top)  → true
//	IsBetTypeAllowed(LotteryTypeThai, BetType3Bottom) → false (หวยไทยไม่มี 3 ตัวล่าง)
func IsBetTypeAllowed(lotteryType types.LotteryType, betType types.BetType) bool {
	rule, ok := GetRule(lotteryType)
	if !ok {
		return false
	}

	for _, allowed := range rule.AllowedBetTypes {
		if allowed == betType {
			return true
		}
	}
	return false
}

// GetDefaultRate ดึง rate จ่ายเริ่มต้นของประเภทหวย + ประเภทการแทง
//
// ตัวอย่าง:
//
//	rate, ok := GetDefaultRate(LotteryTypeThai, BetType3Top)
//	// rate = 900, ok = true
func GetDefaultRate(lotteryType types.LotteryType, betType types.BetType) (float64, bool) {
	rule, ok := GetRule(lotteryType)
	if !ok {
		return 0, false
	}

	rate, ok := rule.DefaultRates[betType]
	return rate, ok
}
