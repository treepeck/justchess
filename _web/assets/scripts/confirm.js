import { popup } from "/assets/scripts/widget/popup.js"

async function confirmSignup() {
	const params = new URLSearchParams(document.location.search)
	const token = params.get("token")

	const res = await fetch(`/auth/confirm-signup/${token}`, {
		method: "POST",
		credentials: "include",
	})
	if (!res.ok) {
		popup("Token not found")
		return
	}
	// Redirect player to home page after successful confirmation.
	window.location.href = "/"
}

const confirm = document.getElementById("confirm")
confirm.onclick = confirmSignup
