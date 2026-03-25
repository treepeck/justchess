import { g } from "./utils/dom"
import showDialog from "./utils/dialog"
import {
	validateName,
	validateEmail,
	validatePassword,
} from "./utils/validator"

/** @param {SubmitEvent} event */
async function submitForm(event) {
	event.preventDefault()
	event.stopPropagation()

	if (!(event.target instanceof HTMLFormElement)) return

	// Clear previous error message.
	g("serverResponse").textContent = ""

	const data = new FormData(event.target)

	const name = data.get("name")
	const email = data.get("email")
	const password = data.get("password")
	if (name == null || email == null || password == null) return

	if (validateInput(name.toString(), email.toString(), password.toString())) {
		showDialog("confirmDialog")
		g("confirmDialogConfirm").onclick = () => confirmSignup(data)
	}
}

/**
 * Validates the user input. Displays error messages if it's invalid.
 * @param {string} name
 * @param {string} email
 * @param {string} password
 * @returns {boolean}
 */
function validateInput(name, email, password) {
	let isValid = true

	let error = g("formNameError")
	try {
		validateName(name)
		// Clear error message if name is valid.
		error.textContent = ""
	} catch (msg) {
		isValid = false
		error.textContent = msg
	}

	error = g("formEmailError")
	try {
		validateEmail(email)
		// Clear error message if email is valid.
		error.textContent = ""
	} catch (msg) {
		isValid = false
		error.textContent = msg
	}

	error = g("formPasswordError")
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

/** @param {FormData} data */
async function confirmSignup(data) {
	// Hide confirmation window
	g("confirmDialog").classList.remove("show")

	// Disable the button while the request is being processed.
	const btn = /** @type {HTMLButtonElement} */ (g("formSubmitButton"))
	btn.disabled = true
	btn.textContent = "Submitting..."

	const resMessage = g("serverResponse")

	// @ts-expect-error - Works as expected, TypeScipt sometimes complains too much.
	const params = new URLSearchParams(data)

	const res = await fetch("/auth/signup", {
		method: "POST",
		body: params,
		headers: { "Content-Type": "application/x-www-form-urlencoded" },
	})

	if (!res.ok) {
		resMessage.textContent = "Sign up failed: " + (await res.text())
		resMessage.style.color = "red"
		// Reenable the submit button.
		btn.disabled = false
		btn.textContent = "Sign up"
		return
	}

	resMessage.textContent =
		"Please, check your email to confirm the registration. It may take several minutes for the email to be delivered and it may end up in spam."
	resMessage.style.color = "green"
	btn.style.display = "none"
}

;(() => {
	if (window.location.pathname != "/signup") return

	g("authForm").onsubmit = submitForm

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

	g("helpDialogActivator").onclick = () => showDialog("helpDialog")
})()
