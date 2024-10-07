package helpers

import "chess-api/models/enums"

type Move struct {
	To          Pos            `json:"to"`
	From        Pos            `json:"from"`
	MoveType    enums.MoveType `json:"moveType"`
	IsCheck     bool           `json:"isCheck"`
	IsCheckmate bool           `json:"isCheckmate"`
	IsCapture   bool           `json:"isCapture"`
	SecondsLeft uint           `json:"secondsLeft"`
	// determines the selected piece that the pawn will be promoted to
	PromotionPayload enums.Piece `json:"promotionPayload"`
}

type MoveDTO struct {
	To               Pos         `json:"to"`
	From             Pos         `json:"from"`
	PromotionPayload enums.Piece `json:"promotionPayload"`
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
