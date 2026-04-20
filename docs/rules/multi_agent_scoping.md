# Multi-Agent Scoping — 1 เว็บ = 1 Root Node

> Last updated: 2026-04-20
> Related code: migrations `016_agent_downline.sql`, `018_per_node_settings.sql`, `020_absorb_agents_into_nodes.sql`

## 🎯 Purpose
ระบบเป็น **multi-tenant**: 1 database รองรับหลายเว็บหวย โดยใช้ `agent_node_id` เป็น scope key — ทุก query ต้อง filter ตาม node (หรือ descendant path) เพื่อกันข้อมูลรั่วข้ามเว็บ

## 📋 Rules (กฏเงื่อนไข)

1. **1 root node = 1 เว็บหวย** (`role='admin' AND parent_id IS NULL`)
2. **โครง tree**: `admin → share_holder → senior → master → agent → agent_downline (ซ้อนได้ไม่จำกัด)`
3. ทุก node เก็บ `path` (materialized path เช่น `/1/5/12/`) + `depth` → หาสายใต้ด้วย `path LIKE '/1/5/%'`
4. ทุก **transactional table** ต้องมี `agent_node_id`:
   - `members`, `bets`, `transactions`, `lottery_rounds`, `yeekee_rounds`, `yeekee_shoots`
   - `referral_commissions`, `agent_profit_transactions`
5. ทุก **settings table** ต้องมี `agent_node_id` (migration 018):
   - `pay_rates`, `agent_bank_accounts`, `promotions`, `member_levels`, `contact_channels`, `agent_banners`, `settings`, `affiliate_settings`, `share_templates`
6. Settings **NULL agent_node_id** = default ทั้งระบบ (admin ตั้ง); **per-node** override default
7. **Branding per root node** (migration 020): `code`, `domain`, `subdomain`, `logo_url`, `site_name`, theme colors, contact — เก็บใน `agent_nodes` row ของ root
8. Non-root inherit branding จาก root ผ่าน path (traverse ขึ้น)

## 🔄 Flow

```
Request มา (member / admin)
  → auth → ได้ member_id / node_id
  → resolve agent_node_id:
      member: members.agent_node_id
      node  : agent_nodes.id
  → resolve root_node_id: walk path [0] (first segment)
  → ทุก query:
      WHERE agent_node_id = ?  (direct)
      หรือ  WHERE agent_node_id IN (SELECT id FROM agent_nodes WHERE path LIKE '<my_path>%')
  → ห้าม query โดยไม่ใส่ filter เด็ดขาด
```

### Per-node setting resolution
```
1. SELECT ... WHERE agent_node_id = <current_node> LIMIT 1
2. ถ้าไม่เจอ → SELECT ... WHERE agent_node_id IS NULL LIMIT 1   (default)
3. ถ้าไม่เจอ → hard-coded default ใน code
```

## 🌍 Per-node vs Global

| ข้อมูล | Scope | หมายเหตุ |
|-------|-------|---------|
| `lottery_types` | **Global** | 39 ประเภท — share ทุก node |
| `bet_types` | **Global** | 10 ประเภท |
| `pay_rates` | **Per-node** | node ตั้ง rate เองได้ |
| `number_bans` | **Per-node** | migration 017 (bans per node) |
| `members` | **Per-node** | ลูกค้าของ node ใด node นั้น |
| `bets`, `transactions`, `rounds` | **Per-node** | agent_node_id บังคับ |
| `yeekee_rounds` | **Per-node** | 88 รอบ/node/วัน (แยกอิสระ) |
| `agent_bank_accounts`, `promotions`, `banners` | **Per-node** | migration 018 |
| `affiliate_settings` | **Per-node** | commission rate ต่อ node |
| branding (logo, theme, domain) | **Per-root-node** | migration 020 |

## ⚠️ Edge Cases

- ลืม filter `agent_node_id` → **ข้อมูลรั่ว** ข้ามเว็บ (security-critical)
- Path ผิดสลับ parent-child → downline profit ผิดทั้ง tree → ต้อง validate ตอน create/move node
- Yeekee รอบซ้ำข้าม agent? — **ไม่ซ้ำ** เพราะ unique key `(lottery_type_id, round_date, agent_node_id)`
- Non-root node (e.g. agent) login หลังบ้าน → เห็นแค่ descendants ของตัวเอง (path prefix)
- `agents` table ถูก **ลบทิ้งแล้ว** (migration 020) — ทุกที่ที่เคย `agent_id` → เปลี่ยนเป็น `agent_node_id` ของ root node
- Fallback cron: ถ้าไม่มี root node → hardcode `rootNodeIDs = [1]` (yeekee_cron.go:107)

## 🔗 Source of Truth (file:line)

- Tree schema: `migrations/016_agent_downline.sql`
- Bans per node: `migrations/017_bans_per_node.sql`
- Settings per node: `migrations/018_per_node_settings.sql`
- Lottery per node: `migrations/019_lottery_per_node.sql`
- Absorb agents → nodes: `migrations/020_absorb_agents_into_nodes.sql`
- Root node lookup (cron): `lotto-standalone-member-api/internal/job/yeekee_cron.go:100-109`
- Scope usage (settle): `lotto-standalone-member-api/internal/job/yeekee_cron.go:255-326`

## 📝 Change Log

- 2026-04-20: Initial — สรุป 1 root node = 1 เว็บ + tree scoping rules
