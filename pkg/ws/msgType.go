package ws

// Message types.
const (
	// Sent by clients.
	CREATE_ROOM = iota
	JOIN_ROOM
	LEAVE_ROOM
	MOVE
	// Sent by server.
	CLIENTS_COUNTER
	ADD_ROOM
	REMOVE_ROOM
	LAST_MOVE
	MOVES
	GAME_INFO
)
