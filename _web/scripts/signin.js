import { request } from "./utils/http"
import { getOrPanic } from "./utils/dom"

/** @param {SubmitEvent} event */
function submitForm(event) {
	event.preventDefault()
	event.stopPropagation()

	if (!(event.target instanceof HTMLFormElement)) return

	// Clear previous error message.
	const error = getOrPanic("formServerError")
	error.textContent = ""

	// Disable the button while the request is being processed.
	const btn = /** @type {HTMLButtonElement} */ (
		getOrPanic("formSubmitButton")
	)
	btn.disabled = true
	btn.textContent = "Submitting..."

	const data = new FormData(event.target)
	// @ts-expect-error - Works as expected, TypeScipt sometimes complains too much.
	const params = new URLSearchParams(data)

	request("/auth/signin", params).then((err) => {
		if (!err) {
			// Redirect user to home page after successful authentication.
			window.location.href = "/"
		}
		error.textContent = "Sign in failed: " + err

		// Enable the submit button.
		btn.disabled = false
		btn.textContent = "Sign in"
	})
}

;(() => {
	// Page guard.
	if (!document.getElementById("signinGuard")) return

	getOrPanic("authForm").onsubmit = submitForm

	const toggle = getOrPanic("formPasswordToggle")
	toggle.onclick = () => {
		const input = getOrPanic("formPasswordInput")

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
