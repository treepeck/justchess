package models

import (
	"chess-api/enums"
	"chess-api/models/helpers"
	"chess-api/models/pieces"
	"time"

	"github.com/google/uuid"
)

type Game struct {
	Id       uuid.UUID          `json:"id"`
	Control  enums.Control      `json:"control"`
	Bonus    uint               `json:"bonus"` // 0 | 1 | 2 | 10
	Status   enums.Status       `json:"status"`
	WhiteId  uuid.UUID          `json:"whiteId"`
	BlackId  uuid.UUID          `json:"blackId"`
	PlayedAt time.Time          `json:"playedAt"`
	Moves    helpers.MovesStack `json:"moves"`
	// string is a Position type presented as string
	Pieces map[string]pieces.Piece `json:"pieces"`
}

type CreateGameDTO struct {
	Id      uuid.UUID     `json:"id"`
	Control enums.Control `json:"control"`
	Bonus   uint          `json:"bonus"`
	WhiteId uuid.UUID     `json:"whiteId"`
	BlackId uuid.UUID     `json:"blackId"`
}

func NewGame(id uuid.UUID, control enums.Control,
	bonus uint, whiteId, blackId uuid.UUID,
) *Game {
	g := &Game{
		Id:       id,
		Control:  control,
		Bonus:    bonus,
		Status:   enums.Waiting,
		WhiteId:  whiteId,
		BlackId:  blackId,
		PlayedAt: time.Now(),
		Moves:    *helpers.NewMovesStack(),
		Pieces:   make(map[string]pieces.Piece),
	}

	g.initPieces()
	return g
}

func (g *Game) initPieces() {
	g.initPawns()
	g.initRooks()
	g.initKnights()
	g.initBishops()
	g.initQueens()
	g.initKings()
}

func (g *Game) initPawns() {
	for i := 0; i < 8; i++ {
		pos := helpers.PosFromInd(1, i)
		g.Pieces[pos.String()] = pieces.NewPawn(enums.Black, pos)

		pos = helpers.PosFromInd(6, i)
		g.Pieces[pos.String()] = pieces.NewPawn(enums.White, pos)
	}
}

func (g *Game) initRooks() {
	positions := []int{0, 7}

	for i := 0; i < 2; i++ {
		pos := helpers.PosFromInd(0, positions[i])
		g.Pieces[pos.String()] = pieces.NewRook(enums.Black, pos)

		pos = helpers.PosFromInd(7, positions[i])
		g.Pieces[pos.String()] = pieces.NewRook(enums.White, pos)
	}
}

func (g *Game) initKnights() {
	positions := []int{1, 6}

	for i := 0; i < 2; i++ {
		pos := helpers.PosFromInd(0, positions[i])
		g.Pieces[pos.String()] = pieces.NewKnight(enums.Black, pos)

		pos = helpers.PosFromInd(7, positions[i])
		g.Pieces[pos.String()] = pieces.NewKnight(enums.White, pos)
	}
}

func (g *Game) initBishops() {
	positions := []int{2, 5}

	for i := 0; i < 2; i++ {
		pos := helpers.PosFromInd(0, positions[i])
		g.Pieces[pos.String()] = pieces.NewBishop(enums.Black, pos)

		pos = helpers.PosFromInd(7, positions[i])
		g.Pieces[pos.String()] = pieces.NewBishop(enums.White, pos)
	}
}

func (g *Game) initQueens() {
	pos := helpers.PosFromInd(0, 3)
	g.Pieces[pos.String()] = pieces.NewQueen(enums.Black, pos)

	pos = helpers.PosFromInd(7, 3)
	g.Pieces[pos.String()] = pieces.NewQueen(enums.White, pos)
}

func (g *Game) initKings() {
	pos := helpers.PosFromInd(0, 4)
	g.Pieces[pos.String()] = pieces.NewKing(enums.Black, pos)

	pos = helpers.PosFromInd(7, 4)
	g.Pieces[pos.String()] = pieces.NewKing(enums.White, pos)
}

func (g *Game) TakeMove(startPos, endPos helpers.Position) {
	// the user can only take a move if they previously select a square
	// if g.selectedSquare.Rank != 0 && g.selectedSquare.File != 0 {
	// 	for _, pos := range g.availibleMoves {
	// 		// check if the move is availible
	// 		if pos.File == endPos.File && pos.Rank == endPos.Rank {
	// 			// remove piece from previous position
	// 			piece := g.Pieces[g.selectedSquare.String()]
	// 			g.Pieces[g.selectedSquare.String()] = nil
	// 			// move piece to a new position
	// 			g.Pieces[endPos.String()] = piece
	// 		}
	// 	}
	// }
}
