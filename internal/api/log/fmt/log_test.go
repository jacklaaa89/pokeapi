package fmt

import (
	"bytes"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jacklaaa89/pokeapi/internal/api/log"
)

var emptyString = regexp.MustCompile(`^$`)

func testRegex(level, message string) *regexp.Regexp {
	return regexp.MustCompile(`\[` + level + `\] ([\w/]+ [\w:]+) ` + message)
}

func TestNewWithOutputs(t *testing.T) {
	b := &bytes.Buffer{}
	l := NewWithOutputs(LevelNone, b, b)
	assert.Implements(t, (*log.Logger)(nil), l)
	assert.Equal(t, LevelNone, l.(*logger).level)
	assert.Equal(t, b, l.(*logger).out)
	assert.Equal(t, b, l.(*logger).err)
}

func TestNew(t *testing.T) {
	tt := []struct {
		Name     string
		Input    Level
		Expected Level
	}{
		{
			Name:     "ValidLevel",
			Input:    LevelNone,
			Expected: LevelNone,
		},
		{
			Name:     "InvalidLevel",
			Input:    Level(25),
			Expected: LevelDebug,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			l := New(tc.Input)
			assert.Implements(t, (*log.Logger)(nil), l)
			assert.Equal(t, tc.Expected, l.(*logger).level)
			assert.Equal(t, os.Stdout, l.(*logger).out)
			assert.Equal(t, os.Stderr, l.(*logger).err)
		})
	}
}

func TestLogger_Debugf(t *testing.T) {
	tt := []struct {
		Name   string
		Level  Level
		Regexp *regexp.Regexp
	}{
		{"None", LevelNone, emptyString},
		{"Error", LevelError, emptyString},
		{"Warn", LevelWarn, emptyString},
		{"Info", LevelInfo, emptyString},
		{"Debug", LevelDebug, testRegex("DEBUG", "message")},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			b := &bytes.Buffer{}
			l := NewWithOutputs(tc.Level, b, b)
			l.Debugf("message")
			assert.True(st, tc.Regexp.MatchString(strings.TrimSpace(b.String())))
		})
	}
}

func TestLogger_Errorf(t *testing.T) {
	tt := []struct {
		Name   string
		Level  Level
		Regexp *regexp.Regexp
	}{
		{"None", LevelNone, emptyString},
		{"Error", LevelError, testRegex("ERROR", "message")},
		{"Warn", LevelWarn, testRegex("ERROR", "message")},
		{"Info", LevelInfo, testRegex("ERROR", "message")},
		{"Debug", LevelDebug, testRegex("ERROR", "message")},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			b := &bytes.Buffer{}
			l := NewWithOutputs(tc.Level, b, b)
			l.Errorf("message")
			assert.True(st, tc.Regexp.MatchString(strings.TrimSpace(b.String())))
		})
	}
}

func TestLogger_Infof(t *testing.T) {
	tt := []struct {
		Name   string
		Level  Level
		Regexp *regexp.Regexp
	}{
		{"None", LevelNone, emptyString},
		{"Error", LevelError, emptyString},
		{"Warn", LevelWarn, emptyString},
		{"Info", LevelInfo, testRegex("INFO", "message")},
		{"Debug", LevelDebug, testRegex("INFO", "message")},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			b := &bytes.Buffer{}
			l := NewWithOutputs(tc.Level, b, b)
			l.Infof("message")
			assert.True(st, tc.Regexp.MatchString(strings.TrimSpace(b.String())))
		})
	}
}

func TestLogger_Warnf(t *testing.T) {
	tt := []struct {
		Name   string
		Level  Level
		Regexp *regexp.Regexp
	}{
		{"None", LevelNone, emptyString},
		{"Error", LevelError, emptyString},
		{"Warn", LevelWarn, testRegex("WARN", "message")},
		{"Info", LevelInfo, testRegex("WARN", "message")},
		{"Debug", LevelDebug, testRegex("WARN", "message")},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			b := &bytes.Buffer{}
			l := NewWithOutputs(tc.Level, b, b)
			l.Warnf("message")
			assert.True(st, tc.Regexp.MatchString(strings.TrimSpace(b.String())))
		})
	}
}
