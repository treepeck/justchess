import { appendMoveToTable, highlightCurrentMove } from "./chess/move"
import { formatTermination, formatResult } from "./chess/state"
import { Clock, Color } from "./utils/clock"
import { getOrPanic } from "./utils/dom"
import showDialog from "./utils/dialog"
import Board from "./chess/board"
;(() => {
	// Page guard.
	if (!document.getElementById("archiveGameGuard")) return

	// Render chessboard.
	const board = new Board(
		() => {},
		getOrPanic("boardContainer").classList.contains("flipped"),
	)

	const r = getOrPanic("endgameDialogResult")
	// @ts-ignore
	r.textContent = formatResult(parseInt(r.textContent))
	const t = getOrPanic("endgameDialogTermination")
	// @ts-ignore
	t.textContent = formatTermination(parseInt(t.textContent))
	showDialog("endgameDialog")

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

	const setCurrentMove = () => {
		clock.setTime(Color.White, times[board.currentFen].whiteTime)
		clock.setTime(Color.Black, times[board.currentFen].blackTime)
		highlightCurrentMove(board.currentFen)
		board.parsePiecePlacement(board.fens[board.currentFen])
	}

	// Go through move history using buttons.
	getOrPanic("nullMoveBtn").onclick = () => {
		board.currentFen = 0
		setCurrentMove()
	}
	getOrPanic("prevMoveBtn").onclick = () => {
		if (board.currentFen == 0) return
		board.currentFen -= 1
		setCurrentMove()
	}
	getOrPanic("nextMoveBtn").onclick = () => {
		if (board.currentFen == board.fens.length - 1) return
		board.currentFen += 1
		setCurrentMove()
	}
	getOrPanic("lastMoveBtn").onclick = () => {
		board.currentFen = board.fens.length - 1
		setCurrentMove()
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
		setCurrentMove()
	}
})()
