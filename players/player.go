package players

import (
	"fmt"
	"math"

	"github.com/tolumide-ng/glickgo"
)

type PlayerArray [3]float64
type Player struct {
	rating float64
	// What does this really mean? It refers to how uncertain we are about a player's true skill
	// High Rating Deviation ->  We don't know much about this player
	// Low Rating Deviation -> We are confident about this player's true skill
	ratingDeviation float64
	volatility      float64
	PlayerID        string
}

// `From` creates a Player from explicit fields
func From(rating, ratingDeviation, volatility float64, playerID string) Player {
	return Player{rating, ratingDeviation, volatility, playerID}
}

// Initialize a Glicko2 player (Creates a plauer initialized to defaults)
func New(PlayerID string) Player {
	return Player{
		rating:          glickgo.DefaultRating,
		ratingDeviation: glickgo.DefaultRatingDeviation,
		volatility:      glickgo.DefaultVolatility,
		PlayerID:        PlayerID,
	}
}

// `ToPlayer` converts PlayerAray -> Player
func (p PlayerArray) ToPlayer(playerID string) Player {
	return Player{p[0], p[1], p[2], playerID}
}

// `ToPlayerArray` converts Player -> PlayerArray
func (p *Player) ToPlayerArray() PlayerArray {
	return PlayerArray{p.rating, p.ratingDeviation, p.volatility}
}

// Convert the ratings and Rating Deviation of a player into the Glicko2 scale
// `Scale` converts a Player into Glicko-2 scaled values (μ, φ)
func (p *Player) Scale() scale {
	return scale{
		miu: ((p.rating - glickgo.DefaultRating) / glickgo.DefaultScalingFactor),
		phi: p.ratingDeviation / glickgo.DefaultScalingFactor,
	}
}

func (p *Player) getV(opponents []Player) float64 {
	s := p.Scale()

	return s.v(opponents)
}

// Δ = v * Σ_j g(φ_j) * (s_j - E(μ, μ_j))
// where:
//
//	v   = variance (getV)
//	g   = g(φ_j) = 1 / sqrt(1 + 3φ_j²/π²)
//	s_j = actual score vs opponent j (1 = win, 0.5 = draw, 0 = loss)
//	E   = expected score vs opponent j = 1 / (1 + exp(-g(φ_j)(μ - μ_j)))
//
// deltaAndSum computes Delta Δ and the internal sum Σ_j g(φ_j)(s_j - E(μ, μ_j))
// The outcome values are always from the perspective of p (the receiver of this method)
func (p *Player) deltaAndSum(opponents map[Player]glickgo.Result) (float64, float64) {
	oppList := make([]Player, 0, len(opponents))
	for p := range opponents {
		oppList = append(oppList, p)
	}

	v := p.getV(oppList)
	meScale := p.Scale()
	sum := 0.0

	for opp, outcome := range opponents {
		oppScale := opp.Scale()
		score := outcome.Value()

		sum += oppScale.g() * (score - meScale.e(oppScale))
	}

	return v * sum, sum
}

// getF implements the function f(x) from Glickman's paper.
func (p *Player) getF(delta, v, x float64) float64 {
	a := math.Log(p.volatility * p.volatility)
	eX := math.Exp(x)

	phi2 := p.Scale().phi * p.Scale().phi

	// numerator = e^x (Δ² - φ² - v - e^x)
	num := eX * ((delta * delta) - (phi2) - v - eX) // numerator

	// preDenominator (φ² + v + e^x)
	preDen := ((phi2) + v + eX)
	// denominator = 2 (φ² + v + e^x)²
	den := 2 * (preDen * preDen) // denominator

	return (num / den) - ((x - a) / (glickgo.DefaultTau * glickgo.DefaultTau))
}

// Decays a player's rating Deviation if they they don't play for s certain rating period.
// Returns the Updated player's value with a decayed struct
func (p Player) DecayDeviation() Player {
	phi := p.Scale().phi
	phiStar := math.Hypot(phi, p.volatility)

	return Player{
		rating:          p.rating,
		ratingDeviation: math.Min(phiStar*glickgo.DefaultScalingFactor, glickgo.DefaultRatingDeviation), // clamp at 350
		volatility:      p.volatility,
		PlayerID:        p.PlayerID,
	}

}

// newVolatility computes the new volatility σ' using the Illinois algorithm
func (p *Player) newVolatility(delta, v float64) float64 {
	// Solve for x
	delta2 := delta * delta
	phi2 := p.Scale().phi * p.Scale().phi

	a := math.Log(p.volatility * p.volatility)

	// Set the initial values of the iterative algorithm
	A := a
	B := 0.0

	if (delta2) > ((phi2) + v) {
		// Case: Δ² > φ² + v
		B = math.Log(delta2 - phi2 - v)
	} else {
		// Case: Δ² ≤ φ² + v
		// Iterate until f(a - kτ) < 0
		k := 1.0
		for p.getF(delta, v, (a-k*glickgo.DefaultTau)) < 0 {
			k += 1
		}

		B = a - k*glickgo.DefaultTau
	}

	fA, fB := p.getF(delta, v, A), p.getF(delta, v, B)

	iterations := 0
	// --- Step 4: Illinois algorithm iteration ---
	for math.Abs(B-A) > glickgo.DefaultConvergenceTolerance {
		iterations += 1

		// (a) False position step
		C := A + (A-B)*fA/(fB-fA)
		fC := p.getF(delta, v, C)

		// (b) Bracket update
		if fC*fB <= 0 {
			A, fA = B, fB
		} else {
			fA /= 2
		}

		if iterations >= glickgo.MaxIterations {
			panic(fmt.Sprintf("Convergence fail at %d iterations", iterations))
		}

		// (c) Move bracket edge
		B, fB = C, fC
	}
	return math.Exp(A / 2)
}

// Update returns a new Player with updated rating, RD, and volatility based on a set of results
// The outcome values are always from the perspective of p (the receiver of this method)
func (p *Player) Update(result map[Player]glickgo.Result) Player {

	if len(result) == 0 {
		return p.DecayDeviation()
	}

	opponents := make([]Player, 0, len(result))
	for opp, _ := range result {
		opponents = append(opponents, opp)
	}

	meScale := p.Scale()

	delta, sum := p.deltaAndSum(result)
	v := p.getV(opponents)

	// σ′
	newVolatility := p.newVolatility(delta, v)

	// Step 6: pre-rating period RD φ* = √(φ² + σ′²)
	phiStar := math.Sqrt((meScale.phi * meScale.phi) + (newVolatility * newVolatility))

	// Step 7: Update the rating and RD to the new values (new RD φ₀ and rating μ₀)
	newRatingDeviation := 1 / (math.Sqrt((1 / (phiStar * phiStar)) + (1 / v)))
	newMiu := meScale.miu + (newRatingDeviation*newRatingDeviation)*sum

	return Player{
		rating:          (newMiu * glickgo.DefaultScalingFactor) + glickgo.DefaultRating,
		ratingDeviation: newRatingDeviation * glickgo.DefaultScalingFactor,
		PlayerID:        p.PlayerID,
		volatility:      newVolatility,
	}
}

// PlayMatch updates two players given a match outcome. Outcome.PlayerID should be set to the winner's ID for wins, empty for draws.
func PlayMatch(players [2]Player, outcome glickgo.Result) [2]Player {
	result := [2]Player{}

	for index, p := range players {
		opponentIndex := 1 - index

		// Determine verdict from this player's perspective
		verdict := outcome

		if verdict.Value() != glickgo.DrawScore { // if it was a draw, we don't care about updating it
			// if it was a win, and the provided ID on the outcome isn't the same as this player, then we assume that this player lost the match
			if outcome.PlayerID != p.PlayerID {
				verdict = glickgo.Result{Outcome: (glickgo.Outcome)(glickgo.LossScore)}
			}
		}

		// verdict
		opponent := map[Player]glickgo.Result{players[opponentIndex]: verdict} //

		newRatings := p.Update(opponent)
		result[index] = newRatings
	}

	return result
}
