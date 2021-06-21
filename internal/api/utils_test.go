package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// stringer type which implements the fmt.Stringer interface.
type stringer string

func (s stringer) String() string { return string(s) }

func TestFormatURLPath(t *testing.T) {
	var (
		stringValue = "12345"
		intValue    = 1
		stringPtr   = &stringValue
	)

	// a pointer which is not pointing to anything.
	var nilPtr *string

	tt := []struct {
		Name     string
		Path     string
		Args     []interface{}
		Expected string
	}{
		{"Int", "/%s/%s/%s/%s/%s", asInterfaceSlice(int(1), int8(2), int16(3), int32(4), int64(5)), "/1/2/3/4/5"},
		{"Uint", "/%s/%s/%s/%s/%s", asInterfaceSlice(uint(1), uint8(2), uint16(3), uint32(4), uint64(5)), "/1/2/3/4/5"},
		{"Float", "/%s/%s", asInterfaceSlice(float32(1.0), float64(2.0)), "/1/2"},
		{"Complex", "/%s", asInterfaceSlice(complex(1.0, 2.0)), "/1%2B2i"}, // /1+2i
		{"String", "/%s", asInterfaceSlice("test"), "/test"},
		{"Bool", "/%s/%s", asInterfaceSlice(true, false), "/true/false"},
		{"StringSlice", "/%s", asInterfaceSlice([]string{"123", "456"}), "/123%2C456"}, // /123,456
		{"IntSlice", "/%s", asInterfaceSlice([]int{1, 2, 3}), "/1%2C2%2C3"},            // /1,2,3
		{"BoolArray", "/%s", asInterfaceSlice([2]bool{true, false}), "/true%2Cfalse"},  // /true,false
		{"FloatArray", "/%s", asInterfaceSlice([2]float64{1.0, 2.0}), "/1%2C2"},        // /1,2
		{"fmt.Stringer", "/%s", asInterfaceSlice(stringer("12345")), "/12345"},
		{"StringPtr", "/%s", asInterfaceSlice(stringPtr), "/12345"},
		{"IntPtr", "/%s", asInterfaceSlice(&intValue), "/1"},
		{"StringPtrPtr", "/%s", asInterfaceSlice(&stringPtr), "/12345"},
		{"Nil", "/%s", asInterfaceSlice(nil), "/"},
		{"NilPtr", "/%s", asInterfaceSlice(nilPtr), "/"},

		// invalid type tests, invalid types are defined as an empty string if found.

		{"Chan", "/%s", asInterfaceSlice(make(chan bool)), "/"},
		{"Struct", "/%s", asInterfaceSlice(&dummyResponseBody{Data: "12345"}), "/"},
		{"InvalidTypesSkipped", "/%s/%s", asInterfaceSlice(make(chan bool), 1), "//1"},
		{"InvalidSliceType", "/%s", asInterfaceSlice([]interface{}{1, 2, make(chan bool)}), "/1%2C2"}, // /1,2
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			assert.Equal(st, tc.Expected, FormatURLPath(tc.Path, tc.Args...))
		})
	}
}

func TestNormalise(t *testing.T) {
	tt := []struct {
		Name     string
		Input    string
		Expected func(t *testing.T, out string)
	}{
		{
			Name:  "NoIllicitCharacters",
			Input: "this is standard text.",
			Expected: func(t *testing.T, out string) {
				assert.Equal(t, "this is standard text.", out)
			},
		},
		{
			Name:  "NoInput",
			Input: "",
			Expected: func(t *testing.T, out string) {
				assert.Equal(t, "", out)
			},
		},
		{
			Name:  "WithHTML",
			Input: "<script>console.log(`ooops`);</script><b>here is some text with HTML</b>",
			Expected: func(t *testing.T, out string) {
				assert.Equal(t, "here is some text with HTML", out)
			},
		},
		{
			Name:  "WithIllicitWhitespaceCharacters",
			Input: "\t\t\tthis is the first line\n this is the second\t line\n, this is the \fthird line   \n\t",
			Expected: func(t *testing.T, out string) {
				assert.Equal(t, "this is the first line this is the second line, this is the third line", out)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(st *testing.T) {
			tc.Expected(st, Normalise(tc.Input))
		})
	}
}

// asInterfaceSlice converts a set of variadic interface parameters into a slice.
func asInterfaceSlice(i ...interface{}) []interface{} { return i }
