package game

import (
	"chess-api/models/game/enums"
	"chess-api/models/game/helpers"
	"chess-api/models/game/pieces"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

// G represents a game and stores all necessary data.
type G struct {
	Id                uuid.UUID                     `json:"-"`
	Bonus             uint                          `json:"bonus"` // time increment.
	Moves             []helpers.Move                `json:"-"`
	Pieces            map[helpers.Pos]pieces.Piece  `json:"-"`
	Status            enums.Status                  `json:"status"`
	Control           enums.Control                 `json:"control"`
	Result            enums.GameResult              `json:"-"`
	White             *helpers.Player               `json:"white"`
	Black             *helpers.Player               `json:"black"`
	PlayedAt          time.Time                     `json:"-"`
	CurrentTurn       enums.Color                   `json:"-"`
	CurrentValidMoves map[helpers.PossibleMove]bool `json:"-"`
}

// NewG creates a new game.
func NewG(id uuid.UUID, control enums.Control,
	bonus uint, whiteId, blackId uuid.UUID,
) *G {
	g := &G{
		Id:                id,
		Bonus:             bonus,
		Moves:             make([]helpers.Move, 0),
		Pieces:            make(map[helpers.Pos]pieces.Piece),
		Status:            enums.Waiting,
		Control:           control,
		White:             helpers.NewPlayer(uuid.Nil, control.ToDuration()),
		Black:             helpers.NewPlayer(uuid.Nil, control.ToDuration()),
		PlayedAt:          time.Now(),
		CurrentTurn:       enums.White,
		CurrentValidMoves: make(map[helpers.PossibleMove]bool),
	}
	g.initPieces()
	g.CurrentValidMoves = g.getValidMoves(enums.White)
	return g
}

// getValidMoves finds all valid moves for the specified player.
// The validity of the returned moves is guaranteed - player`s possible
// moves are filtered down by the validity checking.
// For more details about move validity, see [chess-api/models/pieces.GetPossibleMoves].
func (g *G) getValidMoves(side enums.Color) map[helpers.PossibleMove]bool {
	ppm := g.getPossibleMoves(side)
	// determine enemy side
	es := enums.White
	if side == enums.White {
		es = enums.Black
	}
	// store valid moves
	vm := make(map[helpers.PossibleMove]bool)
	for pm := range ppm { // iterate over all possible moves
		// skip defend moves
		if pm.MoveType == enums.Defend {
			continue
		}
		// make sure that after making this move the allied king is not checked
		isChecked := false
		// store previous piece on end position
		prevPiece := g.Pieces[pm.To]
		// execute possible mvoe
		g.Pieces[pm.To] = g.Pieces[pm.From]
		g.Pieces[pm.From].Move(pm.To)
		delete(g.Pieces, pm.From)
		// find all enemy possible moves on a new position
		eppm := g.getPossibleMoves(es)
		for epm := range eppm {
			p := g.Pieces[epm.To]
			// if the allied king is checked, move is not valid
			if p != nil && p.GetType() == enums.King && p.GetColor() == side {
				isChecked = true
				break
			}
		}
		// restore the board
		g.Pieces[pm.From] = g.Pieces[pm.To]
		delete(g.Pieces, pm.To)
		if prevPiece != nil {
			g.Pieces[pm.To] = prevPiece
		}
		g.Pieces[pm.From].Move(pm.From)
		g.Pieces[pm.From].SetMovesCounter(g.Pieces[pm.From].GetMovesCounter() - 2)
		// if the allied king remained in a safe position, the move is valid
		if !isChecked {
			vm[pm] = true
		}
	}
	return vm
}

// getPossibleMoves finds all possible moves for the specified player
// on the specified board.
// The validity of the returned moves is not guaranteed.
// For more details about move validity, see [chess-api/models/pieces.GetPossibleMoves].
func (g *G) getPossibleMoves(side enums.Color,
) map[helpers.PossibleMove]bool {
	pm := make(map[helpers.PossibleMove]bool)
	for from, piece := range g.Pieces {
		if piece != nil && piece.GetColor() == side {
			ppm := piece.GetPossibleMoves(g.Pieces)
			for pos, mt := range ppm {
				pm[helpers.PossibleMove{
					To:       pos,
					From:     from,
					MoveType: mt,
				}] = true
			}
		}
	}
	return pm
}

// HandleMove handles player`s moves. True is retured if the move is valid,
// false otherwise.
func (g *G) HandleMove(m *helpers.Move) bool {
	for vm := range g.CurrentValidMoves {
		if vm.From.IsEqual(m.From) && vm.To.IsEqual(m.To) {
			// stop the player`s ticker.
			if g.CurrentTurn == enums.White {
				g.White.Ticker.Stop()
			} else {
				g.Black.Ticker.Stop()
			}

			m.MoveType = vm.MoveType
			// determine is move a capture
			if g.Pieces[m.To] != nil {
				m.IsCapture = true
			}
			// iterate over board, find all pawns and reset the en passant flag,
			// since the en passant capture is only availible for one move
			for _, p := range g.Pieces {
				switch p.GetType() {
				case enums.Pawn:
					p.(*pieces.Pawn).IsEnPassant = false
				}
			}
			g.movePiece(m.From, m.To)
			// handle special moves
			switch m.MoveType {
			case enums.Promotion:
				g.handlePromotion(m)

			case enums.EnPassant:
				m.IsCapture = true
				g.handleEnPassant(g.Pieces[m.To])

			case enums.LongCastling:
				g.handleCastling(m.To.Rank, m.To.File-2)

			case enums.ShortCastling:
				g.handleCastling(m.To.Rank, m.To.File+1)
			}
			g.processLastMove(m)
			return true
		}
	}
	return false
}

func (g *G) StartGame(whiteId, blackId uuid.UUID) {
	g.White.Id = whiteId
	g.Black.Id = blackId

	g.Status = enums.Continues
	// start white ticker
	g.White.Ticker.Reset(time.Second)
}

// endGame ends the game due to provided reason.
func (g *G) endGame(r enums.GameResult) {
	g.Result = r
	g.Status = enums.Over
	g.White.Ticker.Stop()
	g.Black.Ticker.Stop()
}

// handlePromotion promotes a pawn to a specified piece.
func (g *G) handlePromotion(m *helpers.Move) {
	if m.PromotionPayload == 0 { // invalid piece for promotion
		m.PromotionPayload = enums.Queen
	}

	pp := g.Pieces[m.To] // previous piece
	g.Pieces[m.To] = pieces.BuildPiece(m.PromotionPayload, pp.GetColor(),
		pp.GetPosition(), pp.GetMovesCounter(),
	)
}

// handleCastling moves the rook according to the type of castling.
func (g *G) handleCastling(rank, file int) {
	var rookPos helpers.Pos
	rookPos.Rank = rank
	rookPos.File = file

	dF := -2             // delta file between rook and moved king - 0-0 by default
	if file == enums.A { // 0-0-0
		dF = 3
	}
	g.movePiece(rookPos, helpers.NewPos(rookPos.File+dF, rank))
}

// handleEnPassant removes captured pawn from the board.
func (g *G) handleEnPassant(lmp pieces.Piece) {
	fn := slog.String("func", "handleEnPassant")
	// determine pawn direction
	dir := 1
	if lmp.GetColor() == enums.Black {
		dir = -1
	}
	// en passant pawn is located behind the moved pawn
	pos := helpers.NewPos(
		lmp.GetPosition().File,
		lmp.GetPosition().Rank-dir,
	)
	capturedPawn := g.Pieces[pos]
	if capturedPawn == nil {
		slog.Warn("error in logic, incorrect EnPassant", fn)
		return
	}
	delete(g.Pieces, pos)
}

func (g *G) processLastMove(lastMove *helpers.Move) {
	g.determineCheck(lastMove)
	// get valid moves for a next player
	g.CurrentValidMoves = g.getValidMoves(g.CurrentTurn.GetOppositeColor())
	// if the player does not have any valid moves,
	// the pevious move is either a stalemate or a checkmate
	if len(g.CurrentValidMoves) == 0 {
		if lastMove.IsCheck {
			lastMove.IsCheckmate = true
			g.endGame(enums.Checkmate)
		} else {
			g.endGame(enums.Stalemate)
		}
	}
	// store the move
	g.Moves = append(g.Moves, *lastMove)
	// switch the turn and reset the next player`s ticker
	g.CurrentTurn = g.CurrentTurn.GetOppositeColor()
	if g.CurrentTurn == enums.White {
		g.White.Ticker.Reset(time.Second)
		lastMove.TimeLeft = g.Black.Time
	} else {
		g.Black.Ticker.Reset(time.Second)
		lastMove.TimeLeft = g.White.Time
	}
}

// determineCheck determines whether the previous move was a check.
func (g *G) determineCheck(lastMove *helpers.Move) {
	lmp := g.Pieces[lastMove.To] // last moved piece
	// get possible moves for the last moved piece
	pm := lmp.GetPossibleMoves(g.Pieces)

	for pos := range pm {
		p := g.Pieces[pos]
		// check is enemy king
		if p != nil && p.GetType() == enums.King &&
			p.GetColor() != lmp.GetColor() {
			lastMove.IsCheck = true
		}
	}
}

// movePiece moves a piece and updates pieces.
func (g *G) movePiece(from, to helpers.Pos) {
	g.Pieces[from].Move(to)
	// update the board state
	g.Pieces[to] = g.Pieces[from]
	delete(g.Pieces, from)
}

// initPieces places the pieces in their standard places.
func (g *G) initPieces() {
	g.initPawns()
	g.initRooks()
	g.initKnights()
	g.initBishops()
	g.initQueens()
	g.initKings()
}

// initPawns is a helper function to initialize pawns.
func (g *G) initPawns() {
	for i := 1; i <= 8; i++ {
		pos := helpers.PosFromInd(1, i)
		g.Pieces[pos] = pieces.NewPawn(enums.Black, pos)

		pos = helpers.PosFromInd(6, i)
		g.Pieces[pos] = pieces.NewPawn(enums.White, pos)
	}
}

// initRook is a helper function to initialize rooks.
func (g *G) initRooks() {
	positions := []int{1, 8}

	for i := 0; i < 2; i++ {
		pos := helpers.PosFromInd(0, positions[i])
		g.Pieces[pos] = pieces.NewRook(enums.Black, pos)

		pos = helpers.PosFromInd(7, positions[i])
		g.Pieces[pos] = pieces.NewRook(enums.White, pos)
	}
}

// initKnights is a helper function to initialize knights.
func (g *G) initKnights() {
	positions := []int{2, 7}

	for i := 0; i < 2; i++ {
		pos := helpers.PosFromInd(0, positions[i])
		g.Pieces[pos] = pieces.NewKnight(enums.Black, pos)

		pos = helpers.PosFromInd(7, positions[i])
		g.Pieces[pos] = pieces.NewKnight(enums.White, pos)
	}
}

// initBishops is a helper function to initialize bishops.
func (g *G) initBishops() {
	positions := []int{3, 6}

	for i := 0; i < 2; i++ {
		pos := helpers.PosFromInd(0, positions[i])
		g.Pieces[pos] = pieces.NewBishop(enums.Black, pos)

		pos = helpers.PosFromInd(7, positions[i])
		g.Pieces[pos] = pieces.NewBishop(enums.White, pos)
	}
}

// initQueens is a helper function to initialize queens.
func (g *G) initQueens() {
	pos := helpers.PosFromInd(0, 4)
	g.Pieces[pos] = pieces.NewQueen(enums.Black, pos)

	pos = helpers.PosFromInd(7, 4)
	g.Pieces[pos] = pieces.NewQueen(enums.White, pos)
}

// initQueens is a helper function to initialize kings.
func (g *G) initKings() {
	pos := helpers.PosFromInd(0, 5)
	g.Pieces[pos] = pieces.NewKing(enums.Black, pos)

	pos = helpers.PosFromInd(7, 5)
	g.Pieces[pos] = pieces.NewKing(enums.White, pos)
}
