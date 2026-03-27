import { g } from "../utils/dom"
import { EventKind } from "./event"
import MessageSystem from "../components/message"

/**
 * Function that handles player's events.
 * @callback EventHandler
 * @param {EventKind} Kind
 * @param {any} payload
 * @returns {void}
 */

/** Wrapper around the WebSocket object that encapsulates the repetitive code. */
export default class Socket {
	/**
	 * @type {WebSocket}
	 */
	#socket

	/** @param {EventHandler} eventHandler */
	constructor(eventHandler) {
		this.isConnected = false

		// Get id.
		const path = window.location.pathname.split("/")
		if (path.length < 2) throw new Error("Invalid pathname.")

		// Connect to the server.
		const id = path[path.length - 1]
		// @ts-expect-error - WS_URL comes from webpack.
		this.#socket = new WebSocket(`${WS_URL}/ws/${id}`)

		const system = new MessageSystem()
		this.#socket.onerror = () => {
			system.create("Please, reload the page to reconnect")
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
					g("ping").textContent = `Latency: ${payload} ms`
					break

				// Something went wrong.  Display the notification and close
				// the connection.
				case EventKind.Error:
					system.create(payload)
					this.#socket.close()
					this.isConnected = false
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
