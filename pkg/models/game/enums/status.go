package enums

type Status int

const (
	Aborted   Status = iota // one of the players did not make the first move or cancelled the game.
	Waiting                 // the game doesn't start until both sides connect.
	Leave                   // one of the players leave the room and has 20 seconds to reconnect.
	Continues               // game continues.
	Over                    // game is over and stored in a database.
)
