package players

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
		rating:          1500,
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
