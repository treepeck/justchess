import { getOrPanic } from "./dom"

/**
 * @param {string} id
 * @param {number} seconds
 */
export function formatTime(id, seconds) {
	const minutes = Math.floor(seconds / 60)
	if (minutes > 0) {
		seconds -= 60 * minutes
	}

	getOrPanic(id).textContent = `${minutes > 9 ? minutes : `0${minutes}`}:${
		seconds > 9 ? seconds : `0${seconds}`
	}`
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
	 * Displayed time left on white player's clock in seconds.
	 * @type {number}
	 */
	whiteTime
	/**
	 * Displayed time left on black player's clock in seconds.
	 * @type {number}
	 */
	blackTime
	/**
	 * Is clock counting down.
	 * @type {boolean}
	 */
	isActive
	/**
	 * Active color.
	 * @type {Color}
	 */
	color

	/**
	 * Initialize the clock state without actually starting it.
	 * @param {number} time
	 * @param {boolean} isActive
	 * @param {Color} color
	 */
	constructor(time, isActive, color) {
		this.whiteTime = time
		this.blackTime = time
		this.isActive = isActive
		this.color = color
	}

	/**
	 * Starts the clock which will tick with the given interval.
	 * @param {number} interval Milliseconds.
	 */
	async start(interval) {
		// Sleep for a single interval.
		while (this.isActive) {
			await new Promise((res) => setTimeout(res, interval))
			// Handle time tick.
			if (this.color == Color.White && this.whiteTime > 1) {
				this.whiteTime--
				formatTime("whiteClock", this.whiteTime)
			} else if (this.blackTime > 1) {
				this.blackTime--
				formatTime("blackClock", this.blackTime)
			}
		}
	}

	/** Stops the clock. */
	stop() {
		this.isActive = false
	}

	/** Switches the active color */
	switchColor() {
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
