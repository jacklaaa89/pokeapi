package format

import (
	"io"
)

// Encoder is an interface which allows us to interchange the way we encode
// and decode the request and response bodies.
type Encoder interface {
	ContentType() string                       // ContentType returns the Content-Type header
	Accept() string                            // Accept returns the Accept header.
	Decode(r io.Reader, rcv interface{}) error // Decode takes a response body and decodes it into the receiver rcv.
	Encode(i interface{}) (io.Reader, error)   // Encode takes the request body and encodes it into the correct
	EncodeTo(w io.Writer, i interface{}) error // EncodeTo similar to Encode except we attempt to encode to the supplied writer.
}
