// Must match event actions from backend.
// Internal actions are skipped.
export const EventAction = {
	Ping: 0,
	Pong: 1,
	JoinMatchmaking: 2,
	LeaveMatchmaking: 3,
	MakeMove: 4,
}
