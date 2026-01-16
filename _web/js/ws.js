// Must match event actions from backend (justchess/internal/ws/event.go).
export const EventAction = {
	Ping: 0,
	Pong: 1,
	MakeMove: 2,
	ClientsCounter: 3,
	Redirect: 4,
}
