import { Board } from "./board.js"
import { _WebSocket } from "./websocket.js"

// Message types.
const ACTION_ROOM_INFO = 0
const ACTION_LAST_MOVE = 2

// Room states.
const STATE_IN_PROGRESS = 0
const STATE_OVER = 1

const eventHandlers = new Map()
eventHandlers.set(ACTION_ROOM_INFO, updateRoomInfo)
eventHandlers.set(ACTION_LAST_MOVE, handleMove)

const connection = new _WebSocket(eventHandlers)

const board = new Board(onMoveCallback)

const boardFlipButton = document.getElementById("board-flip")
boardFlipButton.addEventListener("click", (e) => {
    board.flipPerspective()
})

function onMoveCallback(move) {
    // TODO: check if the move is valid.
    connection.sendMakeMove(move)
}

function updateRoomInfo(roomInfo) {
    // Update spectators counter.
    const counter = document.getElementById("clients-counter")
    counter.innerText = "Clients in room: " + roomInfo.cc.toString()

    // Update players' id.
    const white = document.getElementById("white-status")
    white.innerText = "White: " + roomInfo.w

    const black = document.getElementById("black-status")
    black.innerText = "Black: " + roomInfo.b

    // Update room state.
    const state = document.getElementById("room-state")
    state.innerText = "Game state: "
    switch (roomInfo.rs) {
        case STATE_IN_PROGRESS:
            state.innerText += "in progress"
        break
        case STATE_OVER:
            state.innerText += "over"
        break
    } 
}

function handleMove(move) { 
    console.log("received last move: ", move)

    // Update board state.
    board.setBoard(move.f, move.l)

    // Update legal moves.

}