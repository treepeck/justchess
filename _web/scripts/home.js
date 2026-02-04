import { getElement } from "./utils/dom"
import HelpWindow from "./utils/help"
;(() => {
	// Page guard.
	const container = document.getElementById("container")
	if (!container || container.dataset.page !== "home") {
		return
	}

	getElement("helpText").onclick = () => {
		HelpWindow.show("help")
	}

	for (let i = 1; i <= 9; i++) {
		getElement(`cell${i}`).onclick = () => {
			// Redirect the user to the queue page.
			//@ts-expect-error - API_URL comes from webpack.
			window.location.href = `${API_URL}/queue/${i}`
		}
	}
})()
