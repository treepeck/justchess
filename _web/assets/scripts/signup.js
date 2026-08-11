import {
	validateName,
	validateEmail,
	validatePassword,
} from "/assets/scripts/widget/form.js"

async function submitForm() {
	// Close the confirmation dialog.
	const dialog = document.getElementById("confirmDialog")
	dialog.close()

	// Clear previous error message.
	const resp = document.getElementById("serverResponse")
	resp.textContent = ""

	const form = document.getElementById("form")

	const data = new FormData(form)

	const name = data.get("username")
	const email = data.get("email")
	const password = data.get("password")

	if (validateInput(name, email, password)) {
		// Disable the submit button.
		submit.disabled = true
		submit.textContent = "Processing..."

		const resMessage = document.getElementById("serverResponse")

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
			submit.disabled = false
			submit.textContent = "Confirm via email"
			return
		}

		resMessage.textContent =
			"Please, check your email to confirm the registration. It may take several minutes for the email to be delivered and it may end up in spam."
		resMessage.style.color = "green"
		submit.style.display = "none"
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

	let error = document.getElementById("formUsernameError")
	try {
		validateName(name)
		// Clear error message if name is valid.
		error.textContent = ""
	} catch (msg) {
		isValid = false
		error.textContent = msg
	}

	error = document.getElementById("formEmailError")
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

const submit = document.getElementById("submit")

const confirm = document.getElementById("formConfirm")
confirm.onclick = submitForm

const toggle = document.getElementById("formPasswordToggle")
toggle.onclick = () => {
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
