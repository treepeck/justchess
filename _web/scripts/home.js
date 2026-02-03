import HelpWindow from "./utils/help"
;(() => {
	// Page guard.
	const container = document.getElementById("container")
	if (!container || container.dataset.page !== "home") {
		return
	}

	const helpText = document.getElementById("helpText")
	if (!helpText) {
		console.error("Missing help text.")
		return
	}

	helpText.onclick = () => {
		HelpWindow.show("help")
	}

	for (let i = 1; i <= 9; i++) {
		const cell = document.getElementById(`cell${i}`)
		if (!cell) {
			console.error("Missing button.")
			break
		}
		cell.addEventListener("click", () => {
			// Redirect the user to the queue page.
			//@ts-expect-error - API_URL comes from webpack.
			window.location.href = `${API_URL}/queue/${i}`
		})
	}
})()
