package json

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type receiver struct {
	Data string `json:"data"`
}

func TestNew(t *testing.T) {
	assert.Equal(t, formatter, New())
}

func TestJsonFormatter_Decode(t *testing.T) {
	const expectedData = `{"data":"12345"}`
	r, err := New().Encode(&receiver{Data: "12345"})
	assert.NoError(t, err)

	b, err := io.ReadAll(r)
	assert.Equal(t, expectedData, strings.TrimSpace(string(b)))
	assert.NoError(t, err)
}

func TestJsonFormatter_Encode(t *testing.T) {
	const expectedData = `{"data":"12345"}`
	var rcv = new(receiver)
	err := New().Decode(bytes.NewBuffer([]byte(expectedData)), rcv)
	assert.NoError(t, err)
	assert.Equal(t, "12345", rcv.Data)
}

func TestJsonFormatter_Accept(t *testing.T) {
	assert.Equal(t, jsonContentType, New().Accept())
}

func TestJsonFormatter_ContentType2(t *testing.T) {
	assert.Equal(t, jsonContentType, New().ContentType())
}
