package ws

type MessageType = byte

const (
	// Sent by clients.
	CREATE_ROOM MessageType = iota
	JOIN_ROOM
	MAKE_MOVE

	// Sent by server.
	CLIENTS_COUNTER
	ADD_ROOM
	REMOVE_ROOM
	CHAT_MESSAGE
	ROOM_INFO
	GAME
	LAST_MOVE
	RESULT
)
