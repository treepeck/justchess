import { Piece } from "./enum.js"

const ACTION_MAKE_MOVE = 1

export class _WebSocket {
    // Establishes a WebSocket connection with the server.
	// The connection is stored in the socket field.
    constructor(handlers) {
        if (!window["WebSocket"]) {
            alert("WebSocket protocol is not supported")
            return
        }

        this.socket = new WebSocket("ws://" + document.location.host + "/websocket")

        // Event handlers.
        this.handlers = handlers

        this.socket.addEventListener("message", (e) => this.onMessage(e))
    }

    sendMakeMove(move) {
        this.socket.send(JSON.stringify({
            a: ACTION_MAKE_MOVE,
            p: JSON.stringify({
                t: move.to,
                f: move.from,
                p: Piece.NP
            }),
        }))
    }

    onMessage(e) {
        const msg = JSON.parse(e.data)
        const payload = JSON.parse(msg.p)
        // Call event handler.
        this.handlers.get(msg.a)(payload)
    }
}
