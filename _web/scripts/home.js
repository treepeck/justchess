import { g } from "./utils/dom"
import showDialog from "./utils/dialog"
import MessageSystem from "./components/message"
;(() => {
	if (window.location.pathname != "/") return

	for (let i = 0; i < 9; i++) {
		g(`cell${i}`).onclick = () => {
			// Redirect the user to the queue page.
			//@ts-expect-error - API_URL comes from webpack.
			window.location.href = `${API_URL}/queue/${i}`
		}
	}

	const system = new MessageSystem()
	if (g("main").dataset.isguest === "true") {
		system.create("Guest players can only Play vs Engine")
	}

	g("playVsEngine").onclick = () => {
		showDialog("engineDialog")
		g("engineDialogConfirm").onclick = async () => {
			const difficulty = document.querySelector(
				`input[name="difficulty"]:checked`,
			)
			if (!difficulty) throw new Error("engine difficulty not set")

			const res = await fetch(`/play-vs-engine`, {
				method: "POST",
				body: difficulty.value,
				credentials: "include",
			})
			if (!res) {
				throw new Error("Couldn't create an engine game")
			}
			if (res.redirected) window.location.href = res.url
		}
	}

	g("helpText").onclick = () => showDialog("helpDialog")
})()
