import HelpWindow from "./utils/help"
import { getElement } from "./utils/dom"

// Regular expressions to validate the user input.
const nameEx = /^[a-zA-Z0-9]{2,60}$/i
const emailEx = /^[a-zA-Z0-9._]+@[a-zA-Z0-9._]+\.[a-zA-Z0-9._]+$/i
const pwdEx = /^[a-zA-Z0-9!@#$%^&*()_+-/.<>]{5,71}$/i

/** @param {SubmitEvent} event */
function submitForm(event) {
	event.preventDefault()
	event.stopPropagation()

	if (!(event.target instanceof HTMLFormElement)) return

	// Clear previous error message.
	getElement("serverError").textContent = ""

	const data = new FormData(event.target)

	const name = data.get("name")
	const email = data.get("email")
	const password = data.get("password")
	if (!name || !email || !password) return

	if (validateInput(name.toString(), email.toString(), password.toString())) {
		// Show confirmation window.
		getElement("confirmWindow").classList.add("show")
		getElement("cancelSubmit").focus()
	}
}

/** @param {FormData} data */
function confirmHandler(data) {
	// Hide confirmation window
	getElement("confirmWindow").classList.remove("show")

	// Disable the button while the request is being processed.
	const btn = /** @type {HTMLButtonElement} */ (getElement("submitBtn"))
	btn.disabled = true
	btn.textContent = "Submitting..."

	// @ts-expect-error
	const params = new URLSearchParams(data)

	signUp(params).then((err) => {
		getElement("serverError").textContent = "Sign up failed: " + err

		// Enable the submit button.
		btn.disabled = false
		btn.textContent = "Sign up"
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

	let error = getElement("nameError")
	if (name.length < 2) {
		error.textContent = "Must be at least 2 characters long"
		isValid = false
	} else if (name.length > 60) {
		error.textContent = "Must not exceed 60 characters"
		isValid = false
	} else if (!nameEx.test(name)) {
		error.textContent = "Can only contain letters and numbers"
		isValid = false
	} else {
		// Clear error message.
		error.textContent = ""
	}

	error = getElement("emailError")
	if (email.length < 3) {
		error.textContent = "Must be at least 3 characters long"
		isValid = false
	} else if (!emailEx.test(email)) {
		error.textContent = "Please, enter a valid email address"
		isValid = false
	} else {
		// Clear error message.
		error.textContent = ""
	}

	error = getElement("passwordError")
	if (password.length < 5) {
		error.textContent = "Must be at least 5 characters long"
		isValid = false
	} else if (password.length > 71) {
		error.textContent = "Must not exceed 71 characters"
		isValid = false
	} else if (!pwdEx.test(password)) {
		error.textContent =
			"Can only contain letters, numbers, and !@#$%^&*()_+-/.<>"
		isValid = false
	} else {
		// Clear error message.
		error.textContent = ""
	}

	return isValid
}

/**
 * It's the caller's responsibility to validate the provided data and display errors.
 * @param {URLSearchParams} data
 */
async function signUp(data) {
	try {
		const res = await fetch("/auth/signup", {
			method: "POST",
			credentials: "include",
			body: data,
			headers: { "Content-Type": "application/x-www-form-urlencoded" },
		})

		if (!res.ok) {
			return await res.text()
		}

		// Redirect user to home page after successful authentication.
		window.location.href = "/"
	} catch (err) {
		return err.message
	}
}

;(() => {
	// Page guard.
	const form = /** @type {HTMLFormElement} */ (
		document.getElementById("authForm")
	)
	if (!form || form.dataset.page !== "signup") return

	form.onsubmit = submitForm

	getElement("helpText").onclick = () => {
		HelpWindow.show("help")
	}

	getElement("cancelSubmit").onclick = () => {
		getElement("confirmWindow").classList.remove("show")
	}

	getElement("confirmSubmit").onclick = () => {
		const data = new FormData(form)
		confirmHandler(data)
	}

	const toggle = getElement("passwordToggle")
	toggle.onclick = () => {
		const input = getElement("passwordInput")

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
