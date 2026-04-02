// Package yeekee จัดการ logic ยี่กี — ออกผลอัตโนมัติจากเลขที่สมาชิกยิงมา
//
// ยี่กีทำงานอย่างไร:
//  1. ระบบสร้างรอบอัตโนมัติทุก 15 นาที (88 รอบ/วัน) — cron job ใน API layer
//  2. รอบเปิดรับยิงเลข (shooting phase) — สมาชิกส่งเลข 5 หลัก ผ่าน WebSocket
//  3. หมดเวลา → ระบบ SHA256(server_seed + shoots) → mod ให้เหลือ 5 หลัก → ตัด 2-3 ตัวท้ายเป็นผล
//  4. เทียบผลกับ bets → จ่ายเงิน
//
// ⚠️ SECURITY: Hash Commitment Scheme ป้องกัน result manipulation
//
// ปัญหาเดิม:
//   - ผล = SUM(shoots) % 100000 → คนยิงคนสุดท้ายสามารถคุมผลได้ 100%
//   - เพราะรู้ผลรวมสะสม real-time → คำนวณเลขที่ต้องยิงได้ทันที
//
// วิธีแก้ (Hash Commitment):
//   1. สร้างรอบ → สุ่ม server_seed (32 bytes) + เก็บ seed_hash = SHA256(server_seed)
//   2. seed_hash แสดงให้ลูกค้าเห็น (พิสูจน์ว่าระบบไม่โกง — commit ก่อนเปิดยิง)
//   3. คำนวณผล → SHA256(server_seed + sorted_shoots_concat) → ตัดเป็นตัวเลข → mod 100000
//   4. หลังออกผล → เปิดเผย server_seed → ลูกค้า verify ได้ว่าผลตรงกับ commitment
//
// ทำไมโกงไม่ได้:
//   - ลูกค้าไม่รู้ server_seed (เห็นแค่ hash)
//   - SHA256 เป็น one-way function → รู้ hash ก็คำนวณ seed ย้อนกลับไม่ได้
//   - แม้รู้เลขยิงทั้งหมด → ไม่สามารถคำนวณว่าต้องยิงเลขอะไรเพื่อให้ผลออกตามต้องการ
//
// ความสัมพันธ์:
// - ใช้ types.YeekeeShoot, types.RoundResult
// - ถูกเรียกโดย:
//   - standalone-member-api (#3): cron job หมดเวลายี่กี → CalculateResultWithSeed()
//   - provider-game-api (#7): cron job หมดเวลายี่กี → CalculateResultWithSeed()
// - ผลที่ได้ส่งต่อไปยัง payout.MatchAll() เพื่อเทียบ bets
package yeekee

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/farritpcz/lotto-core/types"
)

// =============================================================================
// Server Seed — สร้างและ verify server seed สำหรับ Hash Commitment
// =============================================================================

// GenerateServerSeed สุ่ม server seed 32 bytes (64 hex characters)
//
// เรียกตอนสร้างรอบยี่กี → เก็บ seed ไว้ secret จนกว่ารอบจะจบ
// คืน: seed (hex string), seedHash (SHA256 hex string)
func GenerateServerSeed() (seed string, seedHash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("failed to generate random seed: %w", err)
	}
	seed = hex.EncodeToString(b)
	seedHash = HashSeed(seed)
	return seed, seedHash, nil
}

// HashSeed คำนวณ SHA256 hash ของ seed (สำหรับ commitment)
func HashSeed(seed string) string {
	h := sha256.Sum256([]byte(seed))
	return hex.EncodeToString(h[:])
}

// VerifySeed ตรวจสอบว่า seed ตรงกับ hash ที่ commit ไว้
//
// ลูกค้าเรียกหลังออกผล → verify ว่าระบบไม่ได้เปลี่ยน seed หลังเห็นเลขยิง
func VerifySeed(seed, expectedHash string) bool {
	return HashSeed(seed) == expectedHash
}

// =============================================================================
// Algorithm — คำนวณผลยี่กี (Hash Commitment Scheme)
// =============================================================================

// CalculateResultWithSeed คำนวณผลยี่กีแบบป้องกันการโกง
//
// อัลกอริทึม:
//  1. เรียงเลขยิงตาม ID (ลำดับที่ยิง) → concat เป็น string เดียว
//  2. SHA256(server_seed + ":" + concat_shoots) → ได้ hash 32 bytes
//  3. ตัด 8 bytes แรก → แปลงเป็น uint64 → mod 100000 → ได้เลข 5 หลัก
//  4. ตัดเป็นผลรางวัล (Top3, Top2, Bottom2)
//
// Parameters:
//   - serverSeed: seed ที่สุ่มไว้ตอนสร้างรอบ (hex string)
//   - shoots: เลขที่สมาชิกยิงมาทั้งหมด (เรียงตาม ID)
//
// Returns:
//   - resultNumber: เลข 5 หลักผลลัพธ์
//   - roundResult:  ผลรางวัลที่ตัดแล้ว
//   - error
func CalculateResultWithSeed(serverSeed string, shoots []types.YeekeeShoot) (resultNumber string, roundResult types.RoundResult, err error) {
	if len(shoots) == 0 {
		return "", types.RoundResult{}, types.ErrNoShoots
	}
	if serverSeed == "" {
		return "", types.RoundResult{}, fmt.Errorf("server seed is empty")
	}

	// Step 1: เรียง shoots ตาม ID (ลำดับที่ยิง) → concat numbers
	sorted := make([]types.YeekeeShoot, len(shoots))
	copy(sorted, shoots)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ID < sorted[j].ID })

	var numbers []string
	for _, s := range sorted {
		numbers = append(numbers, s.Number)
	}
	shootsConcat := strings.Join(numbers, ",")

	// Step 2: SHA256(server_seed + ":" + shoots_concat)
	input := serverSeed + ":" + shootsConcat
	hash := sha256.Sum256([]byte(input))

	// Step 3: ตัด 8 bytes แรก → uint64 → mod 100000
	num := binary.BigEndian.Uint64(hash[:8])
	result := num % 100000

	// Step 4: format เป็น 5 หลัก
	resultNumber = fmt.Sprintf("%05d", result)

	// Step 5: ตัดเป็นผลรางวัล
	roundResult = ExtractResult(resultNumber)

	return resultNumber, roundResult, nil
}

// CalculateResult คำนวณผลยี่กีแบบเก่า (DEPRECATED — ใช้ CalculateResultWithSeed แทน)
//
// ⚠️ INSECURE: ผลคำนวณจาก SUM เลขยิง → คนยิงสุดท้ายโกงได้
// คงไว้เพื่อ backward compatibility เท่านั้น — ห้ามใช้ในรอบใหม่
func CalculateResult(shoots []types.YeekeeShoot) (resultNumber string, roundResult types.RoundResult, err error) {
	if len(shoots) == 0 {
		return "", types.RoundResult{}, types.ErrNoShoots
	}

	var totalSum int64
	for _, shoot := range shoots {
		num, parseErr := strconv.ParseInt(shoot.Number, 10, 64)
		if parseErr != nil {
			continue
		}
		totalSum += num
	}

	result := totalSum % 100000
	resultNumber = fmt.Sprintf("%05d", result)
	roundResult = ExtractResult(resultNumber)

	return resultNumber, roundResult, nil
}

// ExtractResult ตัดเลข 5 หลักเป็นผลรางวัล
//
// จาก 5 หลัก เช่น "83456":
//   - 3 ตัวบน (Top3)   = 3 ตัวท้าย    = "456" (index 2-4)
//   - 2 ตัวบน (Top2)   = 2 ตัวท้าย    = "56"  (index 3-4)
//   - 2 ตัวล่าง (Bottom2) = หลักที่ 2-3 = "34"  (index 1-2)
func ExtractResult(fiveDigit string) types.RoundResult {
	if len(fiveDigit) != 5 {
		return types.RoundResult{}
	}

	return types.RoundResult{
		Top3:    fiveDigit[2:5],
		Top2:    fiveDigit[3:5],
		Bottom2: fiveDigit[1:3],
	}
}

// GetShootCount นับจำนวนเลขที่ยิงในรอบ
func GetShootCount(shoots []types.YeekeeShoot) int {
	return len(shoots)
}

// GetShootSum คำนวณผลรวมของเลขที่ยิง (สำหรับแสดงผลใน UI)
//
// ⚠️ หลังเปลี่ยนเป็น Hash Commitment: ผลรวมนี้ใช้แสดง UI เท่านั้น
// ไม่ได้ใช้คำนวณผลจริงอีกแล้ว (ผลจริงใช้ SHA256)
func GetShootSum(shoots []types.YeekeeShoot) int64 {
	var sum int64
	for _, shoot := range shoots {
		num, err := strconv.ParseInt(shoot.Number, 10, 64)
		if err != nil {
			continue
		}
		sum += num
	}
	return sum
}
