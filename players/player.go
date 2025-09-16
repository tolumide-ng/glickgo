package players

import (
	"math"

	"github.com/tolumide-ng/glickgo"
)

const (
	initialRating          = 1500
	initialRatingDeviation = 350.0
	initialVolatility      = 0.06
)

type PlayerArray [3]float64

type Player struct {
	rating          float64
	ratingDeviation float64
	volatilty       float64
	PlayerID        string
}

func From(rating float64, ratingDeviation float64, volatility float64, playerID string) Player {
	return Player{rating, ratingDeviation, volatility, playerID}
}

// Initialize a Glicko2 player
func New(PlayerID string) Player {
	return Player{
		rating:          initialRating,
		ratingDeviation: initialRatingDeviation,
		volatilty:       initialVolatility,
		PlayerID:        PlayerID,
	}
}

// Convert PlayerAray -> Player
func (p PlayerArray) ToPlayer(playerID string) Player {
	return Player{p[0], p[1], p[2], playerID}
}

// Convert Player -> PlayerArray
func (p Player) ToPlayerArray() PlayerArray {
	return PlayerArray{p.rating, p.ratingDeviation, p.volatilty}
}

// Convert the ratings and RD's onto the Glicko2 scale
// Convert Player to scaled values (μ, φ) for Glicko-2 math
func (p Player) Scale() Scale {
	miu := ((p.rating - initialRating) / glickgo.ScalingFactor)
	phi := p.volatilty / glickgo.ScalingFactor

	return Scale{miu, phi}
}

func (p Player) GetV(opponents []Player) float64 {
	return p.Scale().v(opponents)
}

// Δ = v * Σ_j g(φ_j) * (s_j - E(μ, μ_j))
// where:
//
//	v   = variance (GetV)
//	g   = g(φ_j) = 1 / sqrt(1 + 3φ_j²/π²)
//	s_j = actual score vs opponent j (1 = win, 0.5 = draw, 0 = loss)
//	E   = expected score vs opponent j = 1 / (1 + exp(-g(φ_j)(μ - μ_j)))
func (p Player) Delta(opponents map[Player]glickgo.Outcome) float64 {
	oppList := make([]Player, 0, len(opponents))
	for p := range opponents {
		oppList = append(oppList, p)
	}

	v := p.GetV(oppList)
	meScale := p.Scale()
	sum := 0.0

	for opp, outcome := range opponents {
		oppScale := opp.Scale()

		score, err := glickgo.GameResult{Outcome: outcome, PlayerID: opp.PlayerID}.Value()

		if err != nil {
			continue
		}

		sum += oppScale.g() * (score - meScale.e(oppScale))
	}

	return v * sum
}

func (p Player) getF(delta, v, x float64) float64 {
	a := p.volatilty * p.volatilty
	eX := math.Exp(x)

	// numerator = e^x (Δ² - φ² - v - e^x)
	num := eX * ((delta * delta) - (p.Scale().phi * p.Scale().phi) - v - eX) // numerator

	// preDenominator (φ² + v + e^x)
	preDen := ((p.Scale().phi * p.Scale().phi) + v + eX)
	// denominator = 2 (φ² + v + e^x)²
	den := 2 * preDen * preDen // denominator

	return (num / den) - ((x - a) / glickgo.Tau * glickgo.Tau)
}

func (p Player) newVolatility(delta, v float64) float64 {

	// Solve for x
	delta2 := delta * delta
	phi2 := p.Scale().phi * p.Scale().phi
	a := p.volatilty * p.volatilty

	// Set the initial values of the iterative algorithm
	A := p.volatilty * p.volatilty
	B := 0.0

	if (delta2) > ((phi2) + v) {
		// Case: Δ² > φ² + v
		B = math.Log(delta2 - phi2 - v)
	} else {
		// Case: Δ² ≤ φ² + v
		// Iterate until f(a - kτ) < 0
		k := 1.0
		for p.getF(delta, v, (a-k*glickgo.Tau)) < 0 {
			k += 1
		}

		B = a - k*glickgo.Tau
	}

	fA := p.getF(delta, v, A)
	fB := p.getF(delta, v, B)

	// --- Step 4: Illinois algorithm iteration ---
	for math.Abs(B-A) > glickgo.ConvergenceTolerance {
		// (a) False position step
		C := A + (A-B)*fA/(fB-fA)
		fC := p.getF(delta, v, C)

		// (b) Bracket update
		if fC*fB <= 0 {
			A = B
			fA = fB
		} else {
			fA /= 2
		}

		// (c) Move bracket edge
		B = C
		fB = fC
	}
	return math.Exp(A / 2)
}

func (p Player) Update(opponents map[Player]glickgo.Outcome) {

	oppList := make([]Player, 0, len(opponents))
	for opp, _ := range opponents {
		oppList = append(oppList, opp)
	}

	meScale := p.Scale()

	delta := p.Delta(opponents)
	v := p.GetV(oppList)

	// σ′
	newVolatility := p.newVolatility(delta, v)

	// Step 6: pre-rating period RD φ* = √(φ² + σ′²)
	phiStar := math.Sqrt((meScale.phi * meScale.phi) + (newVolatility * newVolatility))
}
