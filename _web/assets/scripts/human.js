import { Conn } from "/assets/scripts/ws/ws.js"
import { MessageKind, decodeInt } from "/assets/scripts/ws/proto.js"

const hints = [
	"Keep your king safe and connect your rooks",
	"Develop knights and bishops before moving the same piece multiple times",
	"Don't bring your queen out too early",
	"Try to occupy or influence the central squares during early game",
	"Ask: what is my opponent threatening?",
	"Pieces are strongest when they work together",
	"Sacrificing material can sometimes lead to victory",
	"Don't rush trades",
	"Master the basics before experimenting with your own openings",
	"A rook belongs on open files",
	"Don't forget to take small breaks between matches",
	"If the board position is not in your favor, try to put pressure on your opponent's time",
]

class State {
	/**
	 * @type {number}
	 */
	timerTimeout
	/**
	 * @type {number}
	 */
	hintTimeout
	/**
	 * @type {HTMLDivElement}
	 */
	timer
	/**
	 * @type {Conn}
	 */
	conn
	/**
	 * @type {string[]}
	 */
	hintsStack
	/**
	 * @type {HTMLDivElement}
	 */
	hintBox
	/**
	 * @type {HTMLDivElement}
	 */
	playersNumber

	constructor() {
		this.timerTimeout = -1
		this.hintTimeout = -1
		this.timer = document.getElementById("countUpTimer")
		this.hintBox = document.getElementById("hintBox")
		this.hintsStack = []
		this.playersNumber = document.getElementById("playersNumber")
	}

	set() {
		this.displayHint()
		this.displayTime()
		this.conn = new Conn(`${slider.value}`)
		this.conn.handlers.set(MessageKind.KindCounter, (e) => { this.displayCounter(e) })
		this.conn.handlers.set(MessageKind.KindMatch, (e) => {
			window.location.href = `/game/${new TextDecoder().decode(e)}`
		})
	}

	reset() {
		clearTimeout(this.timerTimeout)
		clearTimeout(this.hintTimeout)
		this.timer.textContent = "00:00"

		this.conn.cleanup()
		this.conn = null
	}

	displayTime() {
		// Self-adjusting countup timer.
		const interval = 1000 // Milliseconds.
		const initial = Date.now()
		let expected = initial + interval
		this.timerTimeout = setTimeout(() => countUpHandler(), interval)

		const countUpHandler = () => {
			const current = Date.now()
			const delta = current - expected
			if (delta > interval) {
				// Skip missing steps.
				expected += delta
			}
			expected += interval

			const ms = Math.floor(current - initial)
			const minutes = Math.trunc(ms / 1000 / 60)
			const seconds = Math.trunc(ms / 1000) % 60

			const mins = `${minutes > 9 ? minutes : "0" + minutes.toString()}`
			const secs = `${seconds > 9 ? seconds : "0" + seconds.toString()}`
			this.timer.textContent = `${mins}:${secs}`

			this.timerTimeout = setTimeout(() => countUpHandler(), Math.max(0, interval - delta))
		}
	}

	displayCounter(encoded) {
		const cnt = decodeInt(encoded)
		this.playersNumber.textContent = `Players in queue: ${cnt}`

		if (cnt === 1) {
			window.alert("Queue is emty. Try selecting more time controls to find a game faster.")
		}
	}

	/**
	 * Displays random hint every 5 seconds.
	 */
	displayHint() {
		const i = Math.round(Math.random() * (hints.length - 1))
		for (const ind of this.hintsStack) {
			if (i === ind) {
				this.displayHint()
				return
			}
		}

		this.hintsStack.push(i)
		if (this.hintsStack.length === 12) {
			this.hintsStack.splice(0, 1)
		}

		// Toggle animation to apply smooth text change.
		this.hintBox.classList.add("hide")
		this.hintTimeout = setTimeout(() => {
			this.hintBox.textContent = hints[i]
			this.hintBox.classList.remove("hide")
		}, 750)
		this.hintTimeout = setTimeout(() => this.displayHint(), 5000)
	}

}

const slider = document.getElementById("slider")
// Set default slider thumb image.
const sliderDemo = document.getElementById("sliderDemo")
sliderDemo.textContent = "Fast Blitz with Increment"
slider.addEventListener("input", (e) => {
	switch (e.target.value) {
		case "0":
			sliderDemo.textContent = "Ultra bullet"
			/*
slider.style.setProperty(
				"--slider-thumb-image",
				`url("/assets/images/ultra-bullet.svg")`,
			)
*/

			break
		case "1":
			sliderDemo.textContent = "Bullet"
			/*
slider.style.setProperty(
				"--slider-thumb-image",
				`url("/assets/images/bullet.svg")`,
			)
*/

			break
		case "2":
			sliderDemo.textContent = "Fast Blitz"
/*
			slider.style.setProperty(
				"--slider-thumb-image",
				`url("/assets/images/fast-blitz.svg")`,
			)
*/

			break
		case "3":
			sliderDemo.textContent = "Fast Blitz with Increment"
/*
			slider.style.setProperty(
				"--slider-thumb-image",
				`url("/assets/images/fast-blitz-with-increment.svg")`,
			)
*/

			break
		case "4":
			sliderDemo.textContent = "Blitz"
/*
			slider.style.setProperty(
				"--slider-thumb-image",
				`url("/assets/images/blitz.svg")`,
			)
*/

			break
		case "5":
			sliderDemo.textContent = "Blitz with Increment"
/*
			slider.style.setProperty(
				"--slider-thumb-image",
				`url("/assets/images/blitz-with-increment.svg")`,
			)
*/

			break
		case "6":
			sliderDemo.textContent = "Rapid"
/*
			slider.style.setProperty(
				"--slider-thumb-image",
				`url("/assets/images/rapid.svg")`,
			)
*/

			break
		case "7":
			sliderDemo.textContent = "Rapid with Increment"
/*
			slider.style.setProperty(
				"--slider-thumb-image",
				`url("/assets/images/rapid-with-increment.svg")`,
			)
*/

			break
		case "8":
			sliderDemo.textContent = "Time to think"
/*
			slider.style.setProperty(
				"--slider-thumb-image",
				`url("/assets/images/time-to-think.svg")`,
			)
*/

			break
	}
})

const dialog = document.getElementById("matchmaking")
const state = new State()

document.getElementById("joinMatchmaking").onclick = () => {
	dialog.showModal()
	state.set()
}
document.getElementById("dialogClose").onclick = () => {
	dialog.close()

}
dialog.onclose = () => {
	state.reset()
}