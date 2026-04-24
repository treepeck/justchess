import { g } from "./utils/dom"
import { validateInput } from "./signin"

/** @param {SubmitEvent} event */
async function submitForm(event) {
	event.preventDefault()
	event.stopPropagation()

	if (!(event.target instanceof HTMLFormElement)) return

	// Clear previous error message.
	const resMessage = g("serverResponse")
	resMessage.textContent = ""

	const data = new FormData(event.target)
	// @ts-expect-error - Works as expected, TypeScipt sometimes complains too much.
	const params = new URLSearchParams(data)

	const email = data.get("email")
	const password = data.get("password")
	if (email == null || password == null) return

	if (!validateInput(email.toString(), password.toString())) return

	// Disable the button while the request is being processed.
	const btn = /** @type {HTMLButtonElement} */ (g("formSubmitButton"))
	btn.disabled = true
	btn.textContent = "Submitting..."

	const res = await fetch("/auth/reset-password", {
		method: "POST",
		body: params,
		headers: { "Content-Type": "application/x-www-form-urlencoded" },
	})

	if (!res.ok) {
		resMessage.textContent = "Reset failed: " + (await res.text())
		resMessage.style.color = "red"
		// Reenable the submit button.
		btn.disabled = false
		btn.textContent = "Confirm"
		return
	}

	resMessage.textContent =
		"Please, check your email to confirm the reset. It may take several minutes for the email to be delivered and it may end up in spam."
	resMessage.style.color = "green"
	btn.style.display = "none"
}

;(() => {
	if (window.location.pathname != "/reset-password") return

	g("resetPasswordForm").onsubmit = submitForm

	const toggle = g("formPasswordToggle")
	toggle.onclick = () => {
		const input = g("formPasswordInput")

		const curr = input.getAttribute("type")
		if (curr === "password") {
			input.setAttribute("type", "text")
			toggle.style.backgroundImage = "url('/images/hide.svg')"
		} else {
			input.setAttribute("type", "password")
			toggle.style.backgroundImage = "url('/images/show.svg')"
		}
	}
})()
