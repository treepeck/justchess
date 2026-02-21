/**
 * Must match the event actions defined in [justchess/internal/ws/event.go].
 * @enum {number}
 */
export const EventAction = /** @type {const} */ ({
	Ping: 0,
	Pong: 1,
	Chat: 2,
	Move: 3,
	Game: 4,
	End: 5,
	Conn: 6,
	Disc: 7,
	ClientsCounter: 8,
	Redirect: 9,
	Error: 10,
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
