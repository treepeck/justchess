import { g, c } from "./utils/dom"
import { Color } from "./components/clock"
import MessageSystem from "./components/message"
import {
	formatResult,
	formatTermination,
	Result,
	Termination,
} from "./utils/state"

/**
 * @typedef {Object} RatedGameBrief
 * @property {string} c - Creation date.
 * @property {string} i - Game id.
 * @property {string} w - White name.
 * @property {string} wi - White id.
 * @property {string} b - Black name.
 * @property {string} bi - Black id.
 * @property {Result} r
 * @property {Termination} t
 * @property {number} m - Number of played moves.
 * @property {number} ctl - Time control.
 * @property {number} bns - Time bonus.
 */

/**
 * @param {string} playerId
 * @param {string} [cursorId]
 * @param {string} [cursorCreatedAt]
 * @returns {Promise<RatedGameBrief[]>}
 */
async function getRatedGamesBrief(playerId, cursorId, cursorCreatedAt) {
	let url = `/api/rated?pid=${playerId}`
	if (cursorId && cursorCreatedAt) {
		url += `&cid=${cursorId}&cca=${cursorCreatedAt}`
	}

	const res = await fetch(url, { method: "GET" })
	if (!res.ok) return []
	return await res.json()
}

/**
 * @typedef {Object} EngineGameBrief
 * @property {string} c - Creation date.
 * @property {string} i - Game id.
 * @property {Color} pc - Player color.
 * @property {Result} r
 * @property {Termination} t
 * @property {number} m - Number of played moves.
 */

/**
 * @param {string} playerId
 * @param {string} [cursorId]
 * @param {string} [cursorCreatedAt]
 * @returns {Promise<EngineGameBrief[]>}
 */
async function getEngineGamesBrief(playerId, cursorId, cursorCreatedAt) {
	let url = `/api/engine?pid=${playerId}`
	if (cursorId && cursorCreatedAt) {
		url += `&cid=${cursorId}&cca=${cursorCreatedAt}`
	}

	const res = await fetch(url, { method: "GET" })
	if (!res.ok) return []
	return await res.json()
}

/**
 * @param {RatedGameBrief} brief
 */
function appendRatedGameBrief(brief) {
	const row = /** @type {HTMLAnchorElement} */ (c("a", "profile-games-row"))
	row.classList.add("a")
	row.href = `/rated/${brief.i}`

	const res = c("div")
	res.innerHTML = `${formatResult(brief.r)}<br/>${formatTermination(brief.t)}`

	// If player won the match set result color to green.
	// If player lost the match set color to red.
	const name = g("profileGames").dataset.name
	if (
		(brief.r == Result.WhiteWon && brief.w == name) ||
		(brief.r == Result.BlackWon && brief.b == name)
	) {
		res.style.color = "#66FF00"
	} else {
		res.style.color = "red"
	}
	row.appendChild(res)

	const players = c("div", "profile-games-players")
	row.appendChild(players)

	const white = /** @type {HTMLAnchorElement} */ (c("a", "white-player"))
	white.classList.add("a")
	white.href = `/player/${brief.wi}`
	white.textContent = brief.w

	const black = /** @type {HTMLAnchorElement} */ (c("a", "black-player"))
	black.classList.add("a")
	black.href = `/player/${brief.bi}`
	black.textContent = brief.b

	players.appendChild(white)
	players.appendChild(c("br"))
	players.appendChild(black)

	const control = c("div", "profile-games-time-control")
	control.textContent = `${brief.ctl / 60} + ${brief.bns}`
	row.appendChild(control)

	const moves = c("div")
	moves.textContent = `${Math.ceil(brief.m / 2)}`
	row.appendChild(moves)

	const playedAt = document.createElement("div")
	playedAt.textContent = new Date(brief.c).toLocaleDateString("en-US", {
		month: "short",
		day: "2-digit",
		year: "numeric",
	})
	row.appendChild(playedAt)

	g("tabPane1").appendChild(row)
}

/**
 * @param {EngineGameBrief} brief
 */
function appendEngineGameBrief(brief) {
	const row = /** @type {HTMLAnchorElement} */ (c("a", "profile-games-row"))
	row.classList.add("a")
	row.href = `/engine/${brief.i}`

	const res = c("div")
	res.innerHTML = `${formatResult(brief.r)}<br/>${formatTermination(brief.t)}`

	// If player won the match set result color to green.
	// If player lost the match set color to red.
	if (
		(brief.r == Result.WhiteWon && brief.pc == Color.White) ||
		(brief.r == Result.BlackWon && brief.pc == Color.Black)
	) {
		res.style.color = "#66FF00"
	} else {
		res.style.color = "red"
	}
	row.appendChild(res)

	const moves = c("div")
	moves.textContent = `${Math.ceil(brief.m / 2)}`
	row.appendChild(moves)

	const playedAt = document.createElement("div")
	playedAt.textContent = new Date(brief.c).toLocaleDateString("en-US", {
		month: "short",
		day: "2-digit",
		year: "numeric",
	})
	row.appendChild(playedAt)

	g("tabPane2").appendChild(row)
}

/** @param {number} num */
function activateTab(num) {
	for (const link of document.getElementsByClassName("tab-link")) {
		link.classList.remove("active")
	}
	for (const pane of document.getElementsByClassName("tab-pane")) {
		pane.classList.remove("active")
	}
	g(`tabLink${num}`).classList.add("active")
	g(`tabPane${num}`).classList.add("active")
}

;(() => {
	const parts = window.location.pathname.split("/")
	if (parts.length < 3 || parts[1] != "player") return

	const playerId = parts[2]

	let fetchedRatedGames = 0
	let ratedCursorId = /** @type {string | undefined} */ (undefined)
	let ratedCursorCreatedAt = /** @type {string | undefined } */ (undefined)

	let fetchedEngineGames = 0
	let engineCursorId = /** @type {string | undefined} */ (undefined)
	let engineCursorCreatedAt = /** @type {string | undefined } */ (undefined)

	// Format registration date.
	const registeredAt = g("playerRegistrationDate")
	registeredAt.innerHTML = `Member since<br/>${new Date(
		registeredAt.textContent,
	).toLocaleDateString("en-US", {
		month: "short",
		day: "2-digit",
		year: "numeric",
	})}`

	g("tabLink1").onclick = () => activateTab(1)
	g("tabLink2").onclick = () => {
		activateTab(2)
		if (fetchedEngineGames == 0) {
			appendEngine()
		}
	}

	const appendRated = () => {
		getRatedGamesBrief(playerId, ratedCursorId, ratedCursorCreatedAt).then(
			(games) => {
				if (fetchedRatedGames == 100) {
					g("tabPane1").removeChild(g("loadMoreRated"))
				}

				if (games.length < 1) {
					system.create("Couldn't load more games")
					return
				}

				fetchedRatedGames = 0
				for (const brief of games) {
					appendRatedGameBrief(brief)
					fetchedRatedGames++
				}

				// @ts-expect-error
				ratedCursorId = games.at(-1).i
				// @ts-expect-error
				ratedCursorCreatedAt = games.at(-1).c

				if (fetchedRatedGames == 100) {
					const loadMoreRated = c(
						"button",
						"profile-games-row",
						"loadMoreRated",
					)
					loadMoreRated.classList.add("button")
					loadMoreRated.textContent = "Load more"
					g("tabPane1").appendChild(loadMoreRated)

					loadMoreRated.onclick = appendRated
				}
			},
		)
	}
	appendRated()

	const appendEngine = () => {
		getEngineGamesBrief(
			playerId,
			engineCursorId,
			engineCursorCreatedAt,
		).then((games) => {
			if (fetchedEngineGames == 100) {
				g("tabPane2").removeChild(g("loadMoreEngine"))
			}

			if (games.length < 1) {
				system.create("Couldn't load more games")
				return
			}

			fetchedEngineGames = 0
			for (const brief of games) {
				appendEngineGameBrief(brief)
				fetchedEngineGames++
			}

			// @ts-expect-error
			engineCursorId = games.at(-1).i
			// @ts-expect-error
			engineCursorCreatedAt = games.at(-1).c

			if (fetchedEngineGames == 100) {
				const loadMoreEngine = c(
					"button",
					"profile-games-row",
					"loadMoreEngine",
				)
				loadMoreEngine.classList.add("button")
				loadMoreEngine.textContent = "Load more"
				g("tabPane2").appendChild(loadMoreEngine)

				loadMoreEngine.onclick = appendEngine
			}
		})
	}

	const system = new MessageSystem()
})()
