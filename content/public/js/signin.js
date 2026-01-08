import { signIn } from "/public/js/http.js"

function submitForm(e) {
	e.preventDefault()

	// Clear previous error message.
	serverError.textContent = ""

	// Disable the button while the request is being processed.
	submitBtn.disabled = true
	submitBtn.textContent = "Submitting..."

	const data = new FormData(authForm)
	const params = new URLSearchParams(data)

	signIn(params)
		.then((err) => {
			serverError.textContent = "Sign in failed: " + err

			// Enable the submit button.
			submitBtn.disabled = false
			submitBtn.textContent = "Sign in"
		})
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

authForm.addEventListener("submit", submitForm)

passwordToggle.addEventListener("click", togglePassword)