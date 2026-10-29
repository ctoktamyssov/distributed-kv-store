package wal

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

type SyncPolicy int

const (
	SyncAlways SyncPolicy = iota // fsync every Append
	SyncBatch                    // fsync every batchInterval in the background
	SyncNever                    // let the OS decide
)

const batchInterval = 10 * time.Millisecond

func ParseSyncPolicy(s string) (SyncPolicy, error) {
	// TODO: "always" / "batch" / "never", otherwise an error
	return 0, nil
}

type WAL struct {
	mu     sync.Mutex
	f      *os.File
	policy SyncPolicy
	dirty  bool  // written since the last fsync (batch mode)
	err    error // sticky: once a write or sync fails, every later Append fails

	done chan struct{}
	wg   sync.WaitGroup
}

// Open opens or creates the log at path, replays every valid record into
// apply, truncates a torn tail, and returns a WAL ready for appends.
//
// Taking apply as a parameter makes it impossible to Append before replay.
func Open(path string, policy SyncPolicy, apply func(Record) error) (*WAL, error) {
	// TODO 1: os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o644)
	//         O_APPEND means writes always go to the end, even after you've
	//         seeked around during replay.
	// TODO 2: good, err := replay(f, apply)
	// TODO 3: if good < file size → f.Truncate(good) and log a warning
	//         saying how many bytes you dropped
	// TODO 4: build the WAL; if policy == SyncBatch, start w.syncLoop()
	//         (wg.Add(1) before the `go`)
	return nil, nil
}

// replay decodes records from the start of f and returns the offset just
// past the last good record.
func replay(f *os.File, apply func(Record) error) (int64, error) {
	// TODO 1: f.Seek(0, io.SeekStart); rd := bufio.NewReader(f)
	// TODO 2: loop Decode(rd):
	//         nil                          → apply(r); offset += n
	//         io.EOF                       → return offset, nil
	//         ErrUnexpectedEOF, ErrCorrupt → return offset, nil (the caller truncates)
	//         anything else                → return offset, err (a real I/O error)
	return 0, nil
}

// Append durably (per policy) writes r to the end of the log.
func (w *WAL) Append(r Record) error {
	// TODO 1: lock; if w.err != nil, return it
	// TODO 2: f.Write(Encode(r)); on error set w.err and return it
	// TODO 3: SyncAlways → f.Sync(); on error set w.err
	//         SyncBatch  → w.dirty = true
	//
	// Why sticky errors: if a write half-succeeds, there's a torn record in the
	// middle of the file. Anything appended after it would be truncated away on
	// the next replay. Refusing all further writes is the only safe option.
	// (Postgres learned this the hard way. Search "fsyncgate" for the README.)
	return nil
}

func (w *WAL) Sync() error {
	// TODO: lock, f.Sync(), dirty = false, record any error in w.err
	return nil
}

func (w *WAL) syncLoop() {
	// TODO: defer wg.Done(); ticker := time.NewTicker(batchInterval)
	//       select on ticker.C (sync if dirty) and w.done (return)
}

func (w *WAL) Close() error {
	// TODO: if batch mode: close(w.done); w.wg.Wait()
	//       final Sync(), then f.Close()
	return nil
}

var _ = errors.New // delete these once you use the imports
var _ = fmt.Sprintf
