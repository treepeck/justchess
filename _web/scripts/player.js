import { getOrPanic } from "./utils/dom"
import Notification from "./utils/notification"
import { formatTermination, formatResult, Result } from "./chess/state"

/**
 * @typedef {Object} ProfileGame
 * @property {string} c - Created at timestamp.
 * @property {string} i - Game id.
 * @property {string} w - White player id.
 * @property {string} b - Black player id.
 * @property {import("./chess/state").Result} r - Game result.
 * @property {import("./chess/state").Termination} t - Game termination.
 * @property {number} m - How many moves were played.
 * @property {number} ctl - Time control.
 * @property {number} bns - Time bonus.
 */

/**
 * @param {string} name
 * @param {string} [cursorId]
 * @param {string} [cursorCreatedAt]
 * @returns {Promise<ProfileGame[] | string>}
 */
async function getProfileGames(name, cursorId, cursorCreatedAt) {
	try {
		let url = `/api/profile-games?name=${name}`
		if (cursorId && cursorCreatedAt) {
			url += `&cid=${cursorId}&cat=${cursorCreatedAt}`
		}
		const res = await fetch(url, {
			method: "GET",
			credentials: "include",
		})
		return await res.json()
	} catch (err) {
		return err.message
	}
}

/** @param {ProfileGame} game */
function appendGameToTable(game) {
	const row = document.createElement("a")
	row.classList.add("profile-games-row")
	row.href = `/game/${game.i}`

	const res = document.createElement("div")
	res.innerHTML = `${formatResult(game.r)}<br/>${formatTermination(game.t)}`

	// If player won the match set result color to green.
	// If player lost the match set color to red.
	const name = getOrPanic("profileGames").dataset.name
	if (
		(game.r == Result.WhiteWon && game.w == name) ||
		(game.r == Result.BlackWon && game.b == name)
	) {
		res.style.color = "#66FF00"
	} else {
		res.style.color = "red"
	}
	row.appendChild(res)

	const players = document.createElement("div")
	players.classList.add("profile-games-players")
	const white = document.createElement("a")
	white.href = game.w
	white.classList.add("white-player")
	white.textContent = game.w
	const black = document.createElement("a")
	black.href = game.b
	black.classList.add("black-player")
	black.textContent = game.b
	players.appendChild(white)
	players.appendChild(black)
	row.appendChild(players)

	const control = document.createElement("div")
	control.classList.add("profile-games-time-control")
	control.textContent = `${game.ctl} + ${game.bns}`
	row.appendChild(control)

	const moves = document.createElement("div")
	moves.textContent = `${Math.floor(game.m / 2)}`
	row.appendChild(moves)

	const playedAt = document.createElement("div")
	playedAt.textContent = new Date(game.c).toLocaleDateString("en-US", {
		month: "short",
		day: "2-digit",
		year: "numeric",
	})
	row.appendChild(playedAt)

	getOrPanic("profileGames").appendChild(row)
}

await (async () => {
	// Page guard.
	if (!document.getElementById("playerGuard")) return

	// Format regiration date.
	const registeredAt = getOrPanic("playerRegistrationDate")
	if (!registeredAt.textContent) return
	const d = new Date(registeredAt.textContent)
	registeredAt.innerHTML = `Member since<br/>${d.toLocaleDateString("en-US", {
		month: "short",
		day: "2-digit",
		year: "numeric",
	})}`

	let cursorId = ""
	let cursorCreatedAt = ""

	// Render game history.
	const table = getOrPanic("profileGames")

	const games = await getProfileGames(table.dataset.name)
	if (typeof games == "string" || games.length == 0) {
		table.textContent =
			"Start playing and the game history will be displayed here"
		return
	}

	const notification = new Notification()

	// Render table header.
	const h = document.createElement("div")
	h.classList.add("profile-games-header")

	const c1 = document.createElement("div")
	h.appendChild(c1)

	const c2 = document.createElement("div")
	c2.textContent = "Players"
	h.appendChild(c2)

	const c3 = document.createElement("div")
	c3.textContent = "Time control"
	h.appendChild(c3)

	const c4 = document.createElement("div")
	c4.textContent = "Moves"
	h.appendChild(c4)

	const c5 = document.createElement("div")
	c5.textContent = "Date"
	h.appendChild(c5)

	table.appendChild(h)

	// Render games.
	for (const game of games) {
		appendGameToTable(game)
	}
	// Update pagination cursors.
	const last = games[games.length - 1]
	cursorId = last.i
	cursorCreatedAt = last.c

	// Render "Load more" button.
	const btn = document.createElement("button")
	btn.classList.add("profile-games-load-more")
	btn.textContent = "Load more"

	btn.onclick = async () => {
		const older = await getProfileGames(
			table.dataset.name,
			cursorId,
			cursorCreatedAt,
		)
		if (typeof older == "string" || older.length == 0) {
			notification.create("Couldn't load more games")
			return
		}
		for (const game of older) {
			appendGameToTable(game)
		}
		// Update pagination cursors.
		const last = older[older.length - 1]
		cursorId = last.i
		cursorCreatedAt = last.c

		// Rerender "Load more" button.
		table.removeChild(btn)
		table.appendChild(btn)
	}

	table.appendChild(btn)
})()
