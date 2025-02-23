package enums

type Status byte

const (
	NotStarted Status = iota
	Continues
	WhiteDisconnected
	BlackDisconnected
)
