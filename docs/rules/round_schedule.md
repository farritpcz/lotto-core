# Round Schedule — ตารางเวลาเปิด/ปิดหวย

> Last updated: 2026-04-20 (updated for DB-driven schedule)
> Related code: `lottery/schedule.go`, `lotto-standalone-member-api/internal/job/yeekee_cron.go:91`, `lotto-standalone-admin-api/internal/job/round_lifecycle.go`

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
1. Schedule กำหนดใน DB (`lottery_types.schedule_config` JSON) — migration 025 ขึ้นไป
2. Admin กรอกผลเอง

### ⭐ Auto-Create (หวยที่ไม่ใช่ยี่กี) — admin-api cron
1. **Source of truth:** `lottery_types.schedule_config` JSON:
   ```json
   {"day_type": "daily|weekday|thai_gov", "open_time": "HH:MM", "close_time": "HH:MM"}
   ```
2. **Pre-create window: 30 วัน** (ทุก 1 ชม. cron ตรวจสร้างล่วงหน้า)
3. **Day types:**
   - `thai_gov` — วันที่ 1 และ 16 ของเดือน
   - `weekday` — จันทร์–ศุกร์
   - `daily` — ทุกวัน
4. **Round number format:** `YYYYMMDD` (ไม่มี session suffix — แยก AM/PM ด้วย lottery_type code)
5. **agent_node_id = NULL** (global — ทุก agent ใช้รอบเดียวกัน)
6. **close_time < open_time** = ปิดข้ามวัน (เช่น DJ 20:30 → 03:00 วันถัดไป)
7. **schedule_config = NULL** → ไม่ auto-create (ต้องสร้างผ่าน admin endpoint)

### Auto-Transition (หวยที่ไม่ใช่ยี่กี)
- `upcoming → open` เมื่อ `open_time <= NOW()` (cron 30 วิ, `BatchOpenRounds`)
- `open → closed` เมื่อ `close_time <= NOW()` (cron 30 วิ, `BatchCloseRounds`)

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
- 2026-04-20: Auto-create หวยทั้งหมด (ไม่ใช่แค่ยี่กี) — migration 025 ย้าย schedule → DB, window 7→30 วัน, admin-api cron อ่าน `lottery_types.schedule_config`
