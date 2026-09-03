package compression

import (
	"bytes"
)

const (
	// For x86-64 CPUs int size is 32 bits. For x64 CPUs int size is 64 bits.
	intSize = (32 << (^uint(0) >> 63))
)

// bitWriter writes and stores the bit set (aka bit array) of arbitrary size.
type bitWriter struct {
	buff bytes.Buffer
	// Big endian temporary bit buffer.
	temp          uint
	remainingBits int
}

// write writes data to the writer.  If size is less than or equal to the
// integer size (which depends on the CPU architecture), the data is stored in
// an internal integer field. When the field overflows, its contents are flushed
// to the internal bytes.Buffer.
func (bw *bitWriter) write(data uint, size int) {
	bw.remainingBits -= size
	if bw.remainingBits >= 0 {
		bw.temp |= data << bw.remainingBits
	} else {
		bw.temp |= data >> -bw.remainingBits
		// Split integer into the byte sequence.
		for i := (intSize / 8) - 1; i >= 0; i-- {
			chunk := byte(bw.temp >> (i * 8))
			// Don't handle error since WriteByte always returns nil.
			bw.buff.WriteByte(chunk)
		}
		bw.remainingBits += intSize
		bw.temp = data << bw.remainingBits
	}
}

// writeCompressed writes an unsigned integer in compressed chunks.
//
// If n fits in 5 bits, all of them are written at once, along with a 6th
// continuation bit set to 0.
//
// If n is larger than 5 bits, the first 5 bits are written with the 6th
// continuation bit set to 1. The remaining bits of n are then split into
// 3-bit chunks. Each chunk is written as 4 bits, with the 4th bit acting
// as a continuation flag indicating whether more chunks follow.
//
// Do not mix [bitWriter.write] and [bitWriter.writeCompressed] calls. They are
// intended for different purposes:
//   - [bitWriter.write] is for move encoding;
//   - [bitWriter.writeCompressed] is for clock compression.
//
// To prevent a large number of bits being used for negative numbers,
// n is encoded using zigzag encoding.
func (bw *bitWriter) writeCompressed(n int) {
	n = (n >> (intSize - 1)) ^ (n << 1)

	if n & ^0x1F == 0 {
		bw.write(uint(n), 6)
	} else {
		bw.write(uint(n|0x20)&(1<<6-1), 6)
		n >>= 5

		for n & ^7 != 0 {
			bw.write(uint(n|8)&(1<<4-1), 4)
			n >>= 3
		}
		bw.write(uint(n), 4)
	}
}

// content returns the accumulated bytes.
func (bw *bitWriter) content() []byte {
	// Ceiling division using plain integer arithmetics.
	// ceil(X / N) = (X + N - 1) / N
	remainingBytes := (intSize + 7 - bw.remainingBits) / 8
	// Write remaining bytes to the buffer.
	for i := range remainingBytes {
		chunk := byte(bw.temp >> (intSize - 8 - i*8))
		// Don't handle error since WriteByte always returns nil.
		bw.buff.WriteByte(chunk)
	}
	bw.remainingBits = 0
	return bw.buff.Bytes()
}

// bitReader wraps a byte buffer and reads arbitrary chunks of bits from it.
type bitReader struct {
	buff []byte
	// Big endian temporary bit buffer.
	temp          uint
	remainingBits int
}

// fillTemp fills the internal temporary buffer of the reader.
func (br *bitReader) fillTemp() {
	br.temp = 0
	if len(br.buff) >= intSize/8 {
		// Split integer into the byte sequence.
		for i := intSize/8 - 1; i >= 0; i-- {
			br.temp |= uint(br.buff[0]) << (i * 8)
			// Delete a read chunk.
			br.buff = br.buff[1:]
		}
		br.remainingBits = intSize
	} else {
		br.remainingBits = len(br.buff) * 8
		for i, chunk := range br.buff {
			br.temp |= uint(chunk) << (br.remainingBits - 8 - i*8)
		}
	}
}

// read reads the specified amount of bits.
func (br *bitReader) read(size int) uint {
	if br.remainingBits >= size {
		br.remainingBits -= size
		return br.temp >> br.remainingBits & (1<<size - 1)
	}

	res := br.temp & (1<<br.remainingBits - 1)
	need := size - br.remainingBits
	br.fillTemp()
	return res<<need | br.read(need)
}

// readCompressed reads the unsigned integer from the internal buffer.
//
// Do not mix [bitReader.read] and [bitReader.readCompressed] calls. They are
// intended for different purposes:
//   - [bitReader.read] is for move decoding;
//   - [bitReader.readCompressed] is for clock decompression.
func (br *bitReader) readCompressed() int {
	n := br.read(6)
	// Keep on reading if the continuation bit is set.
	if n&0x20 != 0 {
		// Reset continuation bit.
		n &= 0x1F
		// Read remaining 4-bit chunks.
		i := 5
		for {
			chunk := br.read(4)
			n |= chunk & 7 << i
			i += 3
			if chunk&8 == 0 {
				break
			}
		}
	}
	// Return zigzag decoded number.
	return int(n>>1) ^ -int(n&1)
}
