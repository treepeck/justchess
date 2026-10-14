import { MessageKind, encodeInt, MaxMessageLength  } from "/assets/scripts/ws/proto.js"

const wsUri = "ws://localhost:8888/handshake"
const pingTick = 3000 // In milliseconds.

const ping = document.getElementById("ping")

/**
 * @typedef {MessageHandler}
 * @parameter {Uint8Array} encoded
 */

export class Conn {
	/**
	 * @type {WebSocket}
	 */
	conn
	/**
	 * @type {number}
	 */
	pingInterval
	/**
	 * To calculate the latency.
	 * @type {number}
	 */
	pingTimestamp
	/**
	 * Network latency calculated based on ping-pong roundtrip.
	 * In milliseconds.
	 * @type {number}
	 */
	latency
	/**
	 * The next ping message should be sent only if the previous
	 * one was answered.
	 * @type {boolean}
	 */
	isPingAnswered
	/**
	 * @type {Map<import("/assets/scripts/ws/proto.js").MessageKind, MessageHandler>}
	 */
	handles

	/**
	 * Opens the WebSocket connection with the server.
	 * @parameter {string} id - Id of the endpoint.
	 */
	constructor(id) {
		this.pingInterval = 0
		this.latency = 0
		this.pingTimestamp = 0
		this.isPingAnswered = true
		this.handlers = new Map()

		this.conn = new WebSocket(`${wsUri}/${id}`)

		this.conn.binaryType = "arraybuffer"

		this.conn.onopen = () => {
			this.ping()
		}

		this.conn.onclose = () => {
			this.cleanup()
		}

		this.conn.onmessage = (e) => {
			this.recieve(e)
		}

		this.conn.onerror = (e) => {
			window.assert(e)
		}

		// Close the connection when the client leaves.
		window.onbeforeunload = () => {
			this.conn.close()
		}
	}

	ping() {
		this.pingInterval = setInterval(() => {
			if (this.isPingAnswered) {
				const lat = encodeInt(this.latency)
				const encoded = new Uint8Array(1 + lat.length)
				encoded[0] = MessageKind.KindPing
				for (let i = 0; i < lat.length; i++) {
					encoded[i + 1] = lat[i]
				}
				this.send(encoded)
				this.pingTimestamp = Date.now()
				this.isPingAnswered = false
			}
		}, pingTick)
	}

	pong() {
		this.latency = Math.floor((Date.now() - this.pingTimestamp) / 2)
		this.isPingAnswered = true
		ping.textContent = `Latency: ${this.latency} ms`
	}

	/**
	 * @param {MessageEvent} e
	 */
	recieve(e) {
		try {
			if (e.data.byteLength === 0) {
				this.pong()
				return
			}
			// e.data includes the message separator (last byte) that needs to be ignored when
			// decoding the payload.
			const view = new Uint8Array(e.data)
			// Pass encoded message to the handler EXCLUDING the MessageSeparator.
			console.log(view)
			this.handlers.get(view[0])(view.slice(1,view.length-1))
		} catch (err) {
			console.log(err)
			window.alert("Invalid message recieved from server")
		}
	}

	/**
	 * @param {Uint8Array} encoded
	 */
	send(encoded) {
		this.conn.send(encoded)
		console.log("message sent:", encoded)
	}

	cleanup() {
		if (this.pingInterval) {
			clearInterval(this.pingInterval)
		}
		this.conn.close()
	}
}
