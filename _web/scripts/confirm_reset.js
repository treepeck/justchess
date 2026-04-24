import { g } from "./utils/dom"
import MessageSystem from "./components/message"
;(() => {
	const parts = window.location.pathname.split("/")
	if (parts.length < 2 || parts[1] != "confirm-reset") return

	const params = new URLSearchParams(window.location.search)
	const token = params.get("token")

	const system = new MessageSystem()
	g("confirm").onclick = async () => {
		const res = await fetch(`/auth/confirm-reset/${token}`, {
			method: "POST",
			credentials: "include",
		})

		if (!res.ok) {
			system.create(await res.text())
		} else {
			window.location.href = "/signin"
		}
	}
})()