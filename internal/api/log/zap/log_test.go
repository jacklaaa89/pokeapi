package zap

import (
	"bytes"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/jacklaaa89/pokeapi/internal/api/log"
)

// These tests are mainly for coverage, as we don't need to test any of the
// code in the zap package as that is out of scope, so we just test given an enabled logger, a log is rendered
// using zap.

var (
	zapConfig,
	outputBuffer = config()
)

type sink struct {
	*bytes.Buffer
}

func (*sink) Sync() error  { return nil }
func (*sink) Close() error { return nil }

// config initialises the zap configuration and returns
// the generated config and a buffer where output gets written to
//
// because a `sink` (a output source in zap) can only be registered once
// we need to make a global output buffer and ensure that its reset before
// tests are ran.
func config() (*zap.Config, *bytes.Buffer) {
	b := &bytes.Buffer{}
	err := zap.RegisterSink("test", func(*url.URL) (zap.Sink, error) {
		return &sink{b}, nil
	})

	if err != nil {
		panic("could not initialise zap config")
	}

	cfg := zap.NewProductionConfig()
	cfg.DisableStacktrace = true
	cfg.DisableCaller = true
	cfg.EncoderConfig.EncodeTime = func(_ time.Time, e zapcore.PrimitiveArrayEncoder) {
		// we want to freeze the log time so its always the same for tests
		t := time.Date(1970, time.January, 01, 00, 00, 00, 00, time.UTC)
		e.AppendInt64(t.Unix())
	}
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	cfg.OutputPaths = []string{"test://out"}
	cfg.ErrorOutputPaths = []string{"test://out"}
	return &cfg, b
}

// setup resets the output buffer and sets up the logger instance.
func setup(t *testing.T) log.Logger {
	outputBuffer.Reset()
	l, err := New(zapConfig)
	require.NoError(t, err)
	return l
}

func TestNew(t *testing.T) {
	l := setup(t)
	assert.Implements(t, (*log.Logger)(nil), l)
}

func TestLogger_Infof(t *testing.T) {
	const expectedData = `{"level":"info","ts":0,"msg":"message"}`
	l := setup(t)
	l.Infof("message")
	assert.Equal(t, expectedData, strings.TrimSpace(outputBuffer.String()))
}

func TestLogger_Errorf(t *testing.T) {
	const expectedData = `{"level":"error","ts":0,"msg":"message"}`
	l := setup(t)
	l.Errorf("message")
	assert.Equal(t, expectedData, strings.TrimSpace(outputBuffer.String()))
}

func TestLogger_Debugf(t *testing.T) {
	const expectedData = `{"level":"debug","ts":0,"msg":"message"}`
	l := setup(t)
	l.Debugf("message")
	assert.Equal(t, expectedData, strings.TrimSpace(outputBuffer.String()))
}

func TestLogger_Warnf(t *testing.T) {
	const expectedData = `{"level":"warn","ts":0,"msg":"message"}`
	l := setup(t)
	l.Warnf("message")
	assert.Equal(t, expectedData, strings.TrimSpace(outputBuffer.String()))
}
