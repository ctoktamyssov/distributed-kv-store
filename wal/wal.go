package wal

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
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
	switch s {
	case "always":
		return SyncAlways, nil
	case "batch":
		return SyncBatch, nil
	case "never":
		return SyncNever, nil
	}
	return 0, fmt.Errorf("wal: unknown sync policy %q (want always, batch, or never)", s)
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
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}

	good, err := replay(f, apply)
	if err != nil {
		f.Close()
		return nil, err
	}

	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}

	if size := info.Size(); good < size {
		if err := f.Truncate(good); err != nil {
			f.Close()
			return nil, fmt.Errorf("wal: truncate: %w", err)
		}
		if err := f.Sync(); err != nil { // make the truncate itself durable
			f.Close()
			return nil, fmt.Errorf("wal: sync after truncate: %w", err)
		}
		log.Printf("wal: dropped %d bytes of torn/corrupt tail from %s", size-good, path)
	}

	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		f.Close()
		return nil, err
	}

	w := &WAL{
		f:      f,
		policy: policy,
		done:   make(chan struct{}),
	}
	if policy == SyncBatch {
		w.wg.Add(1)
		go w.syncLoop()
	}
	return w, nil
}

// replay decodes records from the start of f and returns the offset just
// past the last good record.
func replay(f *os.File, apply func(Record) error) (int64, error) {
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return 0, err
	}
	rd := bufio.NewReader(f)

	var offset int64
	for {
		r, n, err := Decode(rd)
		switch {
		case err == nil:
			if err := apply(r); err != nil {
				return offset, err
			}
			offset += int64(n)
		case errors.Is(err, io.EOF):
			return offset, nil
		case errors.Is(err, io.ErrUnexpectedEOF), errors.Is(err, ErrCorrupt):
			return offset, nil // the caller truncates
		default:
			return offset, err // a real I/O error
		}
	}
}

// Append durably (per policy) writes r to the end of the log.
func (w *WAL) Append(r Record) error {
	if len(r.Key)+len(r.Value) > maxRecordSize {
		return ErrTooLarge
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.err != nil {
		return w.err
	}

	if _, err := w.f.Write(Encode(r)); err != nil {
		w.err = fmt.Errorf("wal: write: %w", err)
		return w.err
	}

	switch w.policy {
	case SyncAlways:
		if err := w.f.Sync(); err != nil {
			w.err = fmt.Errorf("wal: sync: %w", err)
			return w.err
		}
	case SyncBatch:
		w.dirty = true // syncLoop will sync it within 10 ms
	case SyncNever:
		// leave it to the OS
	}

	return nil
}

// Sync forces everything written so far onto the disk.
func (w *WAL) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.err != nil {
		return w.err // never retry a failed fsync (fsyncgate)
	}
	if err := w.f.Sync(); err != nil {
		w.err = fmt.Errorf("wal: sync: %w", err)
		return w.err
	}
	w.dirty = false
	return nil
}

// syncLoop runs in the background in batch mode, syncing every
// batchInterval if anything was written since the last sync.
func (w *WAL) syncLoop() {
	defer w.wg.Done()

	ticker := time.NewTicker(batchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.mu.Lock()
			dirty := w.dirty
			w.mu.Unlock()

			if dirty {
				_ = w.Sync() // any error is saved in w.err; the next Append reports it
			}
		case <-w.done:
			return
		}
	}
}

// Close stops background syncing, does a final sync, and closes the file.
func (w *WAL) Close() error {
	if w.policy == SyncBatch {
		close(w.done) // tell syncLoop to stop
		w.wg.Wait()   // wait until it actually has
	}

	syncErr := w.Sync()
	closeErr := w.f.Close()

	if syncErr != nil {
		return syncErr
	}
	return closeErr
}
