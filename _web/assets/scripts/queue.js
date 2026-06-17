import { get } from "/assets/scripts/dom.js"

/**
 * Formats given milliseconds to a "mm:ss" string.
 * @param {number} ms
 * @returns {string}
 */
function formatTime(ms) {
	const minutes = Math.trunc(ms / 1000 / 60)
	const seconds = Math.trunc(ms / 1000) % 60

	let mins = `${minutes > 9 ? minutes : "0" + minutes.toString()}`
	let secs = `${seconds > 9 ? seconds : "0" + seconds.toString()}`
	return `${mins}:${secs}`
}

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

// Self-adjusting countup timer.
const interval = 1000 // Milliseconds.
const initial = Date.now()
let expected = initial + interval
setTimeout(() => countUp(), interval)

const countUp = () => {
	const current = Date.now()
	const delta = current - expected
	if (delta > interval) {
		// Skip missing steps.
		expected += delta
	}
	expected += interval
	get("countUpTimer").textContent = formatTime(Math.floor(current - initial))

	setTimeout(() => countUp(), Math.max(0, interval - delta))
}

// Show a random hint every 5 seconds.
const prev = []
const showHints = () => {
	const i = Math.round(Math.random() * (hints.length - 1))
	for (const ind of prev) {
		if (i === ind) {
			showHints()
			return
		}
	}

	prev.push(i)
	if (prev.length === 12) {
		prev.splice(0, 1)
	}

	// Toggle animation to apply smooth text change.
	const hint = get("hintBox")
	hint.classList.add("hide")
	setTimeout(() => {
		hint.textContent = hints[i]
		hint.classList.remove("hide")
	}, 750)

	setTimeout(() => showHints(prev), 5000)
}
showHints()
