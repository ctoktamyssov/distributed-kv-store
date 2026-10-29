package wal

import "testing"

func TestReopenReplaysEverything(t *testing.T) {
	// path := filepath.Join(t.TempDir(), "test.wal")
	// Open, Append 100 records, Close. Reopen with an apply that collects
	// records. Expect exactly those 100, in order.
}

func TestTornTailIsTruncated(t *testing.T) {
	// Append 3 records and Close. os.Truncate(path, size-3) to fake a crash
	// mid-write. Reopen: expect 2 records, and the file size == the end of
	// record 2. Then Append one more and reopen. Expect 3, which proves new
	// writes don't land after garbage.
}

func TestAppendAfterCloseFails(t *testing.T) {}

// go test -bench=. ./wal/
func BenchmarkAppendAlways(b *testing.B) { benchAppend(b, SyncAlways) }
func BenchmarkAppendBatch(b *testing.B)  { benchAppend(b, SyncBatch) }
func BenchmarkAppendNever(b *testing.B)  { benchAppend(b, SyncNever) }

func benchAppend(b *testing.B, p SyncPolicy) {
	// Open in b.TempDir(), make one 100-byte record, b.ResetTimer(),
	// Append it b.N times. These numbers go in your README.
}
