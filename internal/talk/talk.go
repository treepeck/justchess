// Package talk declares transport-layer types that allow game and ws
// packages to communicate in a distributed manner.
package talk

// GameCreator is a data used to create a new rated game after the best match
// was found.
type GameCreator struct {
	Id          string
	WhiteId     string
	BlackId     string
	TimeControl int
	TimeBonus   int
	Res         chan error
}

// GameFinder is a data used to find a game in a game storage. If the game wasn't
// found, the Res chan will recieve [GameChannels] with nil values.
type GameFinder struct {
	Id  string
	Res chan GameChannels
}

// GameChannels defines channels which are used for communication between
// WebSocket room and game.
type GameChannels struct {
	// In recieves messages from players.
	In chan Message
	// Out recieves messages from game. Those are respones to handled player messages.
	Out chan []byte
	// Ban is needed to be able to terminate connections of malicious players.
	Ban chan string
}
