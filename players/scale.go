package players

import "math"

type Scale struct {
	miu float64
	phi float64
}

func (self Scale) g() float64 {
	return (1 / (math.Sqrt(1 + 3*(self.phi*self.phi)/(math.Pi*math.Pi))))
}

func (self Scale) e(opponent Scale) float64 {
	return 1 / (1 + math.Exp((-opponent.g() * (self.miu - opponent.miu))))
}
