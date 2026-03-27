import { g } from "../utils/dom"

/**
 * @param {string} id
 * @param {number} ms
 */
export function formatTime(id, ms) {
	const minutes = Math.trunc(ms / 1000 / 60)
	const seconds = Math.trunc(ms / 1000) % 60

	let mins = `${minutes > 9 ? minutes : "0" + minutes.toString()}`
	let secs = `${seconds > 9 ? seconds : "0" + seconds.toString()}`
	g(id).textContent = `${mins}:${secs}`
}

/**
 * Enum representing player's colors.
 * @enum {number}
 */
export const Color = /** @type {const} */ ({
	White: 0,
	Black: 1,
})

/** Countdown chess clock. */
export class Clock {
	/**
	 * Displayed time left on white player's clock in milliseconds.
	 * @type {number}
	 */
	whiteTime
	/**
	 * Displayed time left on black player's clock in milliseconds.
	 * @type {number}
	 */
	blackTime
	/**
	 * Is clock counting down.
	 * @type {boolean}
	 */
	isActive
	/**
	 * Interval of time ticks in milliseconds.
	 * @type {number}
	 */
	interval
	/**
	 * Active color.
	 * @type {Color}
	 */
	color
	/**
	 * For self-adjustment.
	 * @type {number}
	 */
	expected

	/**
	 * Initializes the clock state without actually starting it.
	 * @param {number} time
	 * @param {boolean} isActive
	 * @param {Color} color
	 * @param {number} interval
	 */
	constructor(time, isActive, color, interval) {
		this.whiteTime = time
		this.blackTime = time
		this.isActive = isActive
		this.color = color
		this.interval = interval
		this.expected = 0
	}

	/** Handles time ticks. */
	tick() {
		if (
			!this.isActive ||
			this.whiteTime < this.interval ||
			this.blackTime < this.interval
		)
			return

		const delta = Date.now() - this.expected
		this.expected = Date.now() + this.interval

		if (this.color == Color.White) {
			this.whiteTime -= this.interval + delta
			formatTime("whiteClock", this.whiteTime)
		} else {
			this.blackTime -= this.interval + delta
			formatTime("blackClock", this.blackTime)
		}
		setTimeout(() => this.tick(), this.interval + delta)
	}

	start() {
		this.isActive = true
		this.expected = Date.now() + this.interval
		setTimeout(() => this.tick(), this.interval)
	}

	/** Stops the clock. */
	stop() {
		this.isActive = false
	}

	/** Switches the active color */
	flip() {
		this.color ^= 1
	}

	/**
	 * Sets the player's remaining time and updates the UI.
	 * Safe for concurrent use with start since browser JS is single threaded.
	 * @param {Color} color
	 * @param {number} time Seconds
	 */
	setTime(color, time) {
		if (color == Color.White) {
			this.whiteTime = time
			formatTime("whiteClock", time)
		} else {
			this.blackTime = time
			formatTime("blackClock", time)
		}
	}
}
