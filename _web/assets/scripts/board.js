import { Board } from "/assets/scripts/render/board.js"
import { parseFen } from "/assets/scripts/chess/fen.js"

const initialPosition = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

const boardElement = document.getElementById("board")

const position = parseFen(initialPosition)

const board = new Board(position, boardElement)
board.render()
board.observeResize()
board.registerEventHandlers()