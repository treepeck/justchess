export const Piece = {
	WP: 0,
	BP: 1,
	WN: 2,
	BN: 3,
	WB: 4,
	BB: 5,
	WR: 6,
	BR: 7,
	WQ: 8,
	BQ: 9,
	WK: 10,
	BK: 11,
	NP: -1, // No piece.
}

export class DraggedPiece {
	constructor(x, y, piece, fromSquare) {
		this.x = x
		this.y = y
		this.piece = piece
		this.fromSquare = fromSquare
	}
}

// Must be exactly the same as in justchess/internal/ws/event.go
export const Action = {
	Ping: 0,
	Pong: 1,
	MakeMove: 2,
}

export class WSEvent {
	constructor(action, payload) {
		this.action = action
		this.payload =  payload
	}

	toJSON() {
		return JSON.stringify({
			a: this.action,
			p: this.payload,
		})
	}
}