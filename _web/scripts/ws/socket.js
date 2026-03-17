import { getOrPanic } from "../utils/dom"
import Notification from "../utils/notification"
import { EventKind } from "./event"

/**
 * Function that handles player's events.
 * @callback EventHandler
 * @param {EventKind} Kind
 * @param {any} payload
 * @returns {void}
 */

/** Wrapper around the WebSocket object that encapsulates the repetitive code. */
export class Socket {
	/**
	 * @type {WebSocket}
	 */
	#socket

	/** @param {EventHandler} eventHandler */
	constructor(eventHandler) {
		// Get id.
		const path = window.location.pathname.split("/")
		if (path.length < 2) {
			console.error("Invalid pathname.")
			return
		}
		// Connect to the server.
		const id = path[path.length - 1]
		// @ts-expect-error - WS_URL comes from webpack.
		this.#socket = new WebSocket(`${WS_URL}/ws?id=${id}`)

		// Add event listeners.
		const notification = new Notification()
		this.#socket.onerror = () => {
			notification.create("Please, reload the page to reconnect.")
		}

		this.#socket.onmessage = (raw) => {
			/** @type {import("./event").Event} */
			const e = JSON.parse(raw.data)
			const Kind = e.k
			const payload = e.p

			switch (Kind) {
				// Respond with Pong automatically.
				case EventKind.Ping:
					this.#socket.send(
						JSON.stringify({ k: EventKind.Pong, p: null }),
					)

					// Update ping.
					getOrPanic("ping").textContent = `Ping: ${payload} ms`
					break

				// Something went wrong.  Display the notification and close
				// the connection.
				case EventKind.Error:
					notification.create(payload)
					this.#socket.close()
					break

				default:
					eventHandler(Kind, payload)
			}
		}
	}

	/**
	 * @param {EventKind} kind
	 * @param {any} payload
	 */
	sendJSON(kind, payload) {
		this.#socket.send(
			JSON.stringify({
				k: kind,
				p: payload,
			}),
		)
	}
}
