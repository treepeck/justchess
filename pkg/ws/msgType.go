package ws

// Message types.
const (
	// Sent by clients.
	GET_AVAILIBLE_GAMES byte = iota
	CREATE_GAME
	JOIN_GAME
	GET_GAME
	LEAVE_GAME
	// Sent by server.
	CLIENTS_COUNTER
	ADD_GAME
	REMOVE_GAME
	REDIRECT
	GAME_INFO
)
