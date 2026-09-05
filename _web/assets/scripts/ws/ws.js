const wsUri = "ws://localhost:3502/handshake"
const pingTick = 3000 // In milliseconds.

export class Client {
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

	constructor() {
		this.pingInterval = 0
		this.latency = 0
		this.pingTimestamp = 0
		this.isPingAnswered = true

		this.conn = new WebSocket(wsUri)

		this.conn.onopen = () => { this.ping() }

		this.conn.onclose = () => { this.cleanup() }

		this.conn.onmessage = (e) => { this.recieve(e) }

		this.conn.onerror = (e) => {
			console.log(e)
		}

		// Close the connection when the client leaves.
		window.onbeforeunload = () => {
			this.conn.close()
		}
	}

	ping() {
		this.pingInterval = setInterval(() => {
			if (this.isPingAnswered) {
				this.conn.send(JSON.stringify({
					k: 0,
					p: this.latency,
				}))
				this.pingTimestamp = Date.now()
				this.isPingAnswered = false
			}
		}, pingTick)
	}

	/**
	 * @param {MessageEvent} e
	 */
	recieve(e) {
		try {
			const msg = JSON.parse(e.data)

			switch (msg.k) {
				case 1:
					this.latency = Math.floor((Date.now() - this.pingTimestamp) / 2)
					this.isPingAnswered = true
					break
			}
		} catch (err) {
			console.log(err)
			window.alert("Invalid message recieved from server")
		}
	}

	cleanup() {
		if (this.pingInterval) {
			clearInterval(this.pingInterval)
		}
	}
}
