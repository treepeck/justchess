import { getOrPanic } from "./utils/dom"
import showDialog from "./utils/dialog"
import { request } from "./utils/http"
import {
	validateName,
	validatePassword,
	validateEmail,
} from "./utils/validator"

/** @param {SubmitEvent} event */
function submitForm(event) {
	event.preventDefault()
	event.stopPropagation()

	if (!(event.target instanceof HTMLFormElement)) return

	// Clear previous error message.
	getOrPanic("formServerError").textContent = ""

	const data = new FormData(event.target)

	const name = data.get("name")
	const email = data.get("email")
	const password = data.get("password")
	if (name == null || email == null || password == null) return

	if (validateInput(name.toString(), email.toString(), password.toString())) {
		// Show confirmation window.
		getOrPanic("confirmDialog").classList.add("show")
		getOrPanic("confirmDialogCancelButton").focus()
	}
}

/** @param {FormData} data */
function confirmHandler(data) {
	// Hide confirmation window
	getOrPanic("confirmDialog").classList.remove("show")

	// Disable the button while the request is being processed.
	const btn = /** @type {HTMLButtonElement} */ (
		getOrPanic("formSubmitButton")
	)
	btn.disabled = true
	btn.textContent = "Submitting..."

	// @ts-expect-error - Works as expected, TypeScipt sometimes complains too much.
	const params = new URLSearchParams(data)

	request("/auth/signup", params).then((err) => {
		if (err) {
			getOrPanic("formServerError").textContent = "Sign up failed: " + err
			// Enable the submit button.
			btn.disabled = false
			btn.textContent = "Sign up"
		} else {
			getOrPanic("formServerError").textContent =
				"Please, check your email to confirm the registration. It may take several minutes for the email to be delivered and it may end up in spam."
			btn.textContent = "Done"
		}
	})
}

/**
 * Validates the user input and displays error messages.
 * @param {string} name
 * @param {string} email
 * @param {string} password
 * @returns {boolean}
 */
function validateInput(name, email, password) {
	let isValid = true

	let error = getOrPanic("formNameError")
	try {
		validateName(name)
		// Clear error message if name is valid.
		error.textContent = ""
	} catch (msg) {
		isValid = false
		error.textContent = msg
	}

	error = getOrPanic("formEmailError")
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

;(() => {
	// Page guard.
	if (!document.getElementById("signupGuard")) return

	const form = /** @type {HTMLFormElement} */ (getOrPanic("authForm"))
	form.onsubmit = submitForm

	getOrPanic("emailHelpDialogActivator").onclick = () =>
		showDialog("emailHelpDialog")

	getOrPanic("confirmDialogCancelButton").onclick = () => {
		getOrPanic("confirmDialog").classList.remove("show")
	}

	getOrPanic("confirmDialogConfirmButton").onclick = () => {
		const data = new FormData(form)
		confirmHandler(data)
	}

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
