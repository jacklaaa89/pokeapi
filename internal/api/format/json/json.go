package json

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/jacklaaa89/pokeapi/internal/api/format"
)

const jsonContentType = "application/json"

// formatter a pre-allocated instance of the JSON formatter
var formatter format.Encoder = &jsonFormatter{}

// jsonFormatter a format.Encoder implementation which encodes and decodes
// using the encoding/json library.
type jsonFormatter struct{}

func (f *jsonFormatter) ContentType() string { return jsonContentType }
func (f *jsonFormatter) Accept() string      { return jsonContentType }

// Decode implements format.Encoder interface.
// Decodes the supplied data in the io.Reader r into the receiver rcv
func (*jsonFormatter) Decode(r io.Reader, rcv interface{}) error {
	return json.NewDecoder(r).Decode(rcv)
}

// Encode implements the format.Encoder interface
// encodes the interface i using encoding/json
func (j *jsonFormatter) Encode(i interface{}) (io.Reader, error) {
	buf := &bytes.Buffer{}
	return buf, j.EncodeTo(buf, i)
}

// EncodeTo implements the format.Encoder interface
// encodes into the supplied writer using JSON.
func (*jsonFormatter) EncodeTo(w io.Writer, i interface{}) error {
	err := json.NewEncoder(w).Encode(i)
	return err
}

// New returns the JSON encoder.
func New() format.Encoder { return formatter }
