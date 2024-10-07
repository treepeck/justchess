package models

import (
	"chess-api/models/enums"
	"chess-api/models/helpers"
	"chess-api/models/pieces"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Game struct {
	Id       uuid.UUID                    `json:"id"`
	Control  enums.Control                `json:"control"`
	Bonus    uint                         `json:"bonus"` // 0 | 1 | 2 | 10
	Status   enums.Status                 `json:"status"`
	WhiteId  uuid.UUID                    `json:"whiteId"`
	BlackId  uuid.UUID                    `json:"blackId"`
	PlayedAt time.Time                    `json:"playedAt"`
	Moves    helpers.MovesStack           `json:"moves"`
	Pieces   map[helpers.Pos]pieces.Piece `json:"pieces"`
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
		Pieces:   make(map[helpers.Pos]pieces.Piece),
	}

	g.initPieces()
	return g
}

func (g *Game) MarshalJSON() ([]byte, error) {
	gameDTO := struct {
		Id       uuid.UUID               `json:"id"`
		Control  enums.Control           `json:"control"`
		Bonus    uint                    `json:"bonus"` // 0 | 1 | 2 | 10
		Status   enums.Status            `json:"status"`
		WhiteId  uuid.UUID               `json:"whiteId"`
		BlackId  uuid.UUID               `json:"blackId"`
		PlayedAt time.Time               `json:"playedAt"`
		Moves    helpers.MovesStack      `json:"moves"`
		Pieces   map[string]pieces.Piece `json:"pieces"`
	}{
		Id:       g.Id,
		Control:  g.Control,
		Bonus:    g.Bonus,
		Status:   g.Status,
		WhiteId:  g.WhiteId,
		BlackId:  g.BlackId,
		PlayedAt: g.PlayedAt,
		Moves:    g.Moves,
		Pieces:   make(map[string]pieces.Piece),
	}

	for pos, piece := range g.Pieces {
		gameDTO.Pieces[pos.String()] = piece
	}

	return json.Marshal(gameDTO)
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
	for i := 1; i <= 8; i++ {
		pos := helpers.PosFromInd(1, i)
		g.Pieces[pos] = pieces.NewPawn(enums.Black, pos)

		pos = helpers.PosFromInd(6, i)
		g.Pieces[pos] = pieces.NewPawn(enums.White, pos)
	}
}

func (g *Game) initRooks() {
	positions := []int{1, 8}

	for i := 0; i < 2; i++ {
		pos := helpers.PosFromInd(0, positions[i])
		g.Pieces[pos] = pieces.NewRook(enums.Black, pos)

		pos = helpers.PosFromInd(7, positions[i])
		g.Pieces[pos] = pieces.NewRook(enums.White, pos)
	}
}

func (g *Game) initKnights() {
	positions := []int{2, 7}

	for i := 0; i < 2; i++ {
		pos := helpers.PosFromInd(0, positions[i])
		g.Pieces[pos] = pieces.NewKnight(enums.Black, pos)

		pos = helpers.PosFromInd(7, positions[i])
		g.Pieces[pos] = pieces.NewKnight(enums.White, pos)
	}
}

func (g *Game) initBishops() {
	positions := []int{3, 6}

	for i := 0; i < 2; i++ {
		pos := helpers.PosFromInd(0, positions[i])
		g.Pieces[pos] = pieces.NewBishop(enums.Black, pos)

		pos = helpers.PosFromInd(7, positions[i])
		g.Pieces[pos] = pieces.NewBishop(enums.White, pos)
	}
}

func (g *Game) initQueens() {
	pos := helpers.PosFromInd(0, 4)
	g.Pieces[pos] = pieces.NewQueen(enums.Black, pos)

	pos = helpers.PosFromInd(7, 4)
	g.Pieces[pos] = pieces.NewQueen(enums.White, pos)
}

func (g *Game) initKings() {
	pos := helpers.PosFromInd(0, 5)
	g.Pieces[pos] = pieces.NewKing(enums.Black, pos)

	pos = helpers.PosFromInd(7, 5)
	g.Pieces[pos] = pieces.NewKing(enums.White, pos)
}

func (g *Game) TakeMove(md helpers.MoveDTO) bool {
	piece := g.Pieces[md.From]
	if piece != nil {
		move := &helpers.Move{
			To:               md.To,
			From:             md.From,
			PromotionPayload: md.PromotionPayload,
		}

		if piece.Move(g.Pieces, move) {
			// figure out which pawns can be captured en passant
			g.checkEnPassantPawns(move)
			// check is the move was a check
			g.isCheckMove(move)

			g.Moves.Push(*move)
			return true
		}
	}
	return false
}

func (g *Game) checkEnPassantPawns(lastMove *helpers.Move) {
	for pos, piece := range g.Pieces {
		if piece.GetName() == enums.Pawn {
			piece.(*pieces.Pawn).IsEnPassant = false
			if pos == lastMove.To {
				if (pos.Rank == 4 || pos.Rank == 5) &&
					piece.(*pieces.Pawn).MovesCounter == 1 {
					piece.(*pieces.Pawn).IsEnPassant = true
				}
			}
		}
	}
}

func (g *Game) isCheckMove(lastMove *helpers.Move) {
	piece := g.Pieces[lastMove.To]
	possilbeMoves := piece.GetPossibleMoves(g.Pieces)

	for pos, mt := range possilbeMoves {
		if mt != enums.Defend {
			if g.Pieces[pos] != nil && g.Pieces[pos].GetName() == enums.King {
				// check if the lastMove was a checkmate
				lastMove.IsCheckmate = g.isCheckmate(pos)
				lastMove.IsCheck = true
			}
		}
	}
}

func (g *Game) isCheckmate(pos helpers.Pos) bool {
	if g.Pieces[pos] != nil && g.Pieces[pos].GetName() == enums.King {
		if len(g.Pieces[pos].GetPossibleMoves(g.Pieces)) == 0 {
			return true
		}
	}
	return false
}
