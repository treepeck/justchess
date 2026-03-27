package game

// clock stores all time values in seconds.
type clock struct {
	whiteTime      int
	blackTime      int
	whiteReconnect int
	blackReconnect int
	bonus          int
	timeBeforeMove int
}

func newClock(control, bonus int) *clock {
	return &clock{
		whiteTime:      control,
		blackTime:      control,
		whiteReconnect: reconnectDeadline,
		blackReconnect: reconnectDeadline,
		bonus:          bonus,
		timeBeforeMove: control,
	}
}
