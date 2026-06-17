import { get } from "/assets/scripts/dom.js"

// Regular expressions to validate user input.
const nameEx = /^[a-zA-Z0-9]{2,60}$/i
const emailEx = /^[a-zA-Z0-9._]+@[a-zA-Z0-9._]+\.[a-zA-Z0-9._]+$/i
const pwdEx = /^[a-zA-Z0-9!@#$%^&*()_+-/.<>]{5,71}$/i

/**
 * @param {string} name
 * @throws {string} Will throw an error if name is not valid.
 */
function validateUsername(name) {
	if (name.length < 2) {
		throw new Error("Must be at least 2 characters long")
	} else if (name.length > 60) {
		throw new Error("Must not exceed 60 characters")
	} else if (!nameEx.test(name)) {
		throw new Error("Can only contain letters and numbers")
	}
}

/**
 * @param {string} email
 * @throws Will throw an error if email is not valid.
 */
function validateEmail(email) {
	if (email.length < 3) {
		throw new Error("Must be at least 3 characters long")
	} else if (!emailEx.test(email)) {
		throw new Error("Please, enter a valid email address")
	}
}

/**
 * @param {string} password
 * @throws Will throw an error if password is not valid.
 */
function validatePassword(password) {
	if (password.length < 5) {
		throw new Error("Must be at least 5 characters long")
	} else if (password.length > 71) {
		throw new Error("Must not exceed 71 characters")
	} else if (!pwdEx.test(password)) {
		throw new Error(
			"Can only contain letters, numbers, and !@#$%^&*()_+-/.<>",
		)
	}
}

/**
 * Validates the user input. Displays error messages if it's invalid.
 * @param {string} [username] Optional.
 * @param {string} email
 * @param {string} password
 * @returns {boolean}
 */
function validateInput(username, email, password) {
	let isValid = true
	let error = null

	if (username) {
		error = get("formUsernameError")
		try {
			validateUsername(username)
		} catch (msg) {
			isValid = false
			error.textContent = msg
		}
	}

	error = get("formEmailError")
	try {
		validateEmail(email)
		// Clear error message if email is valid.
		error.textContent = ""
	} catch (msg) {
		isValid = false
		error.textContent = msg
	}

	error = get("formPasswordError")
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

get("passwordToggle").onclick = () => {
	const input = /** @type {HTMLInputElement} */ (get("passwordInput"))
	if (input.type === "password") {
		input.type = "text"
	} else {
		input.type = "password"
	}
}

get("form").onsubmit = (e) => submitForm(e)

/** @param {SubmitEvent} e */
async function submitForm(e) {
	e.preventDefault()
	e.stopPropagation()

	// Clear previous error message.
	const data = new FormData(e.target)
	const params = new URLSearchParams(data)

	const username = data.get("username")
	const email = data.get("email")
	const password = data.get("password")

	if (
		!validateInput(
			username ? username.toString() : null,
			email.toString(),
			password.toString(),
		)
	) {
		return
	}

	// Disable the button while the request is being processed.
	const btn = /** @type {HTMLButtonElement} */ (get("submit"))
	btn.disabled = true
	btn.textContent = "Processing..."

	const responseBox = get("serverResponse")

	// Different request URL on different pages.
	let url = "/auth/signin"
	switch (window.location.pathname) {
		case "/signup":
			url = "/auth/signup"
			break

		case "/reset-password":
			url = "/auth/signup"
			break
	}

	const res = await fetch(url, {
		method: "POST",
		body: params,
		headers: {
			"Content-Type": "application/x-www-form-urlencoded",
		},
	})
	if (res.redirected) {
		window.location.href = res.url
		return
	}
	responseBox.textContent = await res.text()
	if (!res.ok) {
		responseBox.style.color = "red"
		// Reenable the submit button.
		btn.disabled = false
		btn.textContent = "Confirm via email"
	} else {
		responseBox.style.color = "green"
		btn.remove()
	}
}
