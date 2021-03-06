package tsm1

// String encoding uses snappy compression to compress each string.  Each string is
// appended to byte slice prefixed with a variable byte length followed by the string
// bytes.  The bytes are compressed using snappy compressor and a 1 byte header is used
// to indicate the type of encoding.

import (
	"encoding/binary"
	"fmt"

	"github.com/golang/snappy"
)

const (
	// stringUncompressed is a an uncompressed format encoding strings as raw bytes.
	// Not yet implemented.
	stringUncompressed = 0

	// stringCompressedSnappy is a compressed encoding using Snappy compression
	stringCompressedSnappy = 1
)

type StringEncoder struct {
	// The encoded bytes
	bytes []byte
}

func NewStringEncoder() StringEncoder {
	return StringEncoder{}
}

func (e *StringEncoder) Write(s string) {
	b := make([]byte, 10)
	// Append the length of the string using variable byte encoding
	i := binary.PutUvarint(b, uint64(len(s)))
	e.bytes = append(e.bytes, b[:i]...)

	// Append the string bytes
	e.bytes = append(e.bytes, s...)
}

func (e *StringEncoder) Bytes() ([]byte, error) {
	// Compress the currently appended bytes using snappy and prefix with
	// a 1 byte header for future extension
	data := snappy.Encode(nil, e.bytes)
	return append([]byte{stringCompressedSnappy << 4}, data...), nil
}

type StringDecoder struct {
	b   []byte
	l   int
	i   int
	err error
}

// SetBytes initializes the decoder with bytes to read from.
// This must be called before calling any other method.
func (e *StringDecoder) SetBytes(b []byte) error {
	// First byte stores the encoding type, only have snappy format
	// currently so ignore for now.
	data, err := snappy.Decode(nil, b[1:])
	if err != nil {
		return fmt.Errorf("failed to decode string block: %v", err.Error())
	}

	e.b = data
	e.l = 0
	e.i = 0
	e.err = nil

	return nil
}

func (e *StringDecoder) Next() bool {
	e.i += e.l
	return e.i < len(e.b)
}

func (e *StringDecoder) Read() string {
	// Read the length of the string
	length, n := binary.Uvarint(e.b[e.i:])

	// The length of this string plus the length of the variable byte encoded length
	e.l = int(length) + n

	return string(e.b[e.i+n : e.i+n+int(length)])
}

func (e *StringDecoder) Error() error {
	return e.err
}
