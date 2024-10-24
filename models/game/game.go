package game

import (
	"chess-api/models/game/enums"
	"chess-api/models/game/helpers"
	"chess-api/models/game/pieces"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// G represents a game and stores all necessary data.
type G struct {
	Id        uuid.UUID                     `json:"id"`        // game id.
	Bonus     uint                          `json:"bonus"`     // time increment.
	Moves     []helpers.Move                `json:"moves"`     // completed moves.
	Pieces    map[helpers.Pos]pieces.Piece  `json:"pieces"`    // board state.
	Status    enums.Status                  `json:"status"`    // game status.
	Control   enums.Control                 `json:"control"`   // time control.
	WhiteId   uuid.UUID                     `json:"whiteId"`   // white player id.
	BlackId   uuid.UUID                     `json:"blackId"`   // black player id.
	PlayedAt  time.Time                     `json:"playedAt"`  // when the game was started.
	WhiteTime time.Duration                 `json:"whiteTime"` // how much time do the white have left.
	BlackTime time.Duration                 `json:"blackTime"` // how much time do the black have left.
	cvm       map[helpers.PossibleMove]bool // current valid moves.
}

// NewG creates a new game.
func NewG(id uuid.UUID, control enums.Control,
	bonus uint, whiteId, blackId uuid.UUID,
) *G {
	g := &G{
		Id:        id,
		Bonus:     bonus,
		Moves:     make([]helpers.Move, 0),
		Pieces:    make(map[helpers.Pos]pieces.Piece),
		Status:    enums.Waiting,
		Control:   control,
		WhiteId:   whiteId,
		BlackId:   blackId,
		PlayedAt:  time.Now(),
		WhiteTime: control.ToDuration(),
		BlackTime: control.ToDuration(),
		cvm:       make(map[helpers.PossibleMove]bool),
	}
	g.initPieces()
	// get white player valid moves in advance to increase performance
	g.cvm = g.getPlayerValidMoves(enums.White)
	return g
}

// getPlayerValidMoves finds all valid moves for the specified player.
// The validity of the returned moves is guaranteed - player`s possible
// moves are filtered down by the validity checking.
// For more details about move validity, see [chess-api/models/pieces.GetPossibleMoves].
func (g *G) getPlayerValidMoves(side enums.Color,
) map[helpers.PossibleMove]bool {
	ppm := getPlayerPossibleMoves(side, g.Pieces)
	// determine enemy side
	var es enums.Color
	if side == enums.White {
		es = enums.Black
	} else {
		es = enums.White
	}
	// store valid moves
	vm := make(map[helpers.PossibleMove]bool)
	for pm := range ppm { // iterate over all possible moves
		// skip defend moves
		if pm.MoveType == enums.Defend {
			continue
		}
		// create a board copy
		bc := make(map[helpers.Pos]pieces.Piece)
		for pos, piece := range g.Pieces {
			// piece is a pointer! so modifying the piece in board copy,
			// piece in original board (g.Pieces) is modified too!
			bc[pos] = piece
		}
		// make sure that after making this move the allied king is not checked
		isChecked := false
		bc[pm.From].Move(pm.To)
		// update board state
		bc[pm.To] = bc[pm.From]
		delete(bc, pm.From)
		// find all enemy possible moves on a new position
		eppm := getPlayerPossibleMoves(es, bc)
		for epm := range eppm {
			p := bc[epm.To]
			// if the allied king is checked
			if p != nil && p.GetType() == enums.King && p.GetColor() == side {
				isChecked = true
				// return the piece to original pos!
				bc[pm.To].Move(pm.From)
				bc[pm.To].SetMovesCounter(bc[pm.To].GetMovesCounter() - 2)
				break
			}
		}
		// if the allied king remained in a safe position, the move is valid
		if !isChecked {
			vm[pm] = true
			// return the piece to original pos!
			bc[pm.To].Move(pm.From)
			bc[pm.To].SetMovesCounter(0)
		}
	}
	return vm
}

// getPlayerPossibleMoves finds all possible moves for the specified player
// and board.
// The validity of the returned moves is not guaranteed.
// For more details about move validity, see [chess-api/models/pieces.GetPossibleMoves].
func getPlayerPossibleMoves(side enums.Color,
	pieces map[helpers.Pos]pieces.Piece) map[helpers.PossibleMove]bool {
	pm := make(map[helpers.PossibleMove]bool)
	for from, piece := range pieces {
		if piece.GetColor() == side {
			ppm := piece.GetPossibleMoves(pieces)
			for pos, mt := range ppm {
				pm[helpers.PossibleMove{
					To:        pos,
					From:      from,
					MoveType:  mt,
					PieceType: piece.GetType(),
				}] = true
			}
		}
	}
	return pm
}

// HandleMove handles player`s moves. True is retured if the move is valid,
// false otherwise.
func (g *G) HandleMove(m helpers.Move) bool {
	for vm := range g.cvm {
		if vm.From.IsEqual(m.From) &&
			vm.To.IsEqual(m.To) {
			// move the piece
			g.Pieces[m.From].Move(m.To)
			// determine is move a capture
			if g.Pieces[m.To] != nil {
				m.IsCapture = true
			}
			// update the board state
			g.Pieces[m.To] = g.Pieces[m.From]
			delete(g.Pieces, m.From)

			g.processLastMove(&m)
			return true
		}
	}
	return false
}

func (g *G) processLastMove(lastMove *helpers.Move) {
	// get opponent`s valid moves to determine checkmate
	es := enums.Black          // enemy side
	if len(g.Moves)+1%2 == 0 { // black has moved
		es = enums.White
	}
	g.cvm = g.getPlayerValidMoves(es)

	// if the opponent does not have any valid moves,
	// it is either a stalemate or a checkmate
	if len(g.cvm) == 0 {
		g.determineCheck(lastMove)
		if lastMove.IsCheck {
			lastMove.IsCheckmate = true
			// TODO: game.endGame
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

// MarshalJSON serializes game object to a json.
func (g *G) MarshalJSON() ([]byte, error) {
	gameDTO := struct {
		Id       uuid.UUID               `json:"id"`
		Control  enums.Control           `json:"control"`
		Bonus    uint                    `json:"bonus"`
		Status   enums.Status            `json:"status"`
		WhiteId  uuid.UUID               `json:"whiteId"`
		BlackId  uuid.UUID               `json:"blackId"`
		PlayedAt time.Time               `json:"playedAt"`
		Moves    []helpers.Move          `json:"moves"`
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
