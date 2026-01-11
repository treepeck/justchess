// Regular expressions to validate the user input.
const nameEx = /^[a-zA-Z0-9]{2,60}$/i
const emailEx = /^[a-zA-Z0-9._]+@[a-zA-Z0-9._]+\.[a-zA-Z0-9._]+$/i
const pwdEx = /^[a-zA-Z0-9!@#$%^&*()_+-/.<>]{5,71}$/i

function submitForm(event) {
	event.preventDefault()
	event.stopPropagation()

	// Clear previous error message.
	serverError.textContent = ""

	const data = new FormData(authForm)

	const name = data.get("name")
	const email = data.get("email")
	const password = data.get("password")

	if (validateInput(name, email, password)) {
		// Show confirmation window.
		confirmWindow.classList.add("show")
		cancelSubmit.focus()
	}
}

function confirmHandler() {
	// Hide confirmation window.
	confirmWindow.classList.remove("show")

	// Disable the button while the request is being processed.
	submitBtn.disabled = true
	submitBtn.textContent = "Submitting..."

	const params = new URLSearchParams(data)

	signUp(params)
		.then((err) => {
			serverError.textContent = "Sign up failed: " + err

			// Enable the submit button.
			submitBtn.disabled = false
			submitBtn.textContent = "Sign up"
		})
}

// Validates the user input and displays error messages.
function validateInput(name, email, password) {
	let isValid = true

	if (name.length < 2) {
		nameError.textContent = "Must be at least 2 characters long"
		isValid = false
	} else if (name.length > 60) {
		nameError.textContent = "Must not exceed 60 characters"
		isValid = false
	} else if (!nameEx.test(name)) {
		nameError.textContent = "Can only contain letters and numbers"
		isValid = false
	} else {
		// Clear error message.
		nameError.textContent = ""
	}

	if (email.length < 3) {
		emailError.textContent = "Must be at least 3 characters long"
		isValid = false
	} else if (!emailEx.test(email)) {
		emailError.textContent = "Please, enter a valid email address"
		isValid = false
	} else {
		// Clear error message.
		emailError.textContent = ""
	}

	if (password.length < 5) {
		passwordError.textContent = "Must be at least 5 characters long"
		isValid = false
	} else if (password.length > 71) {
		passwordError.textContent = "Must not exceed 71 characters"
		isValid = false
	} else if (!pwdEx.test(password)) {
		passwordError.textContent = "Can only contain letters, numbers, and !@#$%^&*()_+-/.<>"
		isValid = false
	} else {
		// Clear error message.
		passwordError.textContent = ""
	}

	return isValid
}

// It's the caller's responsibility to validate the provided data and display errors.
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

function togglePassword() {
	const curr = passwordInput.getAttribute("type")
	if (curr === "password") {
		passwordInput.setAttribute("type", "text")
		passwordToggle.style.backgroundImage = "url('/images/hide.png')"
	} else {
		passwordInput.setAttribute("type", "password")
		passwordToggle.style.backgroundImage = "url('/images/show.png')"
	}
}

authForm.addEventListener("submit", submitForm)
passwordToggle.addEventListener("click", togglePassword)

helpText.addEventListener("click", () => {
	helpWindow.classList.add("show")

	// Focus close button.
	closeHelp.focus()
})

closeHelp.addEventListener("click", () => { helpWindow.classList.remove("show") })

cancelSubmit.addEventListener("click", () => { confirmWindow.classList.remove("show") })
confirmSubmit.addEventListener("click", () => { confirmHandler() })