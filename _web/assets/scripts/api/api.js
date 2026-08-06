/**
 * @param {number} difficulty
 */
export async function createEngine(difficulty) {
	const res = await fetch(`/api/create-engine`, {
		method: "POST",
		body: JSON.stringify(difficulty),
		credentials: "include",
	})
	if (!res) {
		throw new Error("Couldn't create an engine game")
	}
	if (res.redirected) window.location.href = res.url
}
