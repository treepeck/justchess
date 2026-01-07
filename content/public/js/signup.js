import { signUp } from "/public/js/http.js"

const nameEx = /^[a-zA-Z0-9]{2,60}$/i
const emailEx = /^[a-zA-Z0-9._]+@[a-zA-Z0-9._]+\.[a-zA-Z0-9._]+$/i
const pwdEx = /^[a-zA-Z0-9!@#$%^&*()_+-/.<>]{5,71}$/i

// Validates the user input and displays error messages.
function validateInput(name, email, password) {
	let isValid = true

	if (name.length < 2) {
		displayError("name-error", "Must be at least 2 characters long")
		isValid = false
	} else if (name.length > 60) {
		displayError("name-error", "Must not exceed 60 characters")
		isValid = false
	} else if (!nameEx.test(name)) {
		displayError("name-error", "Can only contain letters and numbers")
		isValid = false
	} else {
		displayError("name-error", "")
	}

	if (email.length < 3) {
		displayError("email-error", "Must be at least 3 characters long")
		isValid = false
	} else if (!emailEx.test(email)) {
		displayError("email-error", "Please, enter a valid email address")
		isValid = false
	} else {
		displayError("email-error", "")
	}

	if (password.length < 5) {
		displayError("password-error", "Must be at least 5 characters long")
		isValid = false
	} else if (password.length > 71) {
		displayError("password-error", "Must not exceed 71 characters")
		isValid = false
	} else if (!pwdEx.test(password)) {
		displayError("password-error", "Can only contain letters, numbers, and !@#$%^&*()_+-/.<>")
		isValid = false
	} else {
		displayError("password-error", "")
	}

	return isValid
}

function displayError(containerName, msg) {
	const container = document.getElementById(containerName)
	container.textContent = msg
}

function submitForm(e) {
	e.preventDefault()
	e.stopPropagation()

	// Clear previous error message.
	const container = document.getElementById("server-error")
	container.textContent = ""

	const data = new FormData(form)

	const name = data.get("name")
	const email = data.get("email")
	const password = data.get("password")

	if (validateInput(name, email, password)) {
		// Disable the button while the request is being processed.
		const button = document.getElementById("form-submit")
		button.disabled = true
		button.textContent = "Submitting..."

		const params = new URLSearchParams(data)

		signUp(params)
			.then((err) => {
				container.textContent = "Sign up failed: " + err

				// Enable the submit button.
				button.disabled = false
				button.textContent = "Sign up"
			})
	}
}

function togglePassword() {
	const curr = passwordInput.getAttribute("type")
	if (curr === "password") {
		passwordInput.setAttribute("type", "text")
		passwordToggle.style.backgroundImage = "url('/public/img/hide.png')"
	} else {
		passwordInput.setAttribute("type", "password")
		passwordToggle.style.backgroundImage = "url('/public/img/show.png')"
	}
}

const form = document.getElementById("form")
form.addEventListener("submit", submitForm)

const passwordToggle = document.getElementById("password-toggle")
passwordToggle.addEventListener("click", togglePassword)

const passwordInput = document.getElementById("input-password")

function toggleTooltipVisibility() {
	tooltipContainer.classList.toggle("tooltip-show")
}

const tooltipToggle = document.getElementById("tooltip-toggle")
tooltipToggle.addEventListener("click", toggleTooltipVisibility)

const tooltipContainer = document.getElementById("tooltip-container")
tooltipContainer.addEventListener("click", toggleTooltipVisibility)