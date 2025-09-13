package glickgo

// Glickgo is a glicko-2 library in golang

import "fmt"

var (
	scalingFactor        = 173.7178
	convergenceTolerance = 0.000_001
)

func Greet(name string) string {
	return fmt.Sprintf("Hello %s! %f %f", name, scalingFactor, convergenceTolerance)
}
