package xml

import (
	"bytes"
	"encoding/xml"
	"io"

	"github.com/jacklaaa89/pokeapi/internal/api/format"
)

const xmlContentType = "application/xml"

// formatter a pre-allocated instance of the XML formatter
var formatter format.Encoder = &xmlFormatter{}

// xmlFormatter a format.Encoder implementation which encodes and decodes
// using the encoding/xml library.
type xmlFormatter struct{}

func (f *xmlFormatter) ContentType() string { return xmlContentType }
func (f *xmlFormatter) Accept() string      { return xmlContentType }

// Decode implements format.Encoder interface.
// Decodes the supplied data in the io.Reader r into the receiver rcv
func (*xmlFormatter) Decode(r io.Reader, rcv interface{}) error { return xml.NewDecoder(r).Decode(rcv) }

// Encode implements the format.Encoder interface
// encodes the interface i using encoding/json
func (f *xmlFormatter) Encode(i interface{}) (io.Reader, error) {
	buf := &bytes.Buffer{}
	return buf, f.EncodeTo(buf, i)
}

// EncodeTo implements the format.Encoder interface
// encodes into the supplied writer using JSON.
func (*xmlFormatter) EncodeTo(w io.Writer, i interface{}) error {
	err := xml.NewEncoder(w).Encode(i)
	return err
}

// New returns the XML encoder.
func New() format.Encoder { return formatter }
