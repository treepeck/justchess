import HelpWindow from "./utils/help"
;(() => {
	// Page guard.
	const container = document.getElementById("container")
	if (!container || container.dataset.page !== "home") {
		return
	}

	helpText.onclick = () => {
		HelpWindow.show("help")
	}

	for (let i = 1; i <= 9; i++) {
		const cell = document.getElementById(`cell${i}`)
		cell.addEventListener("click", () => {
			// Redirect the user to the queue page.
			window.location.href = `http://localhost:3502/queue/${i}`
		})
	}
})()
