package helpers

type Move struct {
	Index             uint   `json:"index"`
	SecondsLeft       uint   `json:"secondsLeft"`
	AlgebraicNotation string `json:"algebraicNotation"`
}

type MoveDTO struct {
	BeginPos      Pos  `json:"beginPos"`
	EndPos        Pos  `json:"endPos"`
	IsSpecialMove bool `json:"isSpecialMove"`
}

type MovesStack struct {
	Moves []Move `json:"moves"`
}

func NewMovesStack() *MovesStack {
	return &MovesStack{
		Moves: make([]Move, 0),
	}
}

func (ms *MovesStack) Push(m Move) {
	ms.Moves = append(ms.Moves, m)
}

func (ms *MovesStack) Pop() *Move {
	if len(ms.Moves) == 0 {
		return nil
	}
	el := ms.Moves[len(ms.Moves)-1]
	ms.Moves = ms.Moves[:len(ms.Moves)-1]
	return &el
}

func (ms *MovesStack) Depth() int {
	return len(ms.Moves)
}
