// Package proto defines the protocol for interconnected game and
// ws packages to communicate in a distributed manner.
package proto

// Conn  - player id, game id.
// Disc  - player id, game id.
// Move  - player id, game id, move index.
// I do not need FEN at all. Simply send UCI and SAN.
// The frontend will update the board position itself by performing UCI.
// The SAN will be displayed on frontend. To redo the move, simply
// Restore the position to the beginning, and then redo all moves to the selected
// move index.
// We also need to store the FEN of the starting position, this will become VERY handy
// later on when I will implement 360 random chess. Because then the starting position
// will become not fixed.
// So, when a player connects to the room, we need to send the FEN of the starting position
// (and that's why we need FEN), also it needs to send the array of completed SANs and UCIs.
// To render the active state.
// Story - game id, fen[], san[].
//

Client <-JSON-> WS <-GOB-> GAME

// Status is sent by the game server to notify the clients about active
type Status struct {
	WhiteId    string `json:""`
	BlackId    string `json:""`
	Spectators int    `json:""`
}

// Story is sent by the game server when a new client connects to the game so it
// is able to recieve the active game state.
type Story struct {
	SANs       []string `json:""`
	UCIs       []string `json:""`
	InitialFEN string   `json:""`
}

type Event struct {
	GameId, PlayerId string
}
