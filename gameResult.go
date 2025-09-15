package glickgo

import "errors"

type Outcome int

const (
	Loss Outcome = iota
	Draw
	Win
)

type GameResult struct {
	Outcome Outcome
	// who the this outcome/GameResult is for:
	// i.e if PlayerA sees this as a win, then PlayerB must see it as a loss
	PlayerID string
}

func (g GameResult) Value() (float64, error) {
	switch g.Outcome {
	case Loss:
		return 0, nil
	case Draw:
		return 0.5, nil
	case Win:
		return 1, nil
	default:
		return 0, errors.New("unknown outcome")
	}
}
