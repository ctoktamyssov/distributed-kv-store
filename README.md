# kvstore

A key-value store in Go, built in stages from a single-node
in-memory server up to a replicated Raft cluster.

## Running

    go run . -port 6379
    ncat localhost 6379

## Protocol

| Command         | Reply                      |
|-----------------|----------------------------|
| SET key value   | OK                         |
| GET key         | value, or NOT_FOUND        |
| DEL key         | OK                         |
| anything else   | ERR unknown command        |

## Stages

- [X] Stage 1: in-memory map over TCP
- [ ] Stage 2: write-ahead log
- [ ] Stage 3: compaction
- [ ] Stage 4: on-disk values (Bitcask-style)
- [ ] Stage 5: leader election
- [ ] Stage 6: log replication
- [ ] Stage 7: chaos testing

## Design decisions

- **One goroutine per connection.** A blocked read on one client
  shouldn't stall the others, and goroutines are cheap enough to
  spend one per client.
- **Mutex lives inside `Store`, not the server.** No caller can forget
  to lock, and the store stays reusable when the log (Stage 2) and
  Raft (Stage 6) wrap it.
- **Plain `sync.Mutex` for now.** A map lookup holds the lock for
  nanoseconds, so `RWMutex` would add little. Revisit with a benchmark.
- **Values with spaces:** 

## What broke

`TestConcurrent` runs 100 goroutines doing `Set`/`Get`/`Delete` on 10
shared keys. With the locks removed it crashes:

```
fatal error: concurrent map writes
```

The fix was a `sync.Mutex` in `Store`, taken at the top of every
method with `defer Unlock()`. Reads need the lock too, since a read
concurrent with a write is still a race.