package wal

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
)

const (
	OpSet byte = 1
	OpDel byte = 2
)

// [crc32 4B][op 1B][keyLen 4B][valLen 4B][key][value]
const headerSize = 4 + 1 + 4 + 4

// Guards against a corrupted length field making you allocate 4 GB.
const maxRecordSize = 64 << 20 // 64 MiB

var ErrCorrupt = errors.New("wal: corrupt record")

type Record struct {
	Op    byte
	Key   []byte
	Value []byte // empty for OpDel
}

// Encode returns the on-disk bytes for r.
func Encode(r Record) []byte {
	// TODO 1: buf := make([]byte, headerSize+len(r.Key)+len(r.Value))
	// TODO 2: buf[4] = r.Op
	//         binary.LittleEndian.PutUint32(buf[5:9], keyLen)
	//         binary.LittleEndian.PutUint32(buf[9:13], valLen)
	// TODO 3: copy key, then value, after the header
	// TODO 4: crc := crc32.ChecksumIEEE(buf[4:])  // covers everything after the crc
	//         put it in buf[0:4]
	return nil
}

// Decode reads exactly one record from rd.
// It returns the record, the number of bytes consumed, and:
//
//	io.EOF              clean end of log (zero bytes read)
//	io.ErrUnexpectedEOF partial record (torn write)
//	ErrCorrupt          checksum mismatch or absurd lengths
func Decode(rd io.Reader) (Record, int, error) {
	// TODO 1: header := make([]byte, headerSize)
	//         _, err := io.ReadFull(rd, header)
	//         ReadFull already gives you the right errors here:
	//         io.EOF if 0 bytes were read, io.ErrUnexpectedEOF if only some were.
	// TODO 2: parse op, keyLen, valLen. If keyLen+valLen > maxRecordSize → ErrCorrupt.
	// TODO 3: body := make([]byte, keyLen+valLen); io.ReadFull(rd, body)
	//         GOTCHA: if the header was complete but the body is missing, ReadFull
	//         returns io.EOF. That's a torn write, not a clean end, so convert it
	//         to io.ErrUnexpectedEOF.
	// TODO 4: recompute the crc over header[4:] + body and compare → ErrCorrupt
	// TODO 5: return Record{...}, headerSize+len(body), nil
	return Record{}, 0, nil
}
