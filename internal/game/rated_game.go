package game

import (
	"justchess/internal/db"
	"log"

	"github.com/treepeck/chego"
	"github.com/treepeck/glicko"
)

const (
	minRating    = 10
	maxRating    = 4000
	minDeviation = 30
	minSigma     = 0.04
	maxSigma     = 0.08
)

type RatedGame struct {
	chego.Game

	white db.Player
	black db.Player
	// Indices of played moves for Huffman coding.
	playedIndices []byte
	// Amount of seconds players spent on played moves. For compression.
	timeDiffs         []int
	gameRepo          db.GameRepo
	playerRepo        db.PlayerRepo
	drawIssuer        string
	id                string
	clock             *clock
	didWhiteOfferDraw bool
	bidBlackOfferDraw bool
	isWhiteOnline     bool
	isBlackOnline     bool
}

// SpawnRatedGame inserts a new rated game record into repository and initializes
// [RatedGame] fields.
func SpawnRatedGame(
	white, black db.Player, control, bonus int,
	id string, gr db.GameRepo, pr db.PlayerRepo,
) (*RatedGame, error) {
	if err := gr.InsertRated(id, white.Id, black.Id, control, bonus); err != nil {
		return nil, err
	}
	return &RatedGame{
		id:            id,
		Game:          chego.NewGame(),
		white:         white,
		black:         black,
		gameRepo:      gr,
		playerRepo:    pr,
		playedIndices: make([]byte, 0),
		timeDiffs:     make([]int, 0),
		clock:         newClock(control, bonus),
	}, nil
}

// Play performes the move with the specified index.
func (g *RatedGame) Play(id string, index byte) (MovePayload, bool) {
	if (len(g.Played)%2 == 0 && id != g.white.Id) ||
		(len(g.Played)%2 != 0 && id != g.black.Id) ||
		g.Termination != chego.Unterminated ||
		index > g.Legal.LastMoveIndex {
		return MovePayload{}, false
	}

	g.Push(g.Legal.Moves[index])

	// Store time after completing the move to synchronize clock on frontend.
	var timeDiff, timeLeft int
	if g.Position.ActiveColor == chego.ColorWhite {
		timeDiff = g.clock.blackTime - g.clock.timeBeforeMove
		g.clock.timeBeforeMove = g.clock.whiteTime
		timeLeft = g.clock.blackTime
	} else {
		timeDiff = g.clock.whiteTime - g.clock.timeBeforeMove
		g.clock.timeBeforeMove = g.clock.blackTime
		timeLeft = g.clock.whiteTime
	}

	g.timeDiffs = append(g.timeDiffs, timeDiff)
	g.playedIndices = append(g.playedIndices, index)

	if g.Termination != chego.Unterminated {
		g.store()
	}

	return MovePayload{
		Legal:      g.Legal.Moves[:g.Legal.LastMoveIndex],
		PlayedMove: g.Played[len(g.Played)-1],
		TimeLeft:   timeLeft,
	}, true
}

func (g *RatedGame) Join(id string) {
	switch id {
	case g.white.Id:
		g.isWhiteOnline = true
	case g.black.Id:
		g.isBlackOnline = true
	}
	log.Printf("player %s joins game %s", id, g.id)
}

func (g *RatedGame) Leave(id string) {
	switch id {
	case g.white.Id:
		g.isWhiteOnline = false
	case g.black.Id:
		g.isBlackOnline = false
	}
	log.Printf("player %s leaves game %s", id, g.id)
}

// TimeTick decrements player's time each second. It's the caller's responsibility
// to ensure this function called only when the game is not terminated.
func (g *RatedGame) TimeTick() {
	if g.Termination != chego.Unterminated {
		return
	}

	// If some player is disconnected, decrement their reconnect time.
	if !g.isWhiteOnline {
		g.clock.whiteReconnect--
	}
	if !g.isBlackOnline {
		g.clock.blackReconnect--
	}

	// Decrement player's time.
	if g.Position.ActiveColor == chego.ColorWhite {
		g.clock.whiteTime--
	} else {
		g.clock.blackTime--
	}

	// Terminate the game due to the time forfeit.
	if g.clock.whiteTime == 0 || g.clock.blackTime == 0 ||
		g.clock.whiteReconnect == 0 || g.clock.blackReconnect == 0 {
		if g.IsInsufficientMaterial() {
			// According to chess rules, the game is draw if opponent doesn't have
			// sufficient material to checkmate you.
			g.Terminate(chego.TimeForfeit, chego.Draw)
		} else if len(g.Played) < minMoves {
			// Mark game as abandoned if there was not enough moves played.
			g.Abandon()
		} else {
			if g.clock.blackTime == 0 || g.clock.blackReconnect == 0 {
				g.Terminate(chego.TimeForfeit, chego.WhiteWon)
			} else {
				g.Terminate(chego.TimeForfeit, chego.BlackWon)
			}
		}
	}

	if g.Termination != chego.Unterminated {
		g.store()
	}
}

func (g *RatedGame) Resign(id string) bool {
	if len(g.Played) < minMoves || g.Termination != chego.Unterminated {
		return false
	}

	switch id {
	case g.white.Id:
		g.Terminate(chego.Resignation, chego.BlackWon)
	case g.black.Id:
		g.Terminate(chego.Resignation, chego.WhiteWon)
	}
	g.store()
	return true
}

// OfferDraw handles draw offers. Offer will be discarded if one of the
// following is true:
//   - There were not enough moves played to terminate the game;
//   - The game is already terminated;
//   - Sender is not a white nor a black player;
//   - One of the players has already sent a pending draw offer;
//   - The player has sent a draw offer not so long ago.
//
// Returns empty string if offer was discarded and opponent id otherwise.
func (g *RatedGame) OfferDraw(id string) string {
	if len(g.Played) < minMoves ||
		g.Termination != chego.Unterminated ||
		(id != g.white.Id && id != g.black.Id) ||
		(id == g.white.Id && g.didWhiteOfferDraw) ||
		(id == g.black.Id && g.bidBlackOfferDraw) ||
		len(g.drawIssuer) != 0 {
		return ""
	}
	g.drawIssuer = id

	switch id {
	case g.white.Id:
		g.didWhiteOfferDraw = true
		return g.black.Id
	default:
		g.bidBlackOfferDraw = true
		return g.white.Id
	}
}

func (g *RatedGame) AcceptDraw(id string) bool {
	if len(g.drawIssuer) == 0 ||
		id == g.drawIssuer ||
		(id != g.white.Id && id != g.black.Id) {
		return false
	}
	g.drawIssuer = ""
	g.Terminate(chego.Agreement, chego.Draw)
	g.store()
	return true
}

func (g *RatedGame) DeclineDraw(id string) bool {
	if len(g.drawIssuer) == 0 ||
		id == g.drawIssuer ||
		(id != g.white.Id && id != g.black.Id) {
		return false
	}
	g.drawIssuer = ""
	return true
}

func (g *RatedGame) Abandon() {
	if g.Termination != chego.Unterminated {
		g.Termination = chego.Abandoned
		g.gameRepo.MarkRatedAsAbandoned(g.id)
	}
}

func (g *RatedGame) store() {
	if err := g.gameRepo.UpdateRated(db.RatedGameUpdate{
		Id: g.id, Result: g.Result, Termination: g.Termination,
		EncodedMoves:    chego.HuffmanEncoding(g.playedIndices),
		CompressedDiffs: chego.CompressTimeDiffs(g.timeDiffs),
		MovesLength:     len(g.Played),
	}); err != nil {
		log.Print(err)
		return
	}

	if err := g.updateRatings(); err != nil {
		log.Print(err)
	}
}

func (g *RatedGame) updateRatings() error {
	c := glicko.Converter{
		Rating:    glicko.DefaultRating,
		Deviation: glicko.DefaultDeviation,
		Factor:    glicko.DefaultFactor,
	}

	// Initial players' strength.
	wStr := glicko.Strength{
		Mu:    c.Rating2Mu(g.white.Rating),
		Phi:   c.Deviation2Phi(g.white.Deviation),
		Sigma: g.white.Volatility,
	}
	bStr := glicko.Strength{
		Mu:    c.Rating2Mu(g.black.Rating),
		Phi:   c.Deviation2Phi(g.black.Deviation),
		Sigma: g.black.Volatility,
	}

	var whiteScore, blackScore float64
	switch g.Result {
	case chego.WhiteWon:
		whiteScore = 1
		blackScore = 0
	case chego.BlackWon:
		whiteScore = 0
		blackScore = 1
	case chego.Draw:
		whiteScore = 0.5
		blackScore = 0.5
	}

	wOut := glicko.Outcome{
		Mu:    bStr.Mu,
		Phi:   bStr.Phi,
		Score: whiteScore,
	}
	bOut := glicko.Outcome{
		Mu:    wStr.Mu,
		Phi:   wStr.Phi,
		Score: blackScore,
	}

	e := glicko.Estimator{
		MinMu:    c.Rating2Mu(minRating),
		MaxMu:    c.Rating2Mu(maxRating),
		MinPhi:   c.Deviation2Phi(minDeviation),
		MaxPhi:   c.Deviation2Phi(glicko.DefaultDeviation),
		MinSigma: minSigma, MaxSigma: maxSigma,
		Tau: glicko.DefaultTau, Epsilon: glicko.DefaultEpsilon,
	}

	e.Estimate(&wStr, wOut, 1)
	e.Estimate(&bStr, bOut, 1)

	return g.playerRepo.UpdateRatings(
		db.RatingUpdate{
			Id:         g.white.Id,
			Rating:     c.Mu2Rating(wStr.Mu),
			Deviation:  c.Phi2Deviation(wStr.Phi),
			Volatility: wStr.Sigma,
		},
		db.RatingUpdate{
			Id:         g.black.Id,
			Rating:     c.Mu2Rating(bStr.Mu),
			Deviation:  c.Phi2Deviation(bStr.Phi),
			Volatility: bStr.Sigma,
		},
	)
}

func (g *RatedGame) GamePayload() GamePayload {
	return GamePayload{
		Legal:     g.Legal.Moves[:g.Legal.LastMoveIndex],
		Played:    g.Played,
		WhiteTime: g.clock.whiteTime,
		BlackTime: g.clock.blackTime,
	}
}

func (g *RatedGame) EndPayload() EndPayload {
	return EndPayload{
		Termination: g.Termination,
		Result:      g.Result,
	}
}
