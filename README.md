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

- [ ] Stage 1: in-memory map over TCP
- [ ] Stage 2: write-ahead log
- [ ] Stage 3: compaction
- [ ] Stage 4: on-disk values (Bitcask-style)
- [ ] Stage 5: leader election
- [ ] Stage 6: log replication
- [ ] Stage 7: chaos testing

## Design decisions

<!-- Write these as you go. Why a mutex? Why inside Store
     rather than in the server? What happens to values with spaces? -->

## What broke

<!-- Bugs you hit and what they taught you. Interviewers love this. -->