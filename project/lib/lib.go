package lib

import "io"

// Read reads data from the provided reader and returns
// the number of bytes read and any error encountered.
func Read(reader io.Reader) (int, error) {
	return reader.Read([]byte{})
}
