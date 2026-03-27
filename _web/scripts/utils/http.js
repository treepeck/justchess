/**
 * Calls to /play-vs-engine endpoint.
 * Redirects player to game page if request was successfull.
 */
export async function playVsEngine() {
	const res = await fetch("/play-vs-engine", {
		method: "POST",
		credentials: "include",
	})
}
