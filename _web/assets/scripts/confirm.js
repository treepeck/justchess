import { get } from "/assets/scripts/dom.js"
import { createPopup } from "/assets/scripts/popup.js"

const btn = /** @type {HTMLButtonElement} */ (get("confirm"))

btn.onclick = async () => {
	const token = new URLSearchParams(window.location.search).get("token")
	const url = `/auth/${window.location.pathname.split("/")[1]}/${token}`

	const res = await fetch(url, {
		method: "POST",
		credentials: "include",
	})

	// Disable confirmation button after click.
	btn.disabled = true

	if (!res.ok) {
		createPopup(await res.text())
		return
	}
	window.location.href = res.url
}
