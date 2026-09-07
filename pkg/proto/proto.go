// Package proto defines the protocol for interconnected game and
// ws packages to communicate in a distributed manner.
package proto

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
