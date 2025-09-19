package glickgo

type Outcome int

const (
	Loss Outcome = iota
	Draw
	Win
)

const (
	LossScore float64 = 0
	DrawScore float64 = 0.5
	WinScore  float64 = 1
)

type Result struct {
	Outcome Outcome
	// who the this outcome/Result(GameResult) is for:
	// i.e if PlayerA sees this as a win, then PlayerB must see it as a loss
	PlayerID string
}

func (o Outcome) Value() float64 {
	switch o {
	case Loss:
		return LossScore
	case Draw:
		return DrawScore
	case Win:
		return WinScore
	default:
		panic("unknown outcome")
	}
}
