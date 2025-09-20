package players

import (
	"math"
	"testing"

	"github.com/tolumide-ng/glickgo"
)

const floatApprox = 1e-2

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) <= floatApprox
}

func TestScalegAndeSymmetry(t *testing.T) {
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

func TestVAndDeltaAgainstGlikmanExample(t *testing.T) {
	me := From(1500, 200, 0.06, "me")
	op1 := From(1400, 30, 0.06, "o1")
	op2 := From(1550, 100, 0.06, "o2")
	op3 := From(1700, 300, 0.06, "o3")

	opponents := []Player{op1, op2, op3}

	v := me.getV(opponents)
	if !approxEqual(v, 1.7785) {
		t.Fatalf("v mismatch: want 1.7785 got %v", v)
	}

	// Build results map (scores from me's perspective)
	res := map[Player]glickgo.Result{
		op1: glickgo.NewResult(glickgo.Win, "me"),
		op2: glickgo.NewResult(glickgo.Loss, "o2"),
		op3: glickgo.NewResult(glickgo.Loss, "o3"),
	}

	delta, _ := me.deltaAndSum(res)
	if !approxEqual(delta, -0.4834) {
		t.Fatalf("delta mismatch: want -0.4834 got %v", delta)
	}
}

func TestNewVolatilityAndUpdateAgainstGlickmanExample(t *testing.T) {
	me := From(1500, 200, 0.06, "me")
	op1 := From(1400, 30, 0.06, "o1")
	op2 := From(1550, 100, 0.06, "o2")
	op3 := From(1700, 300, 0.06, "o3")

	res := map[Player]glickgo.Result{
		op1: glickgo.NewResult(glickgo.Win, "me"),
		op2: glickgo.NewResult(glickgo.Loss, "o2"),
		op3: glickgo.NewResult(glickgo.Loss, "o3"),
	}

	updated := me.Update(res)

	// Expected values from Glickman's worked example
	wantRating := 1464.06
	wantRD := 151.52
	wantSigma := 0.0599

	if !approxEqual(updated.rating, wantRating) {
		t.Fatalf("rating mismatch: want %v got %v", wantRating, updated.rating)
	}

	if !approxEqual(updated.ratingDeviation, wantRD) {
		t.Fatalf("RD mismatch: want %v got %v", wantRD, updated.ratingDeviation)
	}
	if !approxEqual(updated.volatility, wantSigma) {
		t.Fatalf("sigma mismatch: want %v got %v", wantSigma, updated.volatility)
	}

}

func TestPlayMatchSymmetryAndDirection(t *testing.T) {
	p1 := New("p1")
	p2 := New("p2")

	// make p1 stronger so their rating increases on win
	p1.rating = 1600
	p2.rating = 1400

	res := PlayMatch([2]Player{p1, p2}, glickgo.NewResult(glickgo.Win, "p1"))

	if res[0].rating <= p1.rating {
		t.Fatalf("winner did not increase rating: before %v after %v", p1.rating, res[0].rating)
	}

	if res[1].rating >= p2.rating {
		t.Fatalf("loser did not decrease rating: before %v after %v", p2.rating, res[1].rating)
	}
}

func TestRoundTripConversions(t *testing.T) {
	p := From(1723.4, 44.2, 0.123, "x")
	arr := p.ToPlayerArray()
	p2 := arr.ToPlayer("x2")

	if !approxEqual(p.rating, arr[0]) || !approxEqual(p.ratingDeviation, arr[1]) || !approxEqual(p.volatility, arr[2]) {
		t.Fatalf("ToPlayerArray mismatch")
	}
	if !approxEqual(p2.rating, arr[0]) {
		t.Fatalf("ToPlayer mismatch")
	}
}

func TestNoGamesPeriodOnlyRDIncrease(t *testing.T) {
	// If a player plays no games, only step 6 applies: φ0 = sqrt(φ^2 + σ^2)
	p := From(1500, 200, 0.06, "p")

	// Simulate no games by passing empty map
	e := map[Player]glickgo.Result{}
	updated := p.Update(e)

	// Expected phiStar = sqrt(phi^2 + sigma^2)
	phi := p.Scale().phi
	phiStar := math.Sqrt(math.Pow(phi, 2) + math.Pow(p.volatility, 2))
	wantRD := phiStar * glickgo.DefaultScalingFactor

	if !approxEqual(updated.ratingDeviation, wantRD) {
		t.Fatalf("RD without games mismatch: want %v got %v", wantRD, updated.ratingDeviation)
	}
}
