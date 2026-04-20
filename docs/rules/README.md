# 📋 Rule Files — `lotto-core`

> **Source of truth** ของกฏ/เงื่อนไขแต่ละฟังก์ชันใน shared Go module
> ทุกครั้งที่แก้ logic ในไฟล์ที่ rule อ้างถึง → **ต้องอัพเดท rule ในคอมมิตเดียวกัน**

---

## 📚 Index — Rule Files ทั้งหมด

| Status | File | ครอบคลุม |
|--------|------|---------|
| ✅ | `lottery_types.md` | 39 ประเภทหวย, enum, ความหมาย |
| ✅ | `round_lifecycle.md` | Status flow (upcoming→open→closed→resulted+missed) |
| ✅ | `round_schedule.md` | ตารางเปิด/ปิดแต่ละประเภท (ยี่กี 88 รอบ, stock ตามตลาด) |
| ✅ | `betting_rules.md` | BetType, rate, min/max, bet number validation |
| ✅ | `yeekee_algorithm.md` | Hash commitment, server seed, auto-shoot bot |
| ✅ | `settlement_engine.md` | กฏตัดสินผล, payout calculation |
| ✅ | `commission_formulas.md` | สูตร profit sharing, diff%, walk up tree |
| ✅ | `multi_agent_scoping.md` | 1 เว็บ=1 node, ข้อมูล per-node vs global |

**Legend:** ✅ done · 🚧 partial · ⏳ not started

---

## ✍️ Template (ทุกไฟล์ต้องมีโครงนี้)

```markdown
# [ชื่อฟังก์ชัน]

> Last updated: YYYY-MM-DD
> Related code: `path/to/file.go:LINE`, `path/to/file2.go:LINE`

## 🎯 Purpose
[ฟังก์ชันนี้ทำอะไร ทำไมต้องมี — 1-3 บรรทัด]

## 📋 Rules (กฏเงื่อนไข)
1. เงื่อนไขข้อ 1
2. เงื่อนไขข้อ 2
   - sub-rule

## 🔄 Flow
[step-by-step หรือ ASCII diagram]

## ⚠️ Edge Cases
- ถ้า X → ทำ Y
- ห้าม Z

## 🔗 Source of Truth (file:line)
- Logic: `path/to/file.go:123`
- Types: `path/to/types.go:45`
- Tests: `path/to/file_test.go`

## 📝 Change Log
- YYYY-MM-DD: [สิ่งที่เปลี่ยน] (commit abc123)
```

---

## 🔒 Convention

1. **ภาษา:** ไทยเป็นหลัก, ศัพท์เทคนิค/โค้ด/enum ใช้ภาษาอังกฤษ
2. **ความยาว:** ไม่เกิน ~200 บรรทัดต่อไฟล์ — ถ้ายาวเกิน → split
3. **ห้ามลอก comment ในโค้ดมาวาง** — rule = "ทำไม + เงื่อนไข", โค้ด = "ยังไง"
4. **file:line ต้อง up-to-date** — ถ้าย้ายโค้ด ต้องอัพเดท reference
5. **Change log** — เขียนย่อๆ 1 บรรทัด + commit hash
6. **Cross-reference** — ลิงก์ rule files ใน repo อื่นได้ (ใช้ relative path ถ้าเป็นได้)
