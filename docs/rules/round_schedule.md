# Round Schedule — ตารางเวลาเปิด/ปิดหวย

> Last updated: 2026-04-20
> Related code: `lottery/schedule.go`, `lotto-standalone-member-api/internal/job/yeekee_cron.go:91`

## 🎯 Purpose
นิยามเวลาสร้าง/เปิด/ปิดของหวยแต่ละประเภท — cron job ใน API layer เรียก helper จาก `lottery/schedule.go` เพื่อสร้างรอบล่วงหน้า และ transition สถานะตามเวลา

## 📋 Rules (กฏเงื่อนไข)

### Yeekee
1. **88 รอบต่อวัน** ต่อ agent node (root) — ห่างกัน **15 นาที** (`DefaultYeekeeConfig`)
2. เริ่มรอบแรก **06:00** — รอบ 88 จบ **04:00 ของวันถัดไป** (28:00)
3. **Multi-agent**: ทุก root node (`role=admin, parent_id IS NULL`) ได้ 88 รอบแยกของตัวเอง
4. Round number format: `YYYYMMDD-NN` (NN = 01–88, 2 digits)
5. Shoot digits = **5 หลัก** (00000–99999)
6. **Auto-generate** ตอน cron เริ่มทำงาน + ทุกครั้งที่ข้ามวัน (midnight detection)
7. Uniqueness: เช็คด้วย `agent_node_id + DATE(start_time)` กันซ้ำ

### หวยไทย (THAI_GOV, BAAC, GSB)
1. ออกผล **วันที่ 1 และ 16** ของทุกเดือน — `GetThaiLotteryDates()` คืน 2 วัน/เดือน
2. Round number format: `YYYYMMDD` (ไม่มี suffix)
3. Admin กรอกผลเอง (ไม่ auto)

### หวยหุ้นไทย (STOCK_TH_PM)
1. 2 session/วัน (AM + PM) — `GetThaiStockTimes()`
   - AM: open **09:00**, close **12:00** (ผลมา ~12:30)
   - PM: open **13:00**, close **16:00** (ผลมา ~16:30)
2. Round number format: `YYYYMMDD-AM` / `YYYYMMDD-PM`
3. **ไม่เปิดวันเสาร์-อาทิตย์** → check ด้วย `IsWeekday(date)`

### หวยอื่น (ลาว, ฮานอย, มาเลย์, หุ้นต่างประเทศ 25 ประเภท)
1. Schedule กำหนดใน DB (`lottery_types` table) — ไม่ hard-code ใน lotto-core
2. Admin กรอกผลเอง

## 🔄 Flow

```
Midnight detection (cron 30s):
  today != lastDate?
    → createDailyRounds(db, now)
        → loop root nodes (role=admin, parent_id IS NULL)
            → GenerateYeekeeSchedule(date, DefaultYeekeeConfig)
            → สร้าง 88 lottery_round + 88 yeekee_round (agent_node_id set)
            → สร้าง server_seed + seed_hash ต่อรอบ
```

## ⚠️ Edge Cases

- ไม่มี root node → fallback เป็น `rootNodeIDs = [1]` (standalone default)
- Config ปัจจุบัน hard-code `DefaultYeekeeConfig` — TODO: อนาคตอ่าน per-agent config จาก DB
- หวยไทยวัน 16 ถ้าตรงกับวันหยุดราชการ → admin อาจเลื่อน (ระบบไม่รู้เอง)
- หวยหุ้นวันจันทร์ที่ตรงกับวันหยุดตลาด → admin ต้อง skip ไม่ออกผล

## 🔗 Source of Truth (file:line)

- Yeekee schedule: `lottery/schedule.go:55-72` (`GenerateYeekeeSchedule`)
- Current round lookup: `lottery/schedule.go:77-85` (`GetCurrentYeekeeRound`)
- Round number format: `lottery/schedule.go:103-118` (`GenerateRoundNumber`)
- หวยไทย dates: `lottery/schedule.go:130-136` (`GetThaiLotteryDates`)
- หวยหุ้นไทย times: `lottery/schedule.go:150-166` (`GetThaiStockTimes`)
- Weekday check: `lottery/schedule.go:170-173`
- Cron job: `lotto-standalone-member-api/internal/job/yeekee_cron.go:91-181`

## 📝 Change Log

- 2026-04-20: Initial — ครอบคลุม yeekee 88 รอบ + หวยไทย + หวยหุ้น TH
