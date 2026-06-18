import Board from "/assets/scripts/board.js"
import { get } from "/assets/scripts/dom.js"

const startPiecePlacement = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR"

const board = new Board()
board.parsePiecePlacement(startPiecePlacement)

get("boardFlip").onclick = () => board.flip()

const gameId = window.location.pathname.split("/").at(-1)

const socket = new WebSocket(`/ws/${gameId}`)
