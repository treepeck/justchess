import { Color } from "/assets/scripts/chess/types.js"
import { Board } from "/assets/scripts/widget/board.js"

const board = new Board(Color.White)
board.registerResizeObserver()
board.registerEventHandlers()
