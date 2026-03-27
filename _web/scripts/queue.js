import { g } from "./utils/dom"
import Socket from "./ws/socket"
import { EventKind } from "./ws/event"
import showDialog from "./utils/dialog"
import { formatTime } from "./components/clock"

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

/**
 * Displays random hint every 5 seconds.
 * @param {number[]} prev - To prevent repetitions.
 */
function showHint(prev) {
	const i = Math.round(Math.random() * (hints.length - 1))
	for (const ind of prev) {
		if (i === ind) {
			showHint(prev)
			return
		}
	}

	prev.push(i)
	if (prev.length === 12) {
		prev.splice(0, 1)
	}

	// Toggle animation to apply smooth text change.
	const hint = g("hint")
	hint.classList.add("hide")
	setTimeout(() => {
		hint.textContent = hints[i]
		hint.classList.remove("hide")
	}, 750)

	setTimeout(() => showHint(prev), 5000)
}

/** @type {import("./ws/socket").EventHandler} */
function eventHandler(Kind, payload) {
	switch (Kind) {
		case EventKind.ClientsCounter:
			// Update clients counter.
			g("playersNumber").textContent = `Players in queue: ${payload}`

			if (payload < 2) {
				g("emptyDialogPlayVsEngine").onclick = async () => {
					const res = await fetch("/play-vs-engine", {
						method: "POST",
						credentials: "include",
					})

					if (!res) {
						throw new Error("Couldn't create an engine game")
					}

					if (res.redirected) window.location.href = res.url
				}
				showDialog("emptyDialog")
			}
			break

		case EventKind.Redirect:
			// Redirect to game room.
			// @ts-expect-error - API_URL comes from webpack.
			window.location.href = `${API_URL}/${payload}`
			break

		default:
			throw new Error("Invalid event from server")
	}
}

;(() => {
	const parts = window.location.pathname.split("/")
	if (parts.length < 3 || parts[1] != "queue") return

	showHint([])

	new Socket(eventHandler)

	// Self-adjusting countup timer.
	const interval = 1000 // Milliseconds.
	const initial = Date.now()
	let expected = initial + interval
	setTimeout(() => countUpHandler(), interval)

	const countUpHandler = () => {
		const current = Date.now()
		const delta = current - expected
		if (delta > interval) {
			// Skip missing steps.
			expected += delta
		}
		expected += interval
		formatTime("countUpTimer", Math.floor(current - initial))

		setTimeout(() => countUpHandler(), Math.max(0, interval - delta))
	}
})()
