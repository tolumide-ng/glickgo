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
}

// func (Player) From(rating float64, ratingDeviation float64, volatility float64) Player {
// 	Player{}
// }

// Initialize a Glicko2 player
func New() Player {
	return Player{
		rating:          initialRating,
		ratingDeviation: initialRatingDeviation,
		volatilty:       initialVolatility,
	}
}

// Convert PlayerAray -> Player
func (p PlayerArray) ToPlayer() Player {
	return Player{p[0], p[1], p[2]}
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
