// Package compression implements move and clock compression.
// Moves are encoded using Huffman's algorithm.
// Clock values are encoded using Zig-Zag algorithm.
package compression

import (
	"github.com/treepeck/chego"
)

// Move represents the decompressed chess move.
type Move struct {
	// Standard Algebraic Notation.
	San string `json:"s"`
	// Forsyth-Edwards Notation.
	Fen string `json:"f"`
}

// HuffmanEncoding encodes the list of legal moves by their indices in the move list.
func HuffmanEncoding(indices []byte) []byte {
	w := bitWriter{remainingBits: intSize}
	for _, ind := range indices {
		e := huffmanCodes[ind]
		w.write(e.code, e.size)
	}
	return w.content()
}

// HuffmanDecoding decodes the given legal moves.
func HuffmanDecoding(encoded []byte, length int) []Move {
	decoded := make([]Move, 0, length)

	r := bitReader{buff: encoded}

	p := chego.ParseFen(chego.InitialPos)
	var l chego.MoveList
	chego.GenLegalMoves(*p, &l)

	for range length {
		// No valid Huffman code is a prefix of another Huffman code.
		// Therefore, to decode the move, traverse the trie until reaching a
		// leaf node.
		n := trie
		for n.left != nil && n.right != nil {
			next := r.read(1)
			if next == 0 {
				n = n.left
			} else {
				n = n.right
			}
		}
		// Collect results of decoding.
		m := l.Moves[n.index]
		decoded = append(decoded, Move{
			San: chego.Move2SAN(m, p, &l),
			Fen: chego.SerializeBitboards(p.Bitboards),
		})
	}

	return decoded
}

func CompressTimeDiffs(timeDiffs []int) []byte {
	w := &bitWriter{remainingBits: intSize}
	for _, diff := range timeDiffs {
		w.writeCompressed(diff)
	}
	return w.content()
}

func DecompressTimeDiffs(compressed []byte, length int) []int {
	timeDiffs := make([]int, 0, length)

	r := &bitReader{buff: compressed}
	for range length {
		timeDiffs = append(timeDiffs, r.readCompressed())
	}

	return timeDiffs
}

// node of the Huffman trie.
type node struct {
	// There are only two symbols in the alphabet - 0 and 1, therefore two child nodes.
	left, right *node
	// Index of the [huffmanEntry] in [huffmanCodes] array.
	index int
}

// makeTrie fills the trie with nodes which represents the [huffmanEntry].
func makeTrie() *node {
	root := &node{index: -1}

	for i := range 218 {
		e := huffmanCodes[i]

		curr := root

		for j := e.size - 1; j >= 0; j-- {
			// Get the most significant bit of the code since [bitReader]
			// read bits in big endian order.
			bit := e.code >> j & 1

			if bit == 0 {
				if curr.left == nil {
					curr.left = &node{index: -1}
				}
				curr = curr.left
			} else {
				if curr.right == nil {
					curr.right = &node{index: -1}
				}
				curr = curr.right
			}
		}

		curr.index = i
	}
	return root
}

// trie (aka prefix tree) is used for Huffman decoding.
var trie = makeTrie()

// Entry of the [huffmanCodes] array.
type huffmanEntry struct {
	code uint
	size int
}

// Huffman codes for legal move list indices.
// To calculate them, 10164006 games with 685863447 moves in total were
// analyzed.  See codegen.go for more details.
//
// The maximum number of legal moves in a chess position appears to be 218,
// so the array contains 218 elements. The latter half of the indices didn't
// actually occur in the dataset, but each was assigned a frequency of 1 to
// ensure valid Huffman codes are generated.
// If a move has a frequency of 0, no code would be produced, which is
// undesirable since the move is still theoretically possible.
//
// The frequencies were generated using data from the Lichess database exports
// (https://database.lichess.org), which are released under the Creative
// Commons CC0 license.
var huffmanCodes = [218]huffmanEntry{
	{0b1011, 4},                            // index 0 | played 35516075 times
	{0b00011, 5},                           // index 1 | played 28863637 times
	{0b1100, 4},                            // index 2 | played 33697520 times
	{0b1111, 4},                            // index 3 | played 31340990 times
	{0b00110, 5},                           // index 4 | played 26616335 times
	{0b00100, 5},                           // index 5 | played 26967376 times
	{0b00111, 5},                           // index 6 | played 26599119 times
	{0b00000, 5},                           // index 7 | played 30127529 times
	{0b00101, 5},                           // index 8 | played 26726290 times
	{0b1110, 4},                            // index 9 | played 31546838 times
	{0b01011, 5},                           // index 10 | played 21719881 times
	{0b01101, 5},                           // index 11 | played 20960808 times
	{0b01111, 5},                           // index 12 | played 20924693 times
	{0b10001, 5},                           // index 13 | played 20426220 times
	{0b10000, 5},                           // index 14 | played 20450176 times
	{0b10010, 5},                           // index 15 | played 20288330 times
	{0b01100, 5},                           // index 16 | played 21182180 times
	{0b10011, 5},                           // index 17 | played 19779373 times
	{0b01010, 5},                           // index 18 | played 22055062 times
	{0b10101, 5},                           // index 19 | played 18959904 times
	{0b11011, 5},                           // index 20 | played 16182542 times
	{0b000011, 6},                          // index 21 | played 14643685 times
	{0b000010, 6},                          // index 22 | played 15035699 times
	{0b000100, 6},                          // index 23 | played 14551558 times
	{0b010000, 6},                          // index 24 | played 12841369 times
	{0b010010, 6},                          // index 25 | played 12121516 times
	{0b011100, 6},                          // index 26 | played 11024918 times
	{0b011101, 6},                          // index 27 | played 9908166 times
	{0b101001, 6},                          // index 28 | played 9388606 times
	{0b110101, 6},                          // index 29 | played 8215047 times
	{0b0001010, 7},                         // index 30 | played 7382257 times
	{0b0100010, 7},                         // index 31 | played 6656836 times
	{0b0100011, 7},                         // index 32 | played 6157014 times
	{0b0100111, 7},                         // index 33 | played 5400835 times
	{0b1010001, 7},                         // index 34 | played 4790308 times
	{0b1101000, 7},                         // index 35 | played 4378929 times
	{0b00010110, 8},                        // index 36 | played 3779824 times
	{0b01001100, 8},                        // index 37 | played 3261509 times
	{0b01001101, 8},                        // index 38 | played 2846448 times
	{0b10100001, 8},                        // index 39 | played 2399087 times
	{0b11010011, 8},                        // index 40 | played 2045159 times
	{0b000101110, 9},                       // index 41 | played 1707181 times
	{0b101000000, 9},                       // index 42 | played 1390278 times
	{0b110100100, 9},                       // index 43 | played 1139651 times
	{0b110100101, 9},                       // index 44 | played 932421 times
	{0b0001011111, 10},                     // index 45 | played 722679 times
	{0b1010000011, 10},                     // index 46 | played 623129 times
	{0b00010111101, 11},                    // index 47 | played 423358 times
	{0b10100000101, 11},                    // index 48 | played 320010 times
	{0b000101111001, 12},                   // index 49 | played 235655 times
	{0b101000001001, 12},                   // index 50 | played 175233 times
	{0b0001011110000, 13},                  // index 51 | played 127442 times
	{0b1010000010000, 13},                  // index 52 | played 91111 times
	{0b00010111100010, 14},                 // index 53 | played 64858 times
	{0b10100000100010, 14},                 // index 54 | played 46568 times
	{0b000101111000110, 15},                // index 55 | played 31905 times
	{0b101000001000110, 15},                // index 56 | played 22068 times
	{0b0001011110001110, 16},               // index 57 | played 15412 times
	{0b1010000010001110, 16},               // index 58 | played 10561 times
	{0b00010111100011111, 17},              // index 59 | played 7044 times
	{0b10100000100011111, 17},              // index 60 | played 4775 times
	{0b000101111000111101, 18},             // index 61 | played 3372 times
	{0b101000001000111101, 18},             // index 62 | played 2320 times
	{0b1010000010001111000, 19},            // index 63 | played 1633 times
	{0b00010111100011110000, 20},           // index 64 | played 1138 times
	{0b00010111100011110011, 20},           // index 65 | played 821 times
	{0b10100000100011110011, 20},           // index 66 | played 646 times
	{0b000101111000111100100, 21},          // index 67 | played 454 times
	{0b101000001000111100101, 21},          // index 68 | played 338 times
	{0b0001011110001111000100, 22},         // index 69 | played 294 times
	{0b0001011110001111001010, 22},         // index 70 | played 207 times
	{0b1010000010001111001000, 22},         // index 71 | played 195 times
	{0b00010111100011110001010, 23},        // index 72 | played 148 times
	{0b00010111100011110001100, 23},        // index 73 | played 134 times
	{0b00010111100011110010111, 23},        // index 74 | played 90 times
	{0b10100000100011110010010, 23},        // index 75 | played 85 times
	{0b000101111000111100010111, 24},       // index 76 | played 71 times
	{0b000101111000111100011100, 24},       // index 77 | played 62 times
	{0b000101111000111100011111, 24},       // index 78 | played 54 times
	{0b000101111000111100011101, 24},       // index 79 | played 59 times
	{0b0001011110001111000111100, 25},      // index 80 | played 30 times
	{0b101000001000111100100111, 24},       // index 81 | played 42 times
	{0b0001011110001111001011000, 25},      // index 82 | played 27 times
	{0b0001011110001111001011010, 25},      // index 83 | played 26 times
	{0b0001011110001111000111101, 25},      // index 84 | played 28 times
	{0b1010000010001111001001100, 25},      // index 85 | played 22 times
	{0b1010000010001111001001101, 25},      // index 86 | played 21 times
	{0b0001011110001111001011001, 25},      // index 87 | played 27 times
	{0b00010111100011110001011001, 26},     // index 88 | played 18 times
	{0b00010111100011110001011011, 26},     // index 89 | played 16 times
	{0b00010111100011110001101010, 26},     // index 90 | played 16 times
	{0b00010111100011110010110111, 26},     // index 91 | played 12 times
	{0b00010111100011110010110110, 26},     // index 92 | played 14 times
	{0b00010111100011110001011000010, 29},  // index 93 | played 3 times
	{0b0001011110001111000101100000, 28},   // index 94 | played 6 times
	{0b0001011110001111000101101001, 28},   // index 95 | played 4 times
	{0b000101111000111100010110001, 27},    // index 96 | played 9 times
	{0b00010111100011110001011000011, 29},  // index 97 | played 3 times
	{0b00010111100011110001011010001, 29},  // index 98 | played 2 times
	{0b00010111100011110001011010000, 29},  // index 99 | played 3 times
	{0b000101111000111100011010111010, 30}, // index 100 | played 1 times
	{0b00010111100011110001011010110, 29},  // index 101 | played 2 times
	{0b000101111000111100011010111011, 30}, // index 102 | played 1 times
	{0b000101111000111100011010111000, 30}, // index 103 | played 1 times
	{0b000101111000111100011010111001, 30}, // index 104 | played 1 times
	{0b000101111000111100011010111110, 30}, // index 105 | played 1 times
	{0b000101111000111100011010111111, 30}, // index 106 | played 1 times
	{0b000101111000111100011010111100, 30}, // index 107 | played 1 times
	{0b000101111000111100011010111101, 30}, // index 108 | played 1 times
	{0b00010111100011110001011010111, 29},  // index 109 | played 2 times
	{0b000101111000111100011010110010, 30}, // index 110 | played 1 times
	{0b000101111000111100011010110011, 30}, // index 111 | played 1 times
	{0b000101111000111100011010110000, 30}, // index 112 | played 1 times
	{0b000101111000111100011010110001, 30}, // index 113 | played 1 times
	{0b000101111000111100011010110110, 30}, // index 114 | played 1 times
	{0b000101111000111100011010110111, 30}, // index 115 | played 1 times
	{0b000101111000111100011010110100, 30}, // index 116 | played 1 times
	{0b000101111000111100011010110101, 30}, // index 117 | played 1 times
	{0b000101111000111100011010001010, 30}, // index 118 | played 1 times
	{0b000101111000111100011010001011, 30}, // index 119 | played 1 times
	{0b000101111000111100011010001000, 30}, // index 120 | played 1 times
	{0b000101111000111100011010001001, 30}, // index 121 | played 1 times
	{0b000101111000111100011010001110, 30}, // index 122 | played 1 times
	{0b000101111000111100011010001111, 30}, // index 123 | played 1 times
	{0b000101111000111100011010001100, 30}, // index 124 | played 1 times
	{0b000101111000111100011010001101, 30}, // index 125 | played 1 times
	{0b000101111000111100011010000010, 30}, // index 126 | played 1 times
	{0b000101111000111100011010000011, 30}, // index 127 | played 1 times
	{0b000101111000111100011010000000, 30}, // index 128 | played 1 times
	{0b000101111000111100011010000001, 30}, // index 129 | played 1 times
	{0b000101111000111100011010000110, 30}, // index 130 | played 1 times
	{0b000101111000111100011010000111, 30}, // index 131 | played 1 times
	{0b000101111000111100011010000100, 30}, // index 132 | played 1 times
	{0b000101111000111100011010000101, 30}, // index 133 | played 1 times
	{0b000101111000111100011010011010, 30}, // index 134 | played 1 times
	{0b000101111000111100011010011011, 30}, // index 135 | played 1 times
	{0b000101111000111100011010011000, 30}, // index 136 | played 1 times
	{0b000101111000111100011010011001, 30}, // index 137 | played 1 times
	{0b000101111000111100011010011110, 30}, // index 138 | played 1 times
	{0b000101111000111100011010011111, 30}, // index 139 | played 1 times
	{0b000101111000111100011010011100, 30}, // index 140 | played 1 times
	{0b000101111000111100011010011101, 30}, // index 141 | played 1 times
	{0b000101111000111100011010010010, 30}, // index 142 | played 1 times
	{0b000101111000111100011010010011, 30}, // index 143 | played 1 times
	{0b000101111000111100011010010000, 30}, // index 144 | played 1 times
	{0b000101111000111100011010010001, 30}, // index 145 | played 1 times
	{0b000101111000111100011010010110, 30}, // index 146 | played 1 times
	{0b000101111000111100011010010111, 30}, // index 147 | played 1 times
	{0b000101111000111100011010010100, 30}, // index 148 | played 1 times
	{0b000101111000111100011010010101, 30}, // index 149 | played 1 times
	{0b000101111000111100011011101010, 30}, // index 150 | played 1 times
	{0b000101111000111100011011101011, 30}, // index 151 | played 1 times
	{0b000101111000111100011011101000, 30}, // index 152 | played 1 times
	{0b000101111000111100011011101001, 30}, // index 153 | played 1 times
	{0b000101111000111100011011101110, 30}, // index 154 | played 1 times
	{0b000101111000111100011011101111, 30}, // index 155 | played 1 times
	{0b000101111000111100011011101100, 30}, // index 156 | played 1 times
	{0b000101111000111100011011101101, 30}, // index 157 | played 1 times
	{0b000101111000111100011011100010, 30}, // index 158 | played 1 times
	{0b000101111000111100011011100011, 30}, // index 159 | played 1 times
	{0b000101111000111100011011100000, 30}, // index 160 | played 1 times
	{0b000101111000111100011011100001, 30}, // index 161 | played 1 times
	{0b000101111000111100011011100110, 30}, // index 162 | played 1 times
	{0b000101111000111100011011100111, 30}, // index 163 | played 1 times
	{0b000101111000111100011011100100, 30}, // index 164 | played 1 times
	{0b000101111000111100011011100101, 30}, // index 165 | played 1 times
	{0b000101111000111100011011111010, 30}, // index 166 | played 1 times
	{0b000101111000111100011011111011, 30}, // index 167 | played 1 times
	{0b000101111000111100011011111000, 30}, // index 168 | played 1 times
	{0b000101111000111100011011111001, 30}, // index 169 | played 1 times
	{0b000101111000111100011011111110, 30}, // index 170 | played 1 times
	{0b000101111000111100011011111111, 30}, // index 171 | played 1 times
	{0b000101111000111100011011111100, 30}, // index 172 | played 1 times
	{0b000101111000111100011011111101, 30}, // index 173 | played 1 times
	{0b000101111000111100011011110010, 30}, // index 174 | played 1 times
	{0b000101111000111100011011110011, 30}, // index 175 | played 1 times
	{0b000101111000111100011011110000, 30}, // index 176 | played 1 times
	{0b000101111000111100011011110001, 30}, // index 177 | played 1 times
	{0b000101111000111100011011110110, 30}, // index 178 | played 1 times
	{0b000101111000111100011011110111, 30}, // index 179 | played 1 times
	{0b000101111000111100011011110100, 30}, // index 180 | played 1 times
	{0b000101111000111100011011110101, 30}, // index 181 | played 1 times
	{0b000101111000111100011011001010, 30}, // index 182 | played 1 times
	{0b000101111000111100011011001011, 30}, // index 183 | played 1 times
	{0b000101111000111100011011001000, 30}, // index 184 | played 1 times
	{0b000101111000111100011011001001, 30}, // index 185 | played 1 times
	{0b000101111000111100011011001110, 30}, // index 186 | played 1 times
	{0b000101111000111100011011001111, 30}, // index 187 | played 1 times
	{0b000101111000111100011011001100, 30}, // index 188 | played 1 times
	{0b000101111000111100011011001101, 30}, // index 189 | played 1 times
	{0b000101111000111100011011000010, 30}, // index 190 | played 1 times
	{0b000101111000111100011011000011, 30}, // index 191 | played 1 times
	{0b000101111000111100011011000000, 30}, // index 192 | played 1 times
	{0b000101111000111100011011000001, 30}, // index 193 | played 1 times
	{0b000101111000111100011011000110, 30}, // index 194 | played 1 times
	{0b000101111000111100011011000111, 30}, // index 195 | played 1 times
	{0b000101111000111100011011000100, 30}, // index 196 | played 1 times
	{0b000101111000111100011011000101, 30}, // index 197 | played 1 times
	{0b000101111000111100011011011010, 30}, // index 198 | played 1 times
	{0b000101111000111100011011011011, 30}, // index 199 | played 1 times
	{0b000101111000111100011011011000, 30}, // index 200 | played 1 times
	{0b000101111000111100011011011001, 30}, // index 201 | played 1 times
	{0b000101111000111100011011011110, 30}, // index 202 | played 1 times
	{0b000101111000111100011011011111, 30}, // index 203 | played 1 times
	{0b000101111000111100011011011100, 30}, // index 204 | played 1 times
	{0b000101111000111100011011011101, 30}, // index 205 | played 1 times
	{0b000101111000111100011011010010, 30}, // index 206 | played 1 times
	{0b000101111000111100011011010011, 30}, // index 207 | played 1 times
	{0b000101111000111100011011010000, 30}, // index 208 | played 1 times
	{0b000101111000111100011011010001, 30}, // index 209 | played 1 times
	{0b000101111000111100011011010110, 30}, // index 210 | played 1 times
	{0b000101111000111100011011010111, 30}, // index 211 | played 1 times
	{0b000101111000111100011011010100, 30}, // index 212 | played 1 times
	{0b000101111000111100011011010101, 30}, // index 213 | played 1 times
	{0b000101111000111100010110101010, 30}, // index 214 | played 1 times
	{0b000101111000111100010110101011, 30}, // index 215 | played 1 times
	{0b000101111000111100010110101000, 30}, // index 216 | played 1 times
	{0b000101111000111100010110101001, 30}, // index 217 | played 1 times
}
