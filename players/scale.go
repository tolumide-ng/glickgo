package players

import "math"

type Scale struct {
	miu float64
	phi float64
}

// g(φ) = 1 / sqrt(1 + 3φ² / π²)
func (s Scale) g() float64 {
	return (1 / (math.Sqrt(1 + 3*(s.phi*s.phi)/(math.Pi*math.Pi))))
}

// E(μ, μ_j) = 1 / (1 + exp(-g(φ_j)(μ - μ_j)))
func (s Scale) e(opponent Scale) float64 {
	return 1 / (1 + math.Exp((-opponent.g() * (s.miu - opponent.miu))))
}

// v = [ Σ g(φ_j)² * E(μ, μ_j) * (1 - E(μ, μ_j)) ]⁻¹
func (me Scale) v(opponents []Player) float64 {
	sum := 0.0

	for _, opp := range opponents {
		oppScale := opp.Scale()
		g := oppScale.g()
		e := me.e(oppScale)
		sum += g * g * (1 - e)
	}

	return 1 / sum
}
