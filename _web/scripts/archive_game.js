import { appendMoveToTable, Move } from "./chess/move"
import { getOrPanic } from "./utils/dom"
import Board from "./chess/board"
;(() => {
	// Page guard.
	if (!document.getElementById("archiveGameGuard")) return

	// Render chessboard.
	const el = /** @type {HTMLDivElement} */ (getOrPanic("board"))
	const board = new Board(el, () => {})

	// Add board event listeners.
	el.onmousedown = (e) => board.onMouseDown(e)
	el.onmousemove = (e) => board.onMouseMove(e)
	el.onmouseup = (e) => board.onMouseUp(e)

	// Make board responsive.
	const observer = new ResizeObserver((entries) => {
		for (const entry of entries) {
			board.setSize(entry.contentRect.width)
		}
	})
	observer.observe(el)

	// Append completed moves the to table.
	const moves = JSON.parse(getOrPanic("movesJson").textContent)
	for (let i = 0; i < moves.length; i++) {
		board.makeMove(new Move(moves[i].Move))
		appendMoveToTable(moves[i].San, i + 1)
	}
})()
