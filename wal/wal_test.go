package wal

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// makeRecord builds a predictable record for test i: key3 → val3, etc.
func makeRecord(i int) Record {
	return Record{
		Op:    OpSet,
		Key:   []byte(fmt.Sprintf("key%d", i)),
		Value: []byte(fmt.Sprintf("val%d", i)),
	}
}

// reopen opens the WAL at path and returns every record replay found.
func reopen(t *testing.T, path string) (*WAL, []Record) {
	t.Helper()
	var got []Record
	w, err := Open(path, SyncNever, func(r Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	return w, got
}

func TestReopenReplaysEverything(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.wal")

	w, _ := reopen(t, path)
	for i := 0; i < 100; i++ {
		if err := w.Append(makeRecord(i)); err != nil {
			t.Fatalf("Append %d: %v", i, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	w, got := reopen(t, path)
	defer w.Close()

	if len(got) != 100 {
		t.Fatalf("replayed %d records, want 100", len(got))
	}
	for i, r := range got {
		want := makeRecord(i)
		if r.Op != want.Op || !bytes.Equal(r.Key, want.Key) || !bytes.Equal(r.Value, want.Value) {
			t.Errorf("record %d = {%d %q %q}, want {%d %q %q}",
				i, r.Op, r.Key, r.Value, want.Op, want.Key, want.Value)
		}
	}
}

func TestTornTailIsTruncated(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.wal")

	// Write 3 records.
	w, _ := reopen(t, path)
	for i := 0; i < 3; i++ {
		if err := w.Append(makeRecord(i)); err != nil {
			t.Fatalf("Append %d: %v", i, err)
		}
	}
	w.Close()

	// Fake a crash: chop the last 3 bytes off record 3.
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Truncate(path, info.Size()-3); err != nil {
		t.Fatal(err)
	}

	// Reopen: only the 2 complete records should come back.
	w, got := reopen(t, path)
	if len(got) != 2 {
		t.Fatalf("after crash: replayed %d records, want 2", len(got))
	}

	// And the file should have been cut back to exactly the end of record 2.
	wantSize := int64(len(Encode(makeRecord(0))) + len(Encode(makeRecord(1))))
	info, err = os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != wantSize {
		t.Fatalf("file size = %d, want %d", info.Size(), wantSize)
	}

	// New writes must land right after record 2, not after garbage.
	if err := w.Append(makeRecord(99)); err != nil {
		t.Fatalf("Append after recovery: %v", err)
	}
	w.Close()

	w, got = reopen(t, path)
	defer w.Close()
	if len(got) != 3 {
		t.Fatalf("after new write: replayed %d records, want 3", len(got))
	}
	if !bytes.Equal(got[2].Key, []byte("key99")) {
		t.Errorf("third record key = %q, want %q", got[2].Key, "key99")
	}
}

func TestAppendAfterCloseFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.wal")

	w, _ := reopen(t, path)
	w.Close()

	err := w.Append(makeRecord(0))
	if err == nil {
		t.Fatal("Append after Close succeeded, want an error")
	}
	if !errors.Is(err, os.ErrClosed) {
		t.Errorf("err = %v, want one wrapping os.ErrClosed", err)
	}
}

// go test -bench=. ./wal/
func BenchmarkAppendAlways(b *testing.B) { benchAppend(b, SyncAlways) }
func BenchmarkAppendBatch(b *testing.B)  { benchAppend(b, SyncBatch) }
func BenchmarkAppendNever(b *testing.B)  { benchAppend(b, SyncNever) }

func benchAppend(b *testing.B, p SyncPolicy) {
	path := filepath.Join(b.TempDir(), "bench.wal")
	w, err := Open(path, p, func(Record) error { return nil })
	if err != nil {
		b.Fatal(err)
	}

	// 3-byte key + 97-byte value = 100 bytes of data per record.
	rec := Record{Op: OpSet, Key: []byte("key"), Value: bytes.Repeat([]byte("v"), 97)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := w.Append(rec); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()

	w.Close()
}
