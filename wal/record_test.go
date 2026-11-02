package wal

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		rec  Record
	}{
		{"normal set", Record{Op: OpSet, Key: []byte("foo"), Value: []byte("bar")}},
		{"empty value", Record{Op: OpSet, Key: []byte("foo"), Value: []byte{}}},
		{"del", Record{Op: OpDel, Key: []byte("foo")}},
		{"newline and null", Record{Op: OpSet, Key: []byte("k"), Value: []byte("a\nb\x00c")}},
		{"1 MB value", Record{Op: OpSet, Key: []byte("big"), Value: bytes.Repeat([]byte("x"), 1<<20)}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := Encode(tc.rec)

			got, n, err := Decode(bytes.NewReader(buf))
			if err != nil {
				t.Fatalf("Decode returned error: %v", err)
			}
			if n != len(buf) {
				t.Errorf("n = %d, want %d", n, len(buf))
			}
			if got.Op != tc.rec.Op {
				t.Errorf("Op = %d, want %d", got.Op, tc.rec.Op)
			}
			if !bytes.Equal(got.Key, tc.rec.Key) {
				t.Errorf("Key = %q, want %q", got.Key, tc.rec.Key)
			}
			if !bytes.Equal(got.Value, tc.rec.Value) {
				// Don't print the values: one of them is 1 MB.
				t.Errorf("Value mismatch: got %d bytes, want %d", len(got.Value), len(tc.rec.Value))
			}
		})
	}
}

func TestDecodeEmptyIsEOF(t *testing.T) {
	_, _, err := Decode(bytes.NewReader(nil))
	if !errors.Is(err, io.EOF) {
		t.Fatalf("err = %v, want io.EOF", err)
	}
}

func TestDecodeTruncated(t *testing.T) {
	buf := Encode(Record{Op: OpSet, Key: []byte("foo"), Value: []byte("bar")})

	for cut := 1; cut < len(buf); cut++ {
		_, _, err := Decode(bytes.NewReader(buf[:cut]))
		if !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Errorf("cut at %d of %d: err = %v, want io.ErrUnexpectedEOF", cut, len(buf), err)
		}
	}
}

func TestDecodeCorrupt(t *testing.T) {
	buf := Encode(Record{Op: OpSet, Key: []byte("foo"), Value: []byte("bar")})

	buf[headerSize] ^= 0xFF // flip every bit of the first key byte

	_, _, err := Decode(bytes.NewReader(buf))
	if !errors.Is(err, ErrCorrupt) {
		t.Fatalf("err = %v, want ErrCorrupt", err)
	}
}
