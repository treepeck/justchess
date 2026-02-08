import { getElement } from "./utils/dom"

/** @param {SubmitEvent} event */
function submitForm(event) {
	event.preventDefault()
	event.stopPropagation()

	if (!(event.target instanceof HTMLFormElement)) return

	// Clear previous error message.
	const error = getElement("authFormServerError")
	error.textContent = ""

	// Disable the button while the request is being processed.
	const btn = /** @type {HTMLButtonElement} */ (
		getElement("authFormSubmitButton")
	)
	btn.disabled = true
	btn.textContent = "Submitting..."

	const data = new FormData(event.target)
	// @ts-expect-error
	const params = new URLSearchParams(data)

	signIn(params).then((err) => {
		error.textContent = "Sign in failed: " + err

		// Enable the submit button.
		btn.disabled = false
		btn.textContent = "Sign in"
	})
}

/** @param {URLSearchParams} data */
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

;(() => {
	// Page guard.
	const form = document.getElementById("authForm")
	if (!form || form.dataset.page !== "signin") return

	form.onsubmit = submitForm

	const toggle = getElement("authFormPasswordToggle")
	toggle.onclick = () => {
		const input = getElement("authFormPasswordInput")

		const curr = input.getAttribute("type")
		if (curr === "password") {
			input.setAttribute("type", "text")
			toggle.style.backgroundImage = "url('/images/hide.svg')"
		} else {
			input.setAttribute("type", "password")
			toggle.style.backgroundImage = "url('/images/show.svg')"
		}
	}
})()
