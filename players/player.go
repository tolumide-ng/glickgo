package players

import (
	"math"

	"github.com/tolumide-ng/glickgo"
)

const (
	initialRating = 1500
)

type PlayerArray [3]float64

type Player struct {
	rating          float64
	ratingDeviation float64
	volatilty       float64
}

// func (Player) New(rating float64, ratingDeviation float64, volatility float64) Player {
// 	Player{}
// }

// Initialize a Glicko2 player
func New() Player {
	return Player{
		rating:          initialRating,
		ratingDeviation: 350,
		volatilty:       0.06,
	}
}

func (p PlayerArray) FromPlayerVector() Player {
	return Player{p[0], p[1], p[2]}
}

func (p Player) ToPlayerArray() PlayerArray {
	return PlayerArray{p.rating, p.ratingDeviation, p.volatilty}
}

// Convert the ratings and RD's onto the Glicko2 scale
func (p Player) Scale() Scale {
	miu := ((p.rating - initialRating) / glickgo.ScalingFactor)
	phi := p.volatilty / glickgo.ScalingFactor

	scale := Scale{miu, phi}

	return scale
}

func (p Player) gValue(phi float64) float64 {
	return (1 / (math.Sqrt(1 + 3*(phi*phi)/(math.Pi*math.Pi))))
}
