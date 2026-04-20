# Yeekee Algorithm — Hash Commitment + Auto-Shoot Bot

> Last updated: 2026-04-20
> Related code: `yeekee/algorithm.go`, `lotto-standalone-member-api/internal/job/yeekee_bot.go`

## 🎯 Purpose
คำนวณผลยี่กีแบบ **ป้องกันการโกง** โดยใช้ SHA-256 hash commitment scheme + bot ยิงเลขช่วงวินาทีสุดท้าย เพื่อกันผู้เล่นคุมผลด้วยการยิงคนสุดท้าย

## 📋 Rules (กฏเงื่อนไข)

### Hash Commitment Scheme
1. **สร้างรอบ** → สุ่ม `server_seed` (32 bytes, `crypto/rand`) + เก็บ `seed_hash = SHA256(seed)`
2. **โชว์ seed_hash** ให้ client ตั้งแต่ต้นรอบ — commitment ผูกล่วงหน้า
3. **คำนวณผล** (ตอน `end_time`):
   - Sort shoots ตาม `ID` (ลำดับยิง)
   - `shootsConcat = number1,number2,...,numberN`
   - `hash = SHA256(server_seed + ":" + shootsConcat)`
   - `result5 = BigEndian.Uint64(hash[:8]) % 100000` → format `%05d`
4. **เปิดเผย seed** หลัง resulted — client เรียก `VerifySeed(seed, hash)` ได้เอง
5. **Legacy** `CalculateResult()` (sum-based) **DEPRECATED** — ใช้เฉพาะ fallback กรณี seed หาย

### Extract Result (จาก 5 หลัก)
| รางวัล | Index | ตัวอย่าง "83456" |
|--------|-------|-----------------|
| `Top3` | `[2:5]` | "456" |
| `Top2` | `[3:5]` | "56" |
| `Bottom2` | `[1:3]` | "34" |

### Auto-Shoot Bot
1. Bot member ต่อ agent: `username = "_system_bot_{agentNodeID}"`, `password = "bot_no_login"` (login ไม่ได้)
2. Tick ทุก 30s ต่อรอบ `shooting`:
   - **ปกติ**: 70% chance ยิง 1 เลขต่อ tick
   - **Last Second Protection**: `secondsRemaining < 10` → ยิง **100%** (เสมอ)
3. `yeekee_shoots.is_bot = 1` — admin filter ดูเฉพาะเลขจริงได้
4. Bot shoot ใช้ `rand.Intn(100000)` → pad 5 หลัก

## 🔄 Flow

```
สร้างรอบ:
  GenerateServerSeed() → (seed, seed_hash)
  INSERT yeekee_rounds (server_seed=seed, seed_hash=hash)
  ส่ง seed_hash ผ่าน WebSocket ให้ทุก client

ระหว่าง shooting:
  member ยิง → INSERT yeekee_shoots (is_bot=0)
  bot cron (30s): 70% random OR <10s → ยิงเลขสุ่ม is_bot=1

ตอน end_time:
  CalculateResultWithSeed(seed, shoots)
    → sort by ID
    → SHA256(seed + ":" + shoots)
    → uint64(hash[:8]) % 100000
    → ExtractResult(5digit) → (Top3, Top2, Bottom2)
  เปิดเผย seed ใน API response

Client verify:
  HashSeed(revealed_seed) == seed_hash (ที่ commit ไว้) → ยืนยันไม่โกง
```

## ⚠️ Edge Cases

- `len(shoots) == 0` → return `ErrNoShoots`, mark round `resulted` พร้อม `total_shoots=0` (ไม่มีใครถูก)
- `serverSeed == ""` → fallback ไป `CalculateResult()` (sum-based legacy) + log warning
- เลขยิงไม่ใช่ตัวเลข → `strconv.ParseInt` fail → skip ใน `GetShootSum` (ใช้แสดง UI เท่านั้น)
- ผู้เล่นรู้ hash + เลขยิงทั้งหมด → **คำนวณย้อน seed ไม่ได้** (SHA-256 one-way)
- ถ้า bot ถูก disable → Last Second Protection หาย → ผู้เล่นยิงท้ายโกงได้ — **ห้ามปิด bot**

## 🔗 Source of Truth (file:line)

- `GenerateServerSeed()`: `yeekee/algorithm.go:55-63`
- `HashSeed()` / `VerifySeed()`: `yeekee/algorithm.go:66-76`
- `CalculateResultWithSeed()`: `yeekee/algorithm.go:98-132`
- Legacy `CalculateResult()`: `yeekee/algorithm.go:138-157` (DEPRECATED)
- `ExtractResult()`: `yeekee/algorithm.go:165-175`
- Bot cron: `lotto-standalone-member-api/internal/job/yeekee_bot.go:75-100`
- `getOrCreateBotMember()`: `lotto-standalone-member-api/internal/job/yeekee_bot.go:38-68`
- Settlement integration: `lotto-standalone-member-api/internal/job/yeekee_cron.go:288`

## 📝 Change Log

- 2026-04-20: Initial — hash commitment + bot last-second protection
