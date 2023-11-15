package io

type BsWriter struct {
	buf    []byte
	offset int
	pos    int
}

// Write x in n bits
func (w *BsWriter) writeBits(x int, n int) {
	if n <= w.offset {
		// Just write everything to this byte
		x = x << (w.offset - n)
		w.buf[w.pos] += byte(x)
		w.offset -= n
	} else {
		// Recursively write to bytes
		actualBitSize := 1
		for {
			if (1 << actualBitSize) > x {
				break
			}
			actualBitSize += 1
		}

		consumedLen := n - w.offset
		mask := 1<<(consumedLen+1) - 1

		// Write to current byte first
		w.writeBits(x>>consumedLen, w.offset)

		// Write to new bytes
		w.writeBits(x&mask, consumedLen)
	}

	if w.offset == 0 {
		w.offset = 8
		w.pos += 1
	}
}

// Public API
func (w *BsWriter) Write(x int, n int) {
	w.writeBits(x, n)
}

func (w *BsWriter) WriteByte(x int) {
	w.writeBits(x, 8)
}

func (w *BsWriter) WriteShort(x int) {
	w.writeBits(x, 16)
}

func (w *BsWriter) WriteInt(x int) {
	w.writeBits(x, 32)
}

func (w *BsWriter) GetBuf() []byte {
	return w.buf
}

func GetBufferWriter(size int) BsWriter {
	return BsWriter{buf: make([]byte, size), offset: 8, pos: 0}
}
