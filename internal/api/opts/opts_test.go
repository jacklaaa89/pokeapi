package opts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApply(t *testing.T) {
	ao := newAPIOption(func(options *Options) {
		options.MaxNetworkRetries = 1
	})
	assert.Equal(t, int64(1), Apply(ao).MaxNetworkRetries) // ensure the value has been set.
}
