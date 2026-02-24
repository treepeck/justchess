import { appendMoveToTable, Move, highlightCurrentMove } from "./chess/move"
import { formatTermination, formatResult } from "./chess/state"
import { getOrPanic } from "./utils/dom"
import Board from "./chess/board"
import { formatTime, Clock, Color } from "./utils/clock"
;(() => {
	// Page guard.
	if (!document.getElementById("archiveGameGuard")) return

	// Render chessboard.
	const board = new Board(() => {})

	const r = getOrPanic("endgameDialogResult")
	// @ts-ignore
	r.textContent = formatResult(parseInt(r.textContent))
	const t = getOrPanic("endgameDialogTermination")
	// @ts-ignore
	t.textContent = formatTermination(parseInt(t.textContent))

	// @ts-ignore
	const control = parseInt(getOrPanic("whiteClock").textContent) * 60 * 1000
	const clock = new Clock(control, false, Color.White, 0)

	/** @type {{whiteTime: number, blackTime: number}[]} */
	const times = [{ whiteTime: control, blackTime: control }]

	// Append completed moves the to table.
	const raw = getOrPanic("movesJson").textContent
	if (raw !== null) {
		const moves = JSON.parse(raw)
		for (let i = 0; i < moves.length; i++) {
			board.parsePiecePlacement(moves[i].f)
			board.currentFen += 1
			board.fens.push(moves[i].f)

			// Update clock to show the time after the move was performed.
			let wt = clock.whiteTime
			let bt = clock.blackTime
			if (i % 2 === 0) {
				wt = clock.whiteTime - moves[i].t * 1000
				clock.setTime(Color.White, wt)
			} else {
				bt = clock.blackTime - moves[i].t * 1000
				clock.setTime(Color.Black, bt)
			}
			times.push({ whiteTime: wt, blackTime: bt })

			appendMoveToTable(moves[i].s, board.currentFen, (index) => {
				board.currentFen = index
				highlightCurrentMove(index)
				clock.setTime(Color.White, wt)
				clock.setTime(Color.Black, bt)
				board.parsePiecePlacement(board.fens[index])
			})
		}
		highlightCurrentMove(board.currentFen)
	}

	// Go through move history using keyboard.
	document.onkeydown = (e) => {
		switch (e.key) {
			case "ArrowUp":
				board.currentFen = board.fens.length - 1
				break
			case "ArrowRight":
				if (board.currentFen == board.fens.length - 1) return
				board.currentFen += 1
				break
			case "ArrowDown":
				board.currentFen = 0
				break
			case "ArrowLeft":
				if (board.currentFen == 0) return
				board.currentFen -= 1
				break
		}
		clock.setTime(Color.White, times[board.currentFen].whiteTime)
		clock.setTime(Color.Black, times[board.currentFen].blackTime)
		highlightCurrentMove(board.currentFen)
		board.parsePiecePlacement(board.fens[board.currentFen])
	}
})()
