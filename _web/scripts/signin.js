;(() => {
	// Page guard.
	const form = document.getElementById("authForm")
	if (!form) {
		return
	}

	form.addEventListener("submit", submitForm)
	passwordToggle.addEventListener("click", togglePassword)
})()

function submitForm(event) {
	event.preventDefault()
	event.stopPropagation()

	// Clear previous error message.
	serverError.textContent = ""

	// Disable the button while the request is being processed.
	submitBtn.disabled = true
	submitBtn.textContent = "Submitting..."

	const data = new FormData(authForm)
	const params = new URLSearchParams(data)

	signIn(params).then((err) => {
		serverError.textContent = "Sign in failed: " + err

		// Enable the submit button.
		submitBtn.disabled = false
		submitBtn.textContent = "Sign in"
	})
}

async function signIn(data) {
	try {
		const res = await fetch("/auth/signin", {
			method: "POST",
			credentials: "include",
			body: data,
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
