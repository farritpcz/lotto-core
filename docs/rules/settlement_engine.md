# Settlement Engine — การตัดสินผล + จ่ายเงินรางวัล

> Last updated: 2026-04-20
> Related code: `payout/matcher.go`, `payout/settler.go`, `lotto-standalone-member-api/internal/job/yeekee_cron.go:333`

## 🎯 Purpose
เทียบ bets กับ `RoundResult` → ได้ผล won/lost ทุก bet, คำนวณเงินรางวัลรวม + กำไร/ขาดทุน → return ให้ API layer จัดการ DB transaction + wallet crediting

## 📋 Rules (กฏเงื่อนไข)

### Match Logic (ต่อ bet)

| BetType | ฟิลด์ใน RoundResult | วิธีเทียบ |
|---------|---------------------|-----------|
| `3TOP` | `Top3` | ตรงตำแหน่งเป๊ะ (`==`) |
| `3TOD` | `Top3` | `isPermutation()` (sort อักษรแล้วเทียบ) |
| `3FRONT` | `Front3` | ตรงตำแหน่ง (ต้องมี `Front3` ไม่ว่าง) |
| `3BOTTOM` | `Bottom3` | split ด้วย `,` → match **รางวัลใดรางวัลหนึ่ง** ก็ถือถูก |
| `4TOP` | `buildFullResult = Front3 + Top3` | ตัด 4 ตัวท้าย ตรงตำแหน่ง |
| `4TOD` | idem | ตัด 4 ตัวท้าย permutation |
| `2TOP` | `Top2` | ตรงตำแหน่ง |
| `2BOTTOM` | `Bottom2` | ตรงตำแหน่ง |
| `RUN_TOP` | `Top3` | `strings.Contains(Top3, number)` |
| `RUN_BOT` | `Bottom2` | `strings.Contains(Bottom2, number)` |

### Settlement Flow
1. **ข้าม bet ที่ `IsSettled()`** แล้ว (กัน double-pay)
2. ถ้า win → `WinAmount = betting.CalculatePayout(Amount, Rate)` (decimal-safe, cap 100M)
3. `SettleRound()` คืน `SettleRoundOutput` ที่มี:
   - `BetResults` — ทุก bet
   - `TotalWinners`, `TotalWinAmount`, `TotalLosers`
   - `TotalBetAmount` — รวมจาก bets ที่ยัง pending เท่านั้น
   - `Profit = TotalBetAmount - TotalWinAmount` (decimal subtraction)
4. **API layer ต้องทำ** (lotto-core ไม่แตะ DB):
   - UPDATE bets: `status`, `win_amount`, `settled_at`
   - UPDATE members: `balance += total_win` (per member, group ด้วย `GroupWinnersByMember()`)
   - INSERT transactions type=`win` (มี `agent_node_id`, `balance_before/after`, `reference`)
   - UPDATE lottery_rounds: `status=resulted`, `resulted_at`
   - เรียก `CalculateCommissions()` + `CalculateDownlineProfits()` async

### Grouping
- `GroupWinnersByMember(bets, results)` → `map[memberID] = totalWin` — จ่ายครั้งเดียวต่อ member
- `GroupWinnersByOperator(bets, results)` → `map[opID][memberID] = totalWin` — เฉพาะ provider mode (มี `OperatorID`), standalone ข้าม

## 🔄 Flow

```
Round resulted (admin กรอก OR yeekee auto):
  → fetch bets WHERE round=X AND status=pending
  → SettleRound({roundID, result, bets})
      → MatchAll(bets, result)                 // ทุก bet → BetResult
      → SummarizeResults()                     // decimal sum
      → คำนวณ profit = totalBet - totalWin
  → API layer:
      tx.Begin()
        UPDATE bets ...
        UPDATE members balance += win
        INSERT transactions (win, agent_node_id)
      tx.Commit()
      go CalculateCommissions(db, roundID, rootNodeID)
      go CalculateDownlineProfits(db, roundID, rootNodeID)
```

## ⚠️ Edge Cases

- **Empty pending bets** → log "no pending bets — skip payout", ไม่สร้าง transaction
- **3BOTTOM หลายรางวัล** → `Bottom3 = "123,456"`; `TrimSpace` + match ใด ๆ
- **4TOP ที่หวยไม่มี Front3** → `buildFullResult()` ใช้ `Top3` เดี่ยว (len 3) → 4TOP **ไม่มีวัน match** (ถูกต้อง)
- **Decimal precision** → ใช้ `decimal.Decimal` ทุกจุด sum/sub (มีกรณี 0.1+0.2 = 0.30000001 ที่ทำเงินหาย)
- **Panic ใน payout** → `tx.Rollback()` ใน recover → log + ไม่ crash cron
- Win transaction ต้อง set `agent_node_id = rootNodeID` จาก round (ห้าม hardcode 1)

## 🔗 Source of Truth (file:line)

- `Match()`: `payout/matcher.go:56-145`
- `buildFullResult()`: `payout/matcher.go:154-159`
- `MatchAll()`: `payout/matcher.go:173-183`
- `isPermutation()`: `payout/matcher.go:199-211`
- `SummarizeResults()`: `payout/matcher.go:221-234`
- `SettleRound()`: `payout/settler.go:67-96`
- `GroupWinnersByMember()`: `payout/settler.go:110-129`
- `GroupWinnersByOperator()`: `payout/settler.go:137-161`
- Integration (standalone yeekee): `lotto-standalone-member-api/internal/job/yeekee_cron.go:333-447`

## 📝 Change Log

- 2026-04-20: Initial — decimal-safe settle + 10 bet types match
