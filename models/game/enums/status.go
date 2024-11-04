package enums

type Status int

const (
	Aborted   Status = iota // one of the players did not make the first move or cancelled the game.
	Waiting                 // the game doesn't start until both sides connect.
	Continues               // game continues.
	Over
)
