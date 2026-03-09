/**
 * It's the caller's responsibility to validate the provided data and display errors.
 * @param {string} url
 * @param {URLSearchParams} data
 */
export async function request(url, data) {
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
		return err.message
	}
}
