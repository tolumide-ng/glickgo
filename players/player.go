package players

import (
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
