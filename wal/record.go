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
	keyLen := len(r.Key)
	valLen := len(r.Value)

	buf := make([]byte, headerSize+keyLen+valLen)

	buf[4] = r.Op
	binary.LittleEndian.PutUint32(buf[5:9], uint32(keyLen))
	binary.LittleEndian.PutUint32(buf[9:13], uint32(valLen))

	copy(buf[headerSize:], r.Key)
	copy(buf[headerSize+keyLen:], r.Value)

	crc := crc32.ChecksumIEEE(buf[4:])
	binary.LittleEndian.PutUint32(buf[0:4], crc)

	return buf
}

// Decode reads exactly one record from rd.
// It returns the record, the number of bytes consumed, and:
//
//	io.EOF              clean end of log (zero bytes read)
//	io.ErrUnexpectedEOF partial record (torn write)
//	ErrCorrupt          checksum mismatch or absurd lengths
func Decode(rd io.Reader) (Record, int, error) {
	header := make([]byte, headerSize)
	if _, err := io.ReadFull(rd, header); err != nil {
		return Record{}, 0, err
	}

	op := header[4]
	keyLen := binary.LittleEndian.Uint32(header[5:9])
	valLen := binary.LittleEndian.Uint32(header[9:13])

	if uint64(keyLen)+uint64(valLen) > maxRecordSize {
		return Record{}, 0, ErrCorrupt
	}
	if op != OpSet && op != OpDel {
		return Record{}, 0, ErrCorrupt
	}

	body := make([]byte, int(keyLen)+int(valLen))
	if _, err := io.ReadFull(rd, body); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return Record{}, 0, err
	}

	want := binary.LittleEndian.Uint32(header[0:4])
	got := crc32.ChecksumIEEE(header[4:])
	got = crc32.Update(got, crc32.IEEETable, body)

	if got != want {
		return Record{}, 0, ErrCorrupt
	}

	r := Record{
		Op:    op,
		Key:   body[:keyLen],
		Value: body[keyLen:],
	}
	return r, headerSize + len(body), nil
}
