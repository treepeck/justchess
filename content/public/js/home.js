import { WSEvent, Action } from "/public/js/types.js"

function handleEvent(e) {
	switch (e.action) {
		case Action.Ping:
			socket.send(new WSEvent(Action.Pong, null).toJSON())
			ping.textContent = "Ping: " + e.payload + " ms"
			break

		case Action.Redirect:
			window.location.href = `/game?id=${e.payload}`
			break
	}
}

function joinMatchmaking(cellInd) {
	const e = new WSEvent(Action.JoinMatchmaking, null)
	switch (cellInd) {
	case 1:
		e.payload = {tc: 60, tb: 0}
		break
	case 2:
		e.payload = {tc: 120, tb: 1}
		break
	case 3:
		e.payload = {tc: 180, tb: 0}
		break
	case 4:
		e.payload = {tc: 180, tb: 2}
		break
	case 5:
		e.payload = {tc: 300, tb: 0}
		break
	case 6:
		e.payload = {tc: 300, tb: 2}
		break
	case 7:
		e.payload = {tc: 600, tb: 0}
		break
	case 8:
		e.payload = {tc: 600, tb: 10}
		break
	default:
		e.payload = {tc: 900, tb: 10}
	}
	socket.send(e.toJSON())
}

const socket = new WebSocket("http://localhost:3502/ws?rid=hub")

socket.onmessage = (raw) => {
	const parsed = JSON.parse(raw.data)
	handleEvent(new WSEvent(parsed.a, parsed.p))
}

socket.onopen = () => {
}

socket.onclose = () => {

}

// Network delay in milliseconds.
const ping = document.getElementById("ping")

// Add cell event listeners.
for (let i = 1; i <= 9; i++) {
	document.getElementById(`cell-${i}`).addEventListener("click", () => {
		joinMatchmaking(i)
	})
}