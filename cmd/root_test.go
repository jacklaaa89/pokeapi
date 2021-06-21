package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRoot tests that when the root command is ran
// the only thing it can really do is display information
// and that no error occurs.
func TestRoot(t *testing.T) {
	b := &bytes.Buffer{}
	cmd := Root()
	cmd.SetOut(b)
	cmd.SetArgs([]string{"--help"})
	assert.NoError(t, cmd.Execute())
	assert.NotEmpty(t, b.String())
}
