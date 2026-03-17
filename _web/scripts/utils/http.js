/**
 * It's the caller's responsibility to validate the provided data and display errors.
 * @param {string} url
 * @param {URLSearchParams} data
 * @returns {Promise<string | undefined>}
 */
export async function formRequest(url, data) {
	try {
		const res = await fetch(url, {
			method: "POST",
			credentials: "include",
			body: data,
			headers: { "Content-Type": "application/x-www-form-urlencoded" },
		})

		if (!res.ok) {
			return await res.text()
		}
	} catch (err) {
		// @ts-expect-error
		return err.message
	}
}

/**
 * @param {string} url
 * @param {string} method
 * @param {any} data
 * @returns {Promise<string | undefined>}
 */
export async function request(url, method, data) {
	try {
		const res = await fetch(url, {
			method: method,
			credentials: "include",
			body: data,
			headers: { "Content-Type": "application/json" },
		})

		if (!res.ok) {
			return await res.text()
		}

		if (res.redirected) {
			window.location.href = res.url
		}
	} catch (err) {
		// @ts-expect-error
		return err.message
	}
}
