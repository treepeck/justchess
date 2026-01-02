
export async function signIn(data) {
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

// It's a caller's responsibility to validate the provided data to display errors.
export async function signUp(data) {
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

		// Redirect user to home page after successful registration.
		window.location.href = "/"
	} catch (err) {
		return err.message
	}
}
