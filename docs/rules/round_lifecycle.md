# Round Lifecycle — สถานะของรอบหวย

> Last updated: 2026-04-20
> Related code: `types/enums.go:167`, `types/enums.go:219`, `lotto-standalone-member-api/internal/job/yeekee_cron.go:188`

## 🎯 Purpose
กำหนด state machine ของรอบหวย (lottery round) และรอบยี่กี (yeekee round) เพื่อให้ทุก service เปลี่ยนสถานะตามลำดับที่ถูกต้อง และกันไม่ให้ action ข้าม state (เช่น แทงตอน closed)

## 📋 Rules (กฏเงื่อนไข)

### LotteryRound.status (หวยทั่วไป)
1. Flow มาตรฐาน: `upcoming → open → closed → resulted`
2. เพิ่ม state พิเศษ: `missed` — server ดาวน์ระหว่างรอบ ต้องให้ admin กรอกผลเอง
3. แทงได้เฉพาะ `open` (`RoundStatus.CanBet() = true เฉพาะ open`)
4. Transition อัตโนมัติ (cron 30s):
   - `upcoming → open` เมื่อ `open_time <= now`
   - `open → closed` เมื่อ `close_time <= now`
   - `closed → resulted` เมื่อ admin กรอกผลแล้ว settle เสร็จ

### YeekeeRound.status (ยี่กีเท่านั้น)
1. Flow: `waiting → shooting → calculating → resulted`
2. เพิ่ม state พิเศษ: `missed` — รอบที่ server ปิดช่วง shooting
3. ยิงเลขได้เฉพาะ `shooting` (`YeekeeStatus.CanShoot()`)
4. Transition (cron 30s):
   - `waiting → shooting` เมื่อ `start_time <= now < end_time`
   - `waiting → missed` เมื่อ `end_time <= now` (ไม่เคยเข้า shooting)
   - `shooting → calculating` เมื่อ `end_time <= now` → trigger settle
   - `calculating → resulted` หลัง `CalculateResultWithSeed()` เสร็จ

### Bet.status
1. Flow: `pending → (won | lost | cancelled | refunded)`
2. `IsSettled()` = true สำหรับทุก state ยกเว้น `pending`
3. `MatchAll()` ข้าม bet ที่ `IsSettled()` แล้ว (กัน double settle)

## 🔄 Flow

```
LotteryRound:
  upcoming ──(open_time)──▶ open ──(close_time)──▶ closed ──(admin ออกผล)──▶ resulted
                            │                                                   ▲
                            └────(server down)────▶ missed ──(admin กรอก)──────┘

YeekeeRound:
  waiting ──(start_time)──▶ shooting ──(end_time)──▶ calculating ──(hash+settle)──▶ resulted
     │                                                                                 ▲
     └─────────(end_time before shooting)─────────▶ missed ──(admin ออกผลเอง)─────────┘
```

## ⚠️ Edge Cases

- ถ้า `shooting` แต่ไม่มี shoots → mark `resulted` พร้อม `total_shoots=0` (ไม่มีใครถูกรางวัล)
- รอบ `missed` ไม่ trigger payout อัตโนมัติ — admin ต้องกรอกผลผ่าน admin-api
- `closed → resulted` ต้องอยู่ใน DB transaction เดียวกับ bet updates + wallet credits
- ห้ามเปลี่ยนสถานะข้าม (เช่น `upcoming → resulted`) — ต้องผ่าน state กลางเสมอ

## 🔗 Source of Truth (file:line)

- `RoundStatus` enum + `CanBet()`: `types/enums.go:167-189`
- `YeekeeStatus` enum + `CanShoot()`: `types/enums.go:219-231`
- `BetStatus` enum + `IsSettled()`: `types/enums.go:197-211`
- Transition logic: `lotto-standalone-member-api/internal/job/yeekee_cron.go:188-218` (open/close/missed)
- Settle trigger: `lotto-standalone-member-api/internal/job/yeekee_cron.go:243-326`

## 📝 Change Log

- 2026-04-20: Initial — รวม `missed` state ที่เพิ่มมาสำหรับกรณี server down
