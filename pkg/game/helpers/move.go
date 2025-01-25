package helpers

// 0-5: To (destination) square index;
// 6-11: From (origin/source) square index;
// 12-14: Move type [see justchess/pkg/game/enums/moveType];
// 15: Unused.
type Move uint16

func NewMove(to, from, mt int) Move {
	return Move(to | (from << 6) | (mt << 12))
}

func (m Move) To() int {
	return int(m) & 0x3F
}

func (m Move) From() int {
	return int(m>>6) & 0x3F
}

func (m Move) MoveType() int {
	return int(m>>12) & 0x7
}
