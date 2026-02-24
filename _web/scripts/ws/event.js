/**
 * Must match the event actions defined in [justchess/internal/ws/event.go].
 * @enum {number}
 */
export const EventAction = /** @type {const} */ ({
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
 * @property {EventAction} a
 * @property {any} p - Specific type depents on the action.
 */

/**
 * @typedef {Object} MovePayload
 * @property {import("../chess/move.js").Move[]} lm - Legal moves for the next player.
 * @property {number} t - Remaining time on the player's clock.
 * @property {import("../chess/move.js").CompletedMove} m
 */

/**
 * Payload of the event with Game action.
 * @typedef {Object} GamePayload
 * @property {import("../chess/move.js").Move[]} lm - Legal moves for current turn.
 * @property {import("../chess/move.js").CompletedMove[]} m - All completed moves.
 * @property {number} wt - White player's remaining time.
 * @property {number} bt - Black player's remaining time.
 */

/**
 * Payload of the event with End action.
 * @typedef {Object} EndPayload
 * @property {string} t - Formatted termination.
 * @property {string} r - Formatted result.
 */
