package players

import (
	"math"
	"testing"
)

const floatApprox = 1e-4

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) <= floatApprox
}

func TestScale_g_and_e_symmetry(t *testing.T) {
	// If one player is more uncertain (higher Rating Deviation i.e. 200 below) than the other:
	// 	- `p` below is very uncertain (high Rating Deviation), `op` is very certain (low RD)
	// 	- p's expected chance to win might be 65%, but op's expected chance to win might only be 35%
	// They don't add up to exactly 100% (1 below) because the system is weighing the confidence of each player,
	// but neither player's confidence can be less than 0, or greater 1 individually
	p := From(1500, 200, 0.06, "p")
	op := From(1400, 30, 0.06, "o")

	sP := p.Scale()
	sOp := op.Scale()

	gOp := sOp.g()
	if gOp <= 0 || gOp > 1 {
		t.Fatalf("g(φ) out of bounds: %v", gOp)
	}

	eP := sP.e(sOp)
	if eP <= 0 || eP >= 1 {
		t.Fatalf("E out of bounds: %v", eP)
	}

	// Because p has higher μ than op, expected score e should be > 0.5
	if eP <= 0.5 {
		t.Fatalf("expected e > 0.5 for stronger player; got %v", eP)
	}

	eOp := sOp.e(sP)

	if eOp+eP <= 1 {
		t.Fatalf("expected asymmetry in expected scores when RDs differ: eP=%v, eOp=%v", eP, eOp)
	}

	// If both players had same RD (and thus same g), E(μ,μ_j)+E(μ_j,μ) ≈ 1
	// If both players are equally uncertain (same Rating Deviation) as seen below
	// 	- If A is stonger -> A expected to win 70%, B expected to win 30%
	// 	- The sum is ~100%. Makes sense.
	pSameRD := From(1500, 200, 0.06, "p2")
	opSameRD := From(1400, 200, 0.06, "o2")

	s1 := pSameRD.Scale()
	s2 := opSameRD.Scale()
	e1 := s1.e(s2)
	e2 := s2.e(s1)

	if !(e1+e2 > 0.9999 && e1+e2 < 1.0001) {
		t.Fatalf("E symmetry failed for equal R")
	}
}
