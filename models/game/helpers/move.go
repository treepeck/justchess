package helpers

import (
	"chess-api/models/game/enums"
	"time"
)

// Move is used to store completed moves in a database.
type Move struct {
	To          Pos            `json:"to"`
	From        Pos            `json:"from"`
	IsCheck     bool           `json:"isCheck"`
	MoveType    enums.MoveType `json:"moveType"`
	TimeLeft    time.Duration  `json:"timeLeft"`
	IsCapture   bool           `json:"isCapture"`
	IsCheckmate bool           `json:"isCheckmate"`
	// determines the selected piece that the pawn will be promoted to
	PromotionPayload enums.PieceType `json:"pp"`
}

// possibleMove is a helper struct for determining player`s possible moves.
type PossibleMove struct {
	To       Pos            `json:"to"`
	From     Pos            `json:"from"`
	MoveType enums.MoveType `json:"moveType"`
}
