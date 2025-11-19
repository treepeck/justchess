import { signIn } from "/public/js/http.js"

const form = document.getElementById("form")
form.addEventListener("submit", submitHandler)

function submitHandler(e) {
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