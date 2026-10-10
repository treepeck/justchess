import { validateEmail, validatePassword } from "/assets/scripts/widget/form.js"

async function submitForm(e) {
	e.preventDefault()
	e.stopPropagation()

	// Clear previous error message.
	const resp = document.getElementById("serverResponse")
	resp.textContent = ""

	const form = document.getElementById("form")
	const data = new FormData(form)

	const email = data.get("email")
	const password = data.get("password")

	if (validateInput(email, password)) {
		// Disable the submit button.
		confirm.disabled = true
		confirm.textContent = "Processing..."

		const resMessage = document.getElementById("serverResponse")

		const params = new URLSearchParams(data)

		const res = await fetch("/auth/signin", {
			method: "POST",
			body: params,
			headers: { "Content-Type": "application/x-www-form-urlencoded" },
		})

		if (!res.ok) {
			resMessage.textContent = "Sign in failed: " + res.statusText
			resMessage.style.color = "red"
			// Reenable the confirm button.
			confirm.disabled = false
			confirm.textContent = "Confirm"
			return
		}

		// Redirect player to home page after successfull sign in.
		window.location.href = "/"
	}
}

/**
 * Validates the user input. Displays error messages if it's invalid.
 * @param {string} email
 * @param {string} password
 * @returns {boolean}
 */
function validateInput(email, password) {
	let isValid = true

	let error = document.getElementById("formEmailError")
	try {
		validateEmail(email)
		// Clear error message if email is valid.
		error.textContent = ""
	} catch (msg) {
		isValid = false
		error.textContent = msg
	}

	error = document.getElementById("formPasswordError")
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

const confirm = document.getElementById("confirm")
confirm.onclick = submitForm

const toggle = document.getElementById("formPasswordToggle")
toggle.onclick = (e) => {
	e.preventDefault()
	e.stopPropagation()

	const input = document.getElementById("passwordInput")

	const curr = input.getAttribute("type")
	if (curr === "password") {
		input.setAttribute("type", "text")
		toggle.style.backgroundImage = `url("/assets/images/hide.svg")`
	} else {
		input.setAttribute("type", "password")
		toggle.style.backgroundImage = `url("/assets//images/show.svg")`
	}
}
