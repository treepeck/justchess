package game

import (
	"chess-api/models/game/enums"
	"chess-api/models/game/helpers"
	"chess-api/models/game/pieces"
	"time"

	"github.com/google/uuid"
)

// G represents a game and stores all necessary data.
type G struct {
	Id          uuid.UUID                     `json:"id"`        // game id.
	Bonus       uint                          `json:"bonus"`     // time increment.
	Moves       []helpers.Move                `json:"moves"`     // completed moves.
	Pieces      map[helpers.Pos]pieces.Piece  `json:"pieces"`    // board state.
	Status      enums.Status                  `json:"status"`    // game status.
	Control     enums.Control                 `json:"control"`   // time control.
	WhiteId     uuid.UUID                     `json:"whiteId"`   // white player id.
	BlackId     uuid.UUID                     `json:"blackId"`   // black player id.
	PlayedAt    time.Time                     `json:"playedAt"`  // when the game was started.
	WhiteTime   time.Duration                 `json:"whiteTime"` // how much time do the white have left.
	BlackTime   time.Duration                 `json:"blackTime"` // how much time do the black have left.
	Cvm         map[helpers.PossibleMove]bool // current valid moves.
	Epm         map[helpers.PossibleMove]bool // enemy possible moves.
	currentTurn enums.Color
	PlayerTurn  uuid.UUID
}

// NewG creates a new game.
func NewG(id uuid.UUID, control enums.Control,
	bonus uint, whiteId, blackId uuid.UUID,
) *G {
	g := &G{
		Id:          id,
		Bonus:       bonus,
		Moves:       make([]helpers.Move, 0),
		Pieces:      make(map[helpers.Pos]pieces.Piece),
		Status:      enums.Waiting,
		Control:     control,
		WhiteId:     whiteId,
		BlackId:     blackId,
		PlayedAt:    time.Now(),
		WhiteTime:   control.ToDuration(),
		BlackTime:   control.ToDuration(),
		Cvm:         make(map[helpers.PossibleMove]bool),
		Epm:         make(map[helpers.PossibleMove]bool),
		currentTurn: enums.White,
		PlayerTurn:  whiteId,
	}
	g.initPieces()
	// get white player valid moves in advance to increase performance
	g.Cvm = g.getValidMoves()
	g.Epm = g.getPossibleMoves(enums.Black)
	return g
}

// getValidMoves finds all valid moves for the specified player.
// The validity of the returned moves is guaranteed - player`s possible
// moves are filtered down by the validity checking.
// For more details about move validity, see [chess-api/models/pieces.GetPossibleMoves].
func (g *G) getValidMoves() map[helpers.PossibleMove]bool {
	ppm := g.getPossibleMoves(g.currentTurn)
	// determine enemy side
	es := enums.White
	if g.currentTurn == enums.White {
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
		prevPiece := g.Pieces[pm.To]
		g.Pieces[pm.To] = g.Pieces[pm.From]
		g.Pieces[pm.From].Move(pm.To)
		delete(g.Pieces, pm.From)
		// find all enemy possible moves on a new position
		eppm := g.getPossibleMoves(es)
		for epm := range eppm {
			p := g.Pieces[epm.To]
			// if the allied king is checked
			if p != nil && p.GetType() == enums.King && p.GetColor() == g.currentTurn {
				isChecked = true
				break
			}
		}
		// respore the board
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
func (g *G) HandleMove(m helpers.Move) bool {
	for vm := range g.Cvm {
		if vm.From.IsEqual(m.From) &&
			vm.To.IsEqual(m.To) {
			m.MoveType = vm.MoveType
			// determine is move a capture
			if g.Pieces[m.To] != nil {
				m.IsCapture = true
			}
			g.movePiece(m.From, m.To)
			// handle special moves
			switch m.MoveType {
			case enums.Promotion:
				g.handlePromotion(&m)

			case enums.LongCastling:
				g.handleCastling(m.To.Rank, m.To.File-2)

			case enums.ShortCastling:
				g.handleCastling(m.To.Rank, m.To.File+1)
			}

			g.processLastMove(&m)
			return true
		}
	}
	return false
}

// handlePromotion promotes a pawn to a specified piece.
func (g *G) handlePromotion(m *helpers.Move) {
	if m.PromotionPayload == 0 { // invalid piece for promotion
		return
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

	dF := -2             // delta file between rook and moved king- 0-0 by default
	if file == enums.A { // 0-0-0
		dF = 3
	}
	g.movePiece(rookPos, helpers.NewPos(rookPos.File+dF, rank))
}

func (g *G) processLastMove(lastMove *helpers.Move) {
	// get opponent`s valid moves to determine checkmate
	es := enums.White // enemy side
	g.PlayerTurn = g.WhiteId
	if g.currentTurn == enums.White {
		es = enums.Black
		g.PlayerTurn = g.BlackId
	}
	g.Epm = g.getPossibleMoves(g.currentTurn)
	g.currentTurn = es
	g.Cvm = g.getValidMoves()

	g.determineCheck(lastMove)
	// if the opponent does not have any valid moves,
	// it is either a stalemate or a checkmate
	if len(g.Cvm) == 0 {
		if lastMove.IsCheck {
			lastMove.IsCheckmate = true
		}
		// } else {
		// 	// TODO: handle stalemate game.endGame
		// }
	}

	// iterate over board, find all pieces and reset the en passant flag,
	// since the en passant capture is only availible for one move
	for _, p := range g.Pieces {
		switch p.GetType() {
		case enums.Pawn:
			p.(*pieces.Pawn).IsEnPassant = false
		}
	}
	g.determineEnPassant(*lastMove)

	// store the move
	g.Moves = append(g.Moves, *lastMove)
}

// determineCheck determines whether the previous move was a check.
func (g *G) determineCheck(lastMove *helpers.Move) {
	lmp := g.Pieces[lastMove.To] // last moved piece
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

// determineEnPassant determines whether the previous move was
// a pawn double squares forward.
func (g *G) determineEnPassant(lastMove helpers.Move) {
	lmp := g.Pieces[lastMove.To] // last moved piece

	if lmp.GetType() == enums.Pawn {
		if (lmp.GetPosition().Rank == 4 || lmp.GetPosition().Rank == 5) &&
			lmp.(*pieces.Pawn).MovesCounter == 1 {
			lmp.(*pieces.Pawn).IsEnPassant = true
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

func (g *G) initKings() {
	pos := helpers.PosFromInd(0, 5)
	g.Pieces[pos] = pieces.NewKing(enums.Black, pos)

	pos = helpers.PosFromInd(7, 5)
	g.Pieces[pos] = pieces.NewKing(enums.White, pos)
}
