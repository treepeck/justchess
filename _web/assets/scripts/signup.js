import {
	validateName,
	validateEmail,
	validatePassword,
} from "/assets/scripts/widget/form.js"

async function submitForm() {
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
		confirm.disabled = true
		confirm.textContent = "Processing..."

		const resMessage = document.getElementById("serverResponse")

		const params = new URLSearchParams(data)

		const res = await fetch("/auth/signup", {
			method: "POST",
			body: params,
			headers: { "Content-Type": "application/x-www-form-urlencoded" },
		})

		if (!res.ok) {
			resMessage.textContent = "Sign up failed: " + res.statusText
			resMessage.style.color = "red"
			// Reenable the confirm button.
			confirm.disabled = false
			confirm.textContent = "Confirm"
			return
		}

		resMessage.textContent =
			"Please, check your email to confirm the registration. It may take several minutes for the email to be delivered and it may end up in spam."
		resMessage.style.color = "green"
		confirm.style.display = "none"
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

const confirmDialog = document.getElementById("formConfirmation")

const confirm = document.getElementById("confirm")
confirm.onclick = (e) => {
	e.preventDefault()
	e.stopPropagation()

	const form = document.getElementById("form")
	const data = new FormData(form)
	const name = data.get("username")
	const email = data.get("email")
	const password = data.get("password")

	// Do not show the confimation dialog if input is invalid.
	if (!validateInput(name, email, password)) return

	confirmDialog.showModal()

	document.getElementById("dialogClose").onclick = () => {
		confirmDialog.close()
	}
	document.getElementById("dialogOK").onclick = async () => {
		confirmDialog.close()
		await submitForm()
	}
}

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
