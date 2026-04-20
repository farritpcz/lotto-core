# Commission & Downline Profit Formulas

> Last updated: 2026-04-20
> Related code: `lotto-standalone-member-api/internal/job/commission_job.go`, migration `016_agent_downline.sql`

## 🎯 Purpose
สูตรคำนวณ (1) **ค่าคอมแนะนำเพื่อน** (referral commission) และ (2) **กำไรสายงาน** (hierarchical profit sharing) — รันหลังรอบ settled เพื่อบันทึกลง `referral_commissions` + `agent_profit_transactions`

## 📋 Rules (กฏเงื่อนไข)

### A) Referral Commission (`CalculateCommissions`)
1. ดึง `affiliate_settings` ของ `agent_node_id = rootNodeID AND status='active'`
2. หา commission rate ตามลำดับ:
   - rate เฉพาะ lottery type (row ที่ `lottery_type_id = round.LotteryTypeID`)
   - fallback → default rate (row ที่ `lottery_type_id IS NULL`)
3. ถ้า `rate <= 0` → **skip** ทั้ง round
4. สำหรับทุก bet ที่ settled (`won` OR `lost`):
   - หา `members.referred_by` ของ bettor — ถ้า NULL → skip
   - `commission_amount = bet.amount × rate / 100` (คำนวณจากยอดแทง ไม่ใช่กำไร)
   - Dedupe: เช็ค `(bet_id, referrer_id)` ซ้ำก่อน insert
   - INSERT `referral_commissions` status=`pending`

### B) Downline Profit Sharing (`CalculateDownlineProfits`)
1. โครงสร้าง role: `admin (100%) → share_holder (95%) → senior (94%) → master (93%) → agent (92%) → agent_downline (91%)` (เป็นค่า default, edit ได้ผ่าน admin)
2. ทุก node **ต้องมี** `share_percent < parent.share_percent` เสมอ
3. `net_result` ต่อ bet:
   - `lost` → `net_result = +bet.amount` (เราได้กำไร)
   - `won`  → `net_result = bet.amount - win_amount` (ปกติติดลบ = ขาดทุน)
4. **Walk up tree** จาก node ที่ member สังกัด → root:
   - `my_percent = overrideMap[node] OR node.share_percent` (override ต่อ lottery type ก่อน)
   - `diff_percent = my_percent - child_percent` (leaf: child_percent=0)
   - `profit = round(net_result × diff_percent / 100, 2)`
   - INSERT `agent_profit_transactions` (round_id, bet_id, agent_node_id, my%, child%, diff%, profit)
   - เดินขึ้น: `child_percent = my_percent`, `current = parent`
5. Dedupe: `(bet_id, agent_node_id)` ซ้ำ → ข้าม แต่ walk up ต่อ
6. Override per lottery type: `agent_node_commission_settings.lottery_type = code`

### C) Formula Example (ลูกค้าเสีย 100 บาท)
```
agent_downline(91%): diff = 91 - 0  = 91  → profit = 91
agent(92%):          diff = 92 - 91 = 1   → profit = 1
master(93%):         diff = 93 - 92 = 1   → profit = 1
senior(94%):         diff = 94 - 93 = 1   → profit = 1
share_holder(95%):   diff = 95 - 94 = 1   → profit = 1
admin(100%):         diff = 100- 95 = 5   → profit = 5
                                       Σ = 100 ✓
```
ถ้าลูกค้าชนะ — `net_result` ติดลบ → ทุก node ขาดทุนตามสัดส่วน diff% ของ `net_result`

## 🔄 Flow

```
settleYeekeeRound() done
  → settleBets() commit
      → go CalculateCommissions(db, roundID, rootNodeID)
          → referral_commissions (status=pending, จ่ายตอน admin approve withdraw)
      → go CalculateDownlineProfits(db, roundID, rootNodeID)
          → agent_profit_transactions (ทุก node ในสาย)
```

## ⚠️ Edge Cases

- Member ไม่มี `agent_node_id` → ข้าม (ลูกค้าลอย, กำไรเข้า admin 100% จาก settle ตรง)
- Member ไม่มี `referred_by` → ไม่มี referral commission (แต่ profit sharing ยังทำ)
- Node ไม่มี parent (root) → loop break หลัง insert
- Override `share_percent` ต้องยัง `< parent` เสมอ — ตรวจใน admin UI (ไม่มีตรวจใน job)
- Rounding: ใช้ `math.Round(x*100)/100` (2 ตำแหน่ง) — ผลรวมทุก node อาจผิด ±0.01 จากทุกบิต
- ห้าม hardcode `rootNodeID = 1` — ต้องใช้ `yr.AgentNodeID` จากรอบ

## 🔗 Source of Truth (file:line)

- Referral: `lotto-standalone-member-api/internal/job/commission_job.go:33-153`
- Downline profit: `lotto-standalone-member-api/internal/job/commission_job.go:188-369`
- Tree schema: `lotto-standalone-member-api/migrations/016_agent_downline.sql`
- Override table: `agent_node_commission_settings` (migration 016)
- Trigger point: `lotto-standalone-member-api/internal/job/yeekee_cron.go:446`

## 📝 Change Log

- 2026-04-20: Initial — referral + downline profit formulas
