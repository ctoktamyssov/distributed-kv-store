package wal

import "testing"

func TestRoundTrip(t *testing.T) {
	// Table-driven: normal SET, empty value, DEL, a value containing "\n" and
	// "\x00", and a 1 MB value. Encode → bytes.NewReader → Decode → compare.
	// Also check the returned n == len(encoded).
}

func TestDecodeEmptyIsEOF(t *testing.T) {
	// Decode an empty reader. Expect io.EOF exactly.
}

func TestDecodeTruncated(t *testing.T) {
	// Encode one record, then for every cut point 1..len-1:
	// Decode(buf[:cut]) must return io.ErrUnexpectedEOF.
	// This loop is what catches the TODO 3 gotcha.
}

func TestDecodeCorrupt(t *testing.T) {
	// Flip one byte in the key region. Expect ErrCorrupt.
}
