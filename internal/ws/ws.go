// Package ws implements the WebSocket server.
package ws

import (
	"log"
	"net/http"

	"justchess/internal/db"
	"justchess/internal/web"

	"github.com/gorilla/websocket"
	"github.com/treepeck/chego"
	"github.com/treepeck/glicko"
)

const (
	msgNotFound string = "Connection will be closed: provided id is not valid."
	msgConflict string = "Please close any previous tabs and reload the page to reconnect"

	minRating    = 10
	maxRating    = 4000
	minDeviation = 30
	minSigma     = 0.04
	maxSigma     = 0.08
)

// upgrader is used to establish a WebSocket connection.
// It is safe for concurrent use.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type createRoom struct {
	id      string
	white   db.Player
	black   db.Player
	control web.QueueData
	res     chan error
}

type storeGame struct {
	white       db.Player
	black       db.Player
	moves       []completedMove
	id          string
	result      chego.Result
	termination chego.Termination
}

type findRegister struct {
	id  string
	res chan chan *client
}

// Service manages the [room] lifecycle (creation and deletion), handles
// incomming handshake requests, and modifies the database (stores completed
// games and updated player's ratings).
type Service struct {
	playerRepo db.PlayerRepo
	gameRepo   db.GameRepo
	create     chan createRoom
	remove     chan string
	find       chan findRegister
	store      chan storeGame
	rooms      map[string]*room
	queues     map[string]queue
}

func NewService(pr db.PlayerRepo, gr db.GameRepo) Service {
	s := Service{
		playerRepo: pr,
		gameRepo:   gr,
		create:     make(chan createRoom),
		remove:     make(chan string),
		find:       make(chan findRegister),
		store:      make(chan storeGame),
		rooms:      make(map[string]*room),
		queues:     make(map[string]queue, 9),
	}

	for i := byte(1); i < 10; i++ {
		q := newQueue(i)
		go q.listenEvents(s.create)
		s.queues[string(i+'0')] = q
	}

	return s
}

// RegisterRoute registers the handshake enpoint to the specified ServeMux.
func (s Service) RegisterRoute(mux *http.ServeMux) {
	mux.HandleFunc("/ws", s.handshake)
}

func (s Service) ListenEvents() {
	for {
		select {
		case e := <-s.create:
			s.handleCreateRoom(e)
		case id := <-s.remove:
			s.handleRemoveRoom(id)
		case e := <-s.find:
			if q, exist := s.queues[e.id]; exist {
				e.res <- q.register
			} else if r, exist := s.rooms[e.id]; exist {
				e.res <- r.register
			} else {
				e.res <- nil
			}
		case e := <-s.store:
			s.handleStoreGame(e)
		}
	}
}

// handshake handles WebSocket handshake requests.  Each incoming request must
// include an 'id' parameter that identifies the room or queue the client is
// attempting to join.  The request will be denied if the session cookie is
// missing or expired.
//
// An error event will be sent to the client immediately after the connection
// is opened in the following cases:
//   - No room or queue exists with the provided id;
//   - The client is already registered in the room or queue.
//
// The connection will be closed after the error event is sent.
func (s Service) handshake(rw http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("Auth")
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	p, err := s.playerRepo.SelectBySessionId(session.Value)
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	id := r.URL.Query().Get("id")

	// Create WebSocket connection.
	conn, err := upgrader.Upgrade(rw, r, nil)
	if err != nil {
		// Simply return here since the upgrader writes the response.
		return
	}
	c := newClient(conn, p)
	go c.read()
	go c.write()

	// Search for a room or queue with the given id.
	e := findRegister{
		id:  id,
		res: make(chan chan *client, 1),
	}
	s.find <- e

	// Handle response.
	if register := <-e.res; register != nil {
		register <- c
		return
	}

	// Send error event to the client.
	if raw, err := newEncodedEvent(actionError, msgNotFound); err == nil {
		c.send <- raw
	}
	c.conn.Close()
}

func (s Service) handleCreateRoom(e createRoom) {
	err := s.gameRepo.Insert(
		e.id, e.white.Id, e.black.Id,
		e.control.Control, e.control.Bonus,
	)
	defer func() { e.res <- err }()
	if err != nil {
		return
	}

	log.Printf("room %s created", e.id)

	r := newRoom(e.id, e.white, e.black, e.control.Control, e.control.Bonus, s.store)
	go r.listenEvents(s.remove)

	s.rooms[e.id] = r
}

func (s Service) handleRemoveRoom(id string) {
	if s.rooms[id] == nil {
		log.Printf("room %s doesn't exist", id)
		return
	}

	log.Printf("room %s removed", id)
	delete(s.rooms, id)
}

// handleStoreGame stores the game state in database. If the game is over, it
// also updates and stores players' ratings based on game outcome.
func (s Service) handleStoreGame(e storeGame) {
	// Shortcut: don't encode moves or update players' ratings if game was
	// abandoned.
	if e.termination == chego.Abandoned {
		err := s.gameRepo.MarkGameAsAbandoned(e.id)
		if err != nil {
			log.Print(err)
		}
		return
	}

	// Prepare moves for encoding and time differences for compression.
	indices := make([]byte, len(e.moves))
	diffs := make([]int, len(e.moves))
	for i, m := range e.moves {
		indices[i] = m.index
		diffs[i] = m.timeDiff
	}

	// Write game state to database.
	if err := s.gameRepo.Update(
		e.result, e.termination,
		len(e.moves),
		chego.HuffmanEncoding(indices),
		chego.CompressTimeDiffs(diffs),
		e.id,
	); err != nil {
		log.Print(err)
		return
	}

	if e.termination != chego.Unterminated {
		err := s.updateRatings(e.white, e.black, e.result)
		if err != nil {
			log.Print(err)
		}
	}
}

// Updates white and black player ratings based on the single match outcome.
func (s Service) updateRatings(white, black db.Player, r chego.Result) error {
	c := glicko.Converter{
		Rating:    glicko.DefaultRating,
		Deviation: glicko.DefaultDeviation,
		Factor:    glicko.DefaultFactor,
	}

	// Initial players' strength.
	wStr := glicko.Strength{
		Mu:    c.Rating2Mu(white.Rating),
		Phi:   c.Deviation2Phi(white.Deviation),
		Sigma: white.Volatility,
	}
	bStr := glicko.Strength{
		Mu:    c.Rating2Mu(black.Rating),
		Phi:   c.Deviation2Phi(black.Deviation),
		Sigma: black.Volatility,
	}

	var whiteScore, blackScore float64
	switch r {
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

	return s.playerRepo.UpdateRatings(white.Id, black.Id,
		c.Mu2Rating(wStr.Mu), c.Phi2Deviation(wStr.Phi), wStr.Sigma,
		c.Mu2Rating(bStr.Mu), c.Phi2Deviation(bStr.Phi), bStr.Sigma,
	)
}
