import Board from "/js/board.js"

const sheet = new Image()
sheet.src = "/images/sheet.png"
sheet.onload = () => {
	const ctx = boardCanvas.getContext("2d")
	const board = new Board(ctx, sheet)

	const observer = new ResizeObserver((entries) => {
		for (const entry of entries) {
			board.setSize(entry.contentRect.width)
		}
	})
	observer.observe(boardCanvas)

	// Add event listeners.
	boardCanvas.addEventListener("mousedown", (e) => {
		board.onMouseDown(e)
	})
	boardCanvas.addEventListener("mousemove", (e) => {
		board.onMouseMove(e)
	})
	boardCanvas.addEventListener("mouseup", (e) => {
		board.onMouseUp(e)
	})
}
