# Lottery Types — ประเภทหวย 39 ประเภท

> Last updated: 2026-04-20
> Related code: `types/enums.go:16`, `lottery/rules.go:72`

## 🎯 Purpose
นิยามประเภทหวยทั้งหมดที่ระบบรองรับ — ใช้เป็น enum กลาง (`LotteryType`) ที่ทุก service (member-api, admin-api, provider-game-api) import เพื่อให้ผลลัพธ์ การ validate และการคำนวณ rate ตรงกันข้าม service

## 📋 Rules (กฏเงื่อนไข)

1. **Enum เป็น string แบบตายตัว** — ห้ามเพิ่ม/เปลี่ยนค่าโดยไม่อัพเดท migration ของทุก API
2. **IsValid() ต้องคืน true เท่านั้น** ก่อนบันทึกลง DB — ถ้า invalid ให้ reject request
3. **จัดกลุ่ม 5 หมวด:**
   - หวยไทย (3): `THAI_GOV`, `BAAC`, `GSB`
   - ยี่กี (1): `YEEKEE` — ประเภทเดียวที่ `IsAutoResult() = true`
   - หวยลาว (5): `LAO_VIP`, `LAO_PATTANA`, `LAO_STAR`, `LAO_SAMAKKEE`, `LAO_THAKHEK_VIP`
   - หวยฮานอย (3): `HANOI`, `HANOI_VIP`, `HANOI_PATTANA`
   - มาเลย์ (1): `MALAY`
   - หวยหุ้น (26): ดู `types/enums.go:43-69`
4. **หวยไทย 3 ประเภท** ใช้ bet types ครบ 10 ชนิด (มี 3ตัวล่าง, 3ตัวหน้า, 4ตัว)
5. **หวยอื่นทั้งหมด** ใช้ bet types มาตรฐาน 6 ชนิด (3TOP, 3TOD, 2TOP, 2BOTTOM, RUN_TOP, RUN_BOT)
6. **IsAutoResult() = true เฉพาะ YEEKEE** — ระบบคำนวณเองจาก shoots, ประเภทอื่น admin กรอกผล

## 🔄 Flow

```
Request มาถึง API
  → parse lottery_type → LotteryType(string)
  → if !LotteryType.IsValid() → reject 400
  → lottery.GetRule(lt) → ได้ LotteryRule (AllowedBetTypes, DefaultRates)
  → ตรวจ bet + คำนวณ rate ต่อ
```

## ⚠️ Edge Cases

- ถ้าเพิ่มหวยใหม่ → ต้องเพิ่มทั้ง 3 จุด: `enums.go` const, `IsValid()` switch, `DefaultRules` map
- หวยหุ้น VIP vs non-VIP เป็น **คนละประเภทกัน** — rate และ schedule แยกจากกัน
- `LAO_THAKHEK_VIP` ใช้ค่า `LotteryTypeLaoThakhek` ใน Go (ชื่อ Go อาจไม่ตรง string value เป๊ะ)
- ห้ามใช้ string literal ตรงๆ — import `types.LotteryTypeXxx` เสมอเพื่อลด typo

## 🔗 Source of Truth (file:line)

- Enum values: `types/enums.go:18-70`
- `IsValid()`: `types/enums.go:73-105`
- `IsAutoResult()`: `types/enums.go:109-111`
- `DefaultRules` map: `lottery/rules.go:72-124`
- `GetRule()` / `IsBetTypeAllowed()` / `GetDefaultRate()`: `lottery/rules.go:135-174`
- DB seed: `lotto-standalone-member-api/migrations/010_lottery_types_restructure.sql`

## 📝 Change Log

- 2026-04-20: Initial — ครอบคลุม 39 ประเภท
