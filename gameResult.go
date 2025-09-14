package glickgo

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
