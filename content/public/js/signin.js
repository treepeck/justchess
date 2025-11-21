import { signIn } from "/public/js/http.js"

function submitForm(e) {
	e.preventDefault()

	/* Clear previous error message. */
	const container = document.getElementById("server-error")
	container.textContent = ""

	/* Disable the button while the request is being processed. */
	const button = document.getElementById("form-submit")
	button.disabled = true
	button.textContent = "Waiting..."

	const data = new FormData(form)
	const params = new URLSearchParams(data)

	signIn(params)
		.then((err) => {
			container.textContent = "Sign in failed: " + err

			/* Enable the submit button. */
			button.disabled = false
			button.textContent = "Sign in"
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

const form = document.getElementById("form")
form.addEventListener("submit", submitForm)

const passwordToggle = document.getElementById("password-toggle")
passwordToggle.addEventListener("click", togglePassword)

const passwordInput = document.getElementById("input-password")