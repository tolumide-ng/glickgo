package glickgo

// Glickgo is a glicko-2 library in golang

import "fmt"

var (
	ScalingFactor        = 173.7178
	ConvergenceTolerance = 0.000_001 // ε
	Tau                  = 0.5
)

func Greet(name string) string {
	return fmt.Sprintf("Hello %s! %f %f", name, ScalingFactor, ConvergenceTolerance)
}

type Glicko2 struct {
	// The system constant, τ, which constraints the change in volatility over time, needs to be set
	// prior to application of the system. Reasonable choices are between 0.3 and 1.2 (systems should be )
	Tau                  float64
	ConvergenceTolerance float64
}

func New() Glicko2 {
	return Glicko2{
		Tau:                  Tau,
		ConvergenceTolerance: ConvergenceTolerance,
	}
}
