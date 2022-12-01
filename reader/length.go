package reader

import "io"

// LenReader is an interface implemented by many in-memory io.Reader's. Used
// for automatically sending the right Content-Length header when possible.
type LenReader interface {
	Len() int
}

// GetLength returns Length of reader using LenReader Interface
func GetLength(reader io.Reader) (int64, bool) {
	len, ok := reader.(LenReader)
	return int64(len.Len()), ok

}
