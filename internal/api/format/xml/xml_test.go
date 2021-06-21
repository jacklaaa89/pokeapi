package xml

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type receiver struct {
	Data string `xml:"data"`
}

func TestNew(t *testing.T) {
	assert.Equal(t, formatter, New())
}

func TestXmlFormatter_Decode(t *testing.T) {
	const expectedData = `<receiver><data>12345</data></receiver>`
	r, err := New().Encode(&receiver{Data: "12345"})
	assert.NoError(t, err)

	b, err := io.ReadAll(r)
	assert.Equal(t, expectedData, strings.TrimSpace(string(b)))
	assert.NoError(t, err)
}

func TestXmlFormatter_Encode(t *testing.T) {
	const expectedData = `<receiver><data>12345</data></receiver>`
	var rcv = new(receiver)
	err := New().Decode(bytes.NewBuffer([]byte(expectedData)), rcv)
	assert.NoError(t, err)
	assert.Equal(t, "12345", rcv.Data)
}

func TestXmlFormatter_Accept(t *testing.T) {
	assert.Equal(t, xmlContentType, New().Accept())
}

func TestXmlFormatter_ContentType2(t *testing.T) {
	assert.Equal(t, xmlContentType, New().ContentType())
}
