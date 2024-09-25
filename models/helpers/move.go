package helpers

import "sync"

type Move struct {
	Index             uint   `json:"index"`
	SecondsLeft       uint   `json:"secondsLeft"`
	AlgebraicNotation string `json:"algebraicNotation"`
}

type MovesStack struct {
	Moves []Move `json:"moves"`
	sync.Mutex
}

func NewMovesStack() *MovesStack {
	return &MovesStack{
		Moves: make([]Move, 0),
	}
}

func (ms *MovesStack) Push(m Move) {
	ms.Lock()
	defer ms.Unlock()

	ms.Moves = append(ms.Moves, m)
}

func (ms *MovesStack) Pop() *Move {
	ms.Lock()
	defer ms.Unlock()

	if len(ms.Moves) == 0 {
		return nil
	}
	el := ms.Moves[len(ms.Moves)-1]
	ms.Moves = ms.Moves[:len(ms.Moves)-1]
	return &el
}
