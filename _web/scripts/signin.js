import { g } from "./utils/dom"
import { validateEmail, validatePassword } from "./utils/validator"

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

	const res = await fetch("/auth/signin", {
		method: "POST",
		body: params,
		headers: { "Content-Type": "application/x-www-form-urlencoded" },
	})

	if (!res.ok) {
		resMessage.textContent = "Sign in failed: " + (await res.text())
		resMessage.style.color = "red"
		// Reenable the submit button.
		btn.disabled = false
		btn.textContent = "Sign in"
		return
	}
	// Redirect user to home page after successful authentication.
	window.location.href = "/"
}

/**
 * Validates the user input. Displays error messages if it's invalid.
 * @param {string} email
 * @param {string} password
 * @returns {boolean}
 */
export function validateInput(email, password) {
	let isValid = true

	let error = g("formEmailError")
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

;(() => {
	if (window.location.pathname != "/signin") return

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
})()
