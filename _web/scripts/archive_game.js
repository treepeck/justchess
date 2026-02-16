import { appendMoveToTable, Move, highlightCurrentMove } from "./chess/move"
import { getOrPanic } from "./utils/dom"
import Board from "./chess/board"
;(() => {
	// Page guard.
	if (!document.getElementById("archiveGameGuard")) return

	// Render chessboard.
	const board = new Board(() => {})

	// Append completed moves the to table.
	const moves = JSON.parse(getOrPanic("movesJson").textContent)
	for (let i = 0; i < moves.length; i++) {
		board.makeMove(new Move(moves[i].Move))
		board.fens.push(board.serializePiecePlacement())
		board.currentFen += 1
		appendMoveToTable(moves[i].San, board.currentFen, (index) => {
			board.currentFen = index
			highlightCurrentMove(index)
			board.parsePiecePlacement(board.fens[index])
		})
	}
	highlightCurrentMove(board.currentFen)
})()
