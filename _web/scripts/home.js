import { getElement } from "./utils/dom"
import showHelpDialog from "./utils/help_dialog"
;(() => {
	// Page guard.
	const container = document.getElementById("mainContainer")
	if (!container || container.dataset.page !== "home") {
		return
	}

	getElement("timeControlHelpDialogActivator").onclick = () => {
		showHelpDialog("timeControlHelpDialog")
	}

	for (let i = 1; i <= 9; i++) {
		getElement(`cell${i}`).onclick = () => {
			// Redirect the user to the queue page.
			//@ts-expect-error - API_URL comes from webpack.
			window.location.href = `${API_URL}/queue/${i}`
		}
	}
})()
