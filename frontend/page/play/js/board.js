import { Piece } from "./enum.js"
import { parseBoard } from "./fen.js"

const BOARD_SIZE = 480
const SQUARE_SIZE = Math.floor(BOARD_SIZE / 8)

const squareString = [
	"a1", "b1", "c1", "d1", "e1", "f1", "g1", "h1",
	"a2", "b2", "c2", "d2", "e2", "f2", "g2", "h2",
	"a3", "b3", "c3", "d3", "e3", "f3", "g3", "h3",
	"a4", "b4", "c4", "d4", "e4", "f4", "g4", "h4",
	"a5", "b5", "c5", "d5", "e5", "f5", "g5", "h5",
	"a6", "b6", "c6", "d6", "e6", "f6", "g6", "h6",
	"a7", "b7", "c7", "d7", "e7", "f7", "g7", "h7",
	"a8", "b8", "c8", "d8", "e8", "f8", "g8", "h8",
]

const defaultLegalMoves = {
    "b1": "a3c3",
    "g1": "f3h3",
    "a2": "a3a4",
    "b2": "b3b4",
    "c2": "c3c4",
    "d2": "d3d4",
    "e2": "e3e4",
    "f2": "f3f4",
    "g2": "g3g4",
    "h2": "h3h4",
}

// Piece that is being dragged now.
class DraggedPiece {
    constructor(canvasX, canvasY, initRank, initFile, type) {
        this.canvasX = canvasX
        this.canvasY = canvasY
        this.initRank = initRank
        this.initFile = initFile
        this.type = type
    }
}

// Board manages a chessboard drawen on a canvas.
export class Board {
    constructor(onMoveCallback, fenString = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR",
    legalMoves = defaultLegalMoves) {
        // Get canvas.
        this.canvas = document.getElementById("board")
        // Get drawing context.
        this.context = this.canvas.getContext("2d")
        // Disable image bluring.
        this.context.imageSmoothingEnabled = false
        // Add event listeners.
        this.canvas.addEventListener("mousedown", (e) => this.#onDragStart(e))
        this.canvas.addEventListener("mousemove", (e) => this.#onMouseMove(e))
        this.canvas.addEventListener("mouseup", (e) => this.#onDragRelease(e))

        // Initialize current selected square.
        this.selectedSquare = -1
        
        // Set default board view perspective.
        this.perspective = "white"

        // Initialize piece that is being dragged now.
        this.draggedPiece = null

        this.legalMoves = legalMoves

        this.board = parseBoard(fenString)

        this.onMoveCallback = onMoveCallback

        // Fetch sprite sheet.
        const sheet = new Image()
        sheet.src = "./sheet.png"

        sheet.onload = () => {
            this.sheet = sheet
            this.#draw()
        }
    }

    flipPerspective() {
        this.perspective = this.perspective === "white" ? "black" : "white"
        this.#draw()
    }

    setBoard(fenString, legalMoves) {
        this.board = parseBoard(fenString)
        this.legalMoves = legalMoves
        this.#draw()
    }

    #draw() {
        if (!this.sheet) {
            console.error("Sprite sheet not finished loading.")
            return
        }

        for (let rank = 0; rank < 8; rank++) {
            for (let file = 0; file < 8; file++) {
                const cRank = this.perspective === "white" ? 7 - rank : rank
                const cFile = this.perspective === "white" ? file : 7 - file

                // Dark square.
                if ((cRank+cFile) % 2 === 1) { 
                    this.#fillSquare(cRank, cFile, "#615F52")
                } 
                // Light square.
                else {
                    this.#fillSquare(cRank, cFile, "#fbf6d4")
                }

                // Highlight selected square.
                if (this.selectedSquare === 8 * rank + file) {
                    this.#fillSquare(cRank, cFile, "green")
                }            

                // Highlight legal moves for the selected piece.
                const dests = this.legalMoves[squareString[this.selectedSquare]]
                if (dests !== undefined) {
                    for (let i = 0; i < dests.length; i+=2) {
                        const s = string2Square(dests.substring(i, i+2))
                        const r = this.perspective === "white" ? 7 - s.rank : s.rank
                        const f = this.perspective === "white" ? s.file : 7 - s.file
                    
                        this.context.strokeStyle = "green"
                        this.context.strokeRect(
                            SQUARE_SIZE * f, SQUARE_SIZE * r,
                            SQUARE_SIZE, SQUARE_SIZE
                        )
                    }
                }

                // Draw piece on square.
                const pieceType = this.board[rank][file]
                if (pieceType !== Piece.NP) {
                    this.#drawPiece(cRank, cFile, pieceType % 6, pieceType > 5 ? 1 : 0)
                }
            }
        }

        // Draw piece that is being dragged now.
        if (this.draggedPiece !== null) {
            this.#drawDraggedPiece()
        }       
    }

    #fillSquare(rank, file, color) {
        this.context.fillStyle = color
        this.context.fillRect(
            SQUARE_SIZE * file,
            SQUARE_SIZE * rank,
            SQUARE_SIZE,
            SQUARE_SIZE
        )
    }

    #drawPiece(rank, file, srcX, srcY) {
        this.context.drawImage(
            this.sheet,
            16*srcX, 16*srcY,
            16, 16,
            SQUARE_SIZE * file + 2,
            SQUARE_SIZE * rank + 2,
            SQUARE_SIZE - 5, SQUARE_SIZE - 5
        )
    }

    #drawDraggedPiece() {
        this.context.drawImage(
            this.sheet,
            16*(this.draggedPiece.type % 6),
            16*(this.draggedPiece.type > 5 ? 1 : 0),
            16, 16,
            Math.floor(this.draggedPiece.canvasX) - 8,
            Math.floor(this.draggedPiece.canvasY) - 8,
            // Make dragged piece a bigger.
            SQUARE_SIZE + 16, SQUARE_SIZE + 16
        )
    }

    #onDragStart(e) {
        e.preventDefault()

        const coords = this.#getSquareCoords(e)
        this.selectedSquare = 8 * coords.rank + coords.file

        const piece = this.board[coords.rank][coords.file]
        if (piece !== Piece.NP) { 
            // Begin moving piece that is being dragged now.
            this.board[coords.rank][coords.file] = Piece.NP

            this.#setDraggedPiece(e, piece, coords)
        }

        this.#draw()
    }

    #onMouseMove(e) {
        e.preventDefault()

        if (this.draggedPiece === null) {
            return
        }

        const rect = this.canvas.getBoundingClientRect()
        this.draggedPiece.canvasX = (e.clientX - rect.left) - SQUARE_SIZE / 2
        this.draggedPiece.canvasY = (e.clientY - rect.top) - SQUARE_SIZE / 2
        this.#draw()
    }

    #onDragRelease(e) {
        e.preventDefault()

        if (this.draggedPiece === null) {
            return
        }

        const coords = this.#getSquareCoords(e)

        this.board[this.draggedPiece.initRank][this.draggedPiece.initFile] = this.draggedPiece.type
        
        this.onMoveCallback({
            to: 8 * coords.rank + coords.file,
            from: 8 * this.draggedPiece.initRank + this.draggedPiece.initFile,
            promotionPiece: null,
        })

        this.draggedPiece = null
        this.#draw()
    }

    #getSquareCoords(e) {
        const rect = this.canvas.getBoundingClientRect()
        const cRank = Math.floor((e.clientY - rect.top)  / SQUARE_SIZE)
        const cFile = Math.floor((e.clientX - rect.left) / SQUARE_SIZE)

        let coords = {}
        if (this.perspective === "white") {
            coords = { rank: 7 - cRank, file: cFile }
        } else {
            coords =  { rank: cRank, file: 7 - cFile }
        }    

        return coords
    }

    #setDraggedPiece(e, piece, coords) {
        const rect = this.canvas.getBoundingClientRect()

        this.draggedPiece = new DraggedPiece(
            (e.clientX - rect.left) - SQUARE_SIZE / 2, // Center dragged piece.
            (e.clientY - rect.top) - SQUARE_SIZE / 2, // Center dragged piece.
            coords.rank, coords.file,
            piece
        )
    }
}

function string2Square(str) {
    let file = 0
    switch (str[0]) {
        case 'b': file = 1; break;
        case 'c': file = 2; break;
        case 'd': file = 3; break;
        case 'e': file = 4; break;
        case 'f': file = 5; break;
        case 'g': file = 6; break;
        case 'h': file = 7; break;
    }
    return { rank: parseInt(str[1]) - 1, file: file }
}