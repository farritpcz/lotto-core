# Betting Rules — กฏการแทงหวย

> Last updated: 2026-04-20
> Related code: `betting/validator.go`, `betting/calculator.go`, `betting/limit.go`, `numberban/checker.go`

## 🎯 Purpose
นิยามกฏ validate bet request, คำนวณเงินจ่าย (rate × amount), และตรวจ limit/เลขอั้น ก่อนบันทึกลง DB — ทุก API (standalone + provider) เรียก function ใน `betting/` แบบเดียวกัน

## 📋 Rules (กฏเงื่อนไข)

### Bet Types (10 ประเภท)

| BetType | DigitCount | อธิบาย |
|---------|-----------|--------|
| `3TOP` | 3 | 3 ตัวบน ตรงตำแหน่ง |
| `3BOTTOM` | 3 | 3 ตัวล่าง (บางระบบไม่มี — เฉพาะหวยไทย) |
| `3TOD` | 3 | 3 ตัวโต๊ด สลับตำแหน่งได้ |
| `3FRONT` | 3 | 3 ตัวหน้า (เฉพาะหวยไทย) |
| `2TOP` | 2 | 2 ตัวบน |
| `2BOTTOM` | 2 | 2 ตัวล่าง |
| `4TOP` | 4 | 4 ตัวบน (เฉพาะหวยไทย) |
| `4TOD` | 4 | 4 ตัวโต๊ด (เฉพาะหวยไทย) |
| `RUN_TOP` | 1 | วิ่งบน — อยู่ใน 3 ตัวบนถือว่าถูก |
| `RUN_BOT` | 1 | วิ่งล่าง — อยู่ใน 2 ตัวล่างถือว่าถูก |

### Validation (ลำดับ)
1. `BetType.IsValid()` — ต้องเป็นหนึ่งใน 10 ประเภท
2. `LotteryType.IsValid()`
3. `ValidateNumber(number, betType)`:
   - regex `^\d+$` (ตัวเลขล้วน)
   - `len(number) == betType.DigitCount()`
   - **อนุญาตเลขนำศูนย์** (เช่น `"007"`, `"05"`)
4. `ValidateAmount(amount, minBet, maxBet)`:
   - `amount > 0`
   - `amount >= minBet`
   - `maxBet > 0 ? amount <= maxBet : ไม่จำกัด`
5. `numberban.Check()` — ถ้า `full_ban` → reject
6. `CheckBetLimit()` — `currentTotal + amount <= maxPerNumber`
7. `betting.ValidateYeekeeShoot(number)` — 5 หลัก (สำหรับ yeekee shoot เท่านั้น ไม่ใช่ bet)

### Rate / Payout
1. ใช้ `shopspring/decimal` ทุกจุดคำนวณเงิน — **ห้าม** ใช้ float arithmetic ตรง
2. `CalculatePayout(amount, rate) = amount × rate`
3. **Cap สูงสุด**: `MaxPayoutAmount = 100,000,000` (100 ล้าน) — ถ้าเกิน clip ที่ค่านี้
4. Default rates:
   - **หวยไทย full** — 3TOP=900, 3TOD=150, 4TOP=6000, 2TOP/BOTTOM=90, RUN_TOP=3.2, RUN_BOT=4.2 (`lottery/rules.go:45`)
   - **หวยมาตรฐาน** — 3TOP=850, 3TOD=120, 2TOP/BOTTOM=90, RUN_TOP=3.2, RUN_BOT=4.2 (`lottery/rules.go:59`)
5. เลขอั้น `reduce_rate` → ใช้ `ReducedRate` แทน rate ปกติ (`numberban.GetEffectiveRate()`)

## 🔄 Flow

```
POST /bets (request)
  → betting.Validate(req, minBet, maxBet)            // number + amount + types
  → numberban.Check(number, betType, bans)           // full_ban → reject 403
  → rate = numberban.GetEffectiveRate(check, default)
  → betting.CheckBetLimit(num, amt, current, max)    // over limit → reject 400
  → Redis: INCR bet_total:{round}:{type}:{num} += amt
  → DB: INSERT bets (status=pending, rate=rate)
  → return payout_preview = CalculatePayout(amt, rate)
```

## ⚠️ Edge Cases

- `maxBet = 0` หรือ `maxPerNumber = 0` → **ไม่จำกัด**
- 3TOD เลขซ้ำ (เช่น 884) — ไม่เป็น permutation ของ 847 → ไม่ถูก (sort แล้วเทียบ string)
- RUN_TOP เทียบด้วย `strings.Contains(Top3, number)` — digit ซ้ำใน Top3 นับครั้งเดียวก็พอ
- `Validate()` **ไม่เช็คเลขอั้น + limit** — API layer ต้องเรียกเอง (เพราะต้องดึง DB/Redis)

## 🔗 Source of Truth (file:line)

- BetType enum: `types/enums.go:119-143`
- `Validate()`: `betting/validator.go:103-125`
- `ValidateNumber()`: `betting/validator.go:40-58`
- `ValidateAmount()`: `betting/validator.go:72-87`
- `ValidateYeekeeShoot()`: `betting/validator.go:138-153`
- `CalculatePayout()`: `betting/calculator.go:49-56`
- `CheckBetLimit()`: `betting/limit.go:32-49`
- Number ban check: `numberban/checker.go:51-68`
- Per-operator ban filter: `numberban/checker.go:120-167`

## 📝 Change Log

- 2026-04-20: Initial — 10 bet types + decimal-safe rate calc
