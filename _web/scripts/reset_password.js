import { request } from "./utils/http"
import { getOrPanic } from "./utils/dom"
import { validateEmail, validatePassword } from "./utils/validator"

/**
 * Validates the user input and displays error messages.
 * @param {string} email
 * @param {string} password
 * @returns {boolean}
 */
function validateInput(email, password) {
	let isValid = true

	let error = getOrPanic("formEmailError")
	try {
		validateEmail(email)
		// Clear error message if email is valid.
		error.textContent = ""
	} catch (msg) {
		isValid = false
		error.textContent = msg
	}

	error = getOrPanic("formPasswordError")
	try {
		validatePassword(password)
		// Clear error message if password is valid.
		error.textContent = ""
	} catch (msg) {
		isValid = false
		error.textContent = msg
	}

	return isValid
}

/** @param {SubmitEvent} event */
function submitForm(event) {
	event.preventDefault()
	event.stopPropagation()

	if (!(event.target instanceof HTMLFormElement)) return

	// Clear previous error message.
	getOrPanic("formServerError").textContent = ""

	const data = new FormData(event.target)

	const email = data.get("email")
	const password = data.get("password")
	if (email == null || password == null) return

	// @ts-expect-error - Works as expected, TypeScipt sometimes complains too much.
	const params = new URLSearchParams(data)

	if (validateInput(email.toString(), password.toString())) {
		// Disable the button while the request is being processed.
		const btn = /** @type {HTMLButtonElement} */ (
			getOrPanic("formSubmitButton")
		)
		btn.disabled = true
		btn.textContent = "Submitting..."
		request("/auth/reset-password", params).then((err) => {
			if (err) {
				getOrPanic("formServerError").textContent =
					"Password reset failed: " + err
			} else {
				getOrPanic("formServerError").textContent =
					"Please, check your email to confirm the registration. It may take several minutes for the email to be delivered and it may end up in spam."
				btn.textContent = "Done"
			}
		})
		// Reenable the submit button.
		btn.disabled = false
		btn.textContent = "Submit"
	}
}

;(() => {
	// Page guard.
	if (!document.getElementById("resetPasswordGuard")) return

	const form = /** @type {HTMLFormElement} */ (
		getOrPanic("resetPasswordForm")
	)
	form.onsubmit = submitForm

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
