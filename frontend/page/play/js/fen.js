import { Piece } from "./enum.js"

// parseBoard parses piece position of the specified part of FEN string.
export function parseBoard(piecePlacementData) {
    const board = new Array(8)
    for (let rank = 0; rank < 8; rank++) {
        board[rank] = new Array(8)
        for (let file = 0; file < 8; file++) {
            board[rank][file] = Piece.NP
        }
    }

    let squareIndex = 56
    
    for (let i = 0; i < piecePlacementData.length; i++) {
        const char = piecePlacementData[i]
        const num = parseInt(char)

        // Rank separator.
        if (char === "/") {
            squareIndex -= 16
        }
        // Number of consecutive empty squares.
        else if (num !== NaN && num >= 1 && num <= 8) {
            squareIndex += num
        }
        // There is piece on a square.
        else {
            let pieceType = Piece.WP // White pawn by default.

            switch(char) {
                case "N": pieceType = Piece.WN; break
                case "B": pieceType = Piece.WB; break
                case "R": pieceType = Piece.WR; break
                case "Q": pieceType = Piece.WQ; break
                case "K": pieceType = Piece.WK; break
                case "p": pieceType = Piece.BP; break
                case "n": pieceType = Piece.BN; break
                case "b": pieceType = Piece.BB; break
                case "r": pieceType = Piece.BR; break
                case "q": pieceType = Piece.BQ; break
                case "k": pieceType = Piece.BK; break
            }
            // Place the piece on a board.
            board[Math.floor(squareIndex/8)][squareIndex%8] = pieceType
            squareIndex++
        }
    }

    return board
}
