import { getOrPanic } from "./utils/dom"
import showDialog from "./utils/dialog"
import { request } from "./utils/http"
;(() => {
	// Page guard.
	if (!document.getElementById("homeGuard")) return

	getOrPanic("helpDialogActivator").onclick = () => {
		showDialog("helpDialog")
	}

	for (let i = 0; i < 9; i++) {
		getOrPanic(`cell${i}`).onclick = () => {
			// Redirect the user to the queue page.
			//@ts-expect-error - API_URL comes from webpack.
			window.location.href = `${API_URL}/queue/${i}`
		}
	}

	getOrPanic("playVsEngine").onclick = () => {
		request("/play-vs-engine", "POST", null)
	}
})()
