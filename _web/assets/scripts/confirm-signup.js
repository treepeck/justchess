async function confirmSignup() {
	const params = new URLSearchParams(document.location.search)
	const token = params.get("token")

	const res = await fetch(`/auth/confirm-signup/${token}`, {
		method: "POST",
		credentials: "include",
	})
	if (!res.ok) {
		errorDialog.showModal()
		return
	}
	// Redirect player to home page after successful confirmation.
	window.location.href = "/"
}

const confirm = document.getElementById("confirm")
confirm.onclick = confirmSignup

const errorDialog = document.getElementById("error")

const close = (document.getElementById("dialogClose").onclick = (e) => {
	errorDialog.close()
})
