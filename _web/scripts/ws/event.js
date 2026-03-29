/**
 * Must match the event Kinds defined in [justchess/internal/ws/event.go].
 * @enum {number}
 */
export const EventKind = /** @type {const} */ ({
	Ping: 0,
	Pong: 1,
	Chat: 2,
	Resign: 3,
	OfferDraw: 4,
	AcceptDraw: 5,
	DeclineDraw: 6,
	Move: 7,
	Game: 8,
	End: 9,
	Conn: 10,
	Disc: 11,
	ClientsCounter: 12,
	Redirect: 13,
	Error: 14,
})

/**
 * Arbitrary event recieved from the server.
 * @typedef {Object} Event
 * @property {EventKind} k
 * @property {any} p - Specific type depents on the Kind.
 */

/**
 * @typedef {Object} MovePayload
 * @property {import("../components/board.js").Move[]} lm - Legal moves for the next player.
 * @property {string} s - San.
 * @property {string} f - Fen.
 * @property {number} tl - Time left.
 */

/**
 * @typedef {Object} PlayedMove
 * @property {string} s - Standard Algebraic Notation of the move.
 * @property {string} f -  Serialized piece placement (Forsyth-Edwards Notation).
 */

/**
 * Payload of the event with Game Kind.
 * @typedef {Object} GamePayload
 * @property {import("../components/board.js").Move[]} lm - Legal moves for current turn.
 * @property {PlayedMove[]} m - All played moves.
 * @property {number} wt - White time in seconds.
 * @property {number} bt - Black time in seconds.
 */

/**
 * Payload of the event with End Kind.
 * @typedef {Object} EndPayload
 * @property {import("../utils/state.js").Termination} t
 * @property {import("../utils/state.js").Result} r
 */
