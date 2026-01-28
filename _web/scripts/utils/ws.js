// Must match event actions from backend (justchess/internal/ws/event.go).
export const EventAction = {
	Ping: 0,
	Pong: 1,
	Chat: 2,
	Move: 3,
	ClientsCounter: 4,
	Redirect: 5,
}
