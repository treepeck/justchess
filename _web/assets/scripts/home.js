import "/assets/scripts/dialog.js"
import { get } from "/assets/scripts/dom.js"

// Add event handlers to grid buttons.
for (let i = 0; i < 9; i++) {
	get(`time-control-${i}`).onclick = () => {
		// Redirect player to queue.
		window.location.href = `/queue/${i}`
	}
}

get("play").onclick = async () => {
	const difficulty = /** @type {HTMLInputElement} */ (
		document.querySelector(`input[name="difficulty"]:checked`)
	)

	const res = await fetch("/play-vs-engine", {
		method: "POST",
		body: difficulty.value,
	})
	if (res.redirected) {
		window.location.href = res.url
	}
	// TODO: show error notification
}

get("help").onclick = () => {
	get("helpDialog").showModal()
}
