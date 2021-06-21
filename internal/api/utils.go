package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var (
	specialCharsRegex = regexp.MustCompile(`[\t\r\n\f\v]+`) // use to match any special whitespace chars.
	whitespaceRegex   = regexp.MustCompile(`\s{2,}`)        // use to match any collections of 2 or more white-space chars.
	invalidSpaceRegex = regexp.MustCompile(`\s(\W+)`)       // use to match a white space char next to a non-word literal.
)

// regexList the list of regexps to apply when normalising text.
var regexList = []*regexp.Regexp{
	specialCharsRegex,
	whitespaceRegex,
}

// isHTTPWriteMethod determines if the HTTP supplied is a mutation type.
func isHTTPWriteMethod(method string) bool {
	return method == http.MethodPost || method == http.MethodPut ||
		method == http.MethodPatch || method == http.MethodDelete
}

// FormatURLPath takes a format string (of the kind used in the fmt package)
// representing a URL path with a number of parameters that belong in the path
// and returns a formatted string.
//
// we perform a quick type comparison to handle some well-known types
// including a fmt.Stringer or string or any primitive int/uint type, it any of the
// parameters are not of these types, then we ignore the value
// as it should only be of of these three types.
// this would cause come very noticeable formatting helpers.
func FormatURLPath(format string, params ...interface{}) string {
	// Convert parameters to interface{} and URL-escape them
	untypedParams := make([]interface{}, len(params))
	var i int
	for _, param := range params {
		formatted, err := formatInterface(param)
		if err != nil {
			formatted = ""
		}

		untypedParams[i] = interface{}(url.QueryEscape(formatted))
		i++
	}

	return fmt.Sprintf(format, untypedParams[:i]...)
}

// formatInterface attempts to format an interface into a string.
func formatInterface(i interface{}) (string, error) {
	if i == nil {
		return "", nil
	}

	v := reflect.ValueOf(i)

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "", nil
		}
		return formatInterface(v.Elem().Interface())
	}

	// handle special types.
	switch p := i.(type) {
	case fmt.Stringer:
		return p.String(), nil
	}

	// handle primitive types.
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		return formatSlice(v.Interface()), nil
	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'g', -1, 64), nil
	case reflect.Complex64, reflect.Complex128:
		c := v.Complex()
		return strconv.FormatFloat(real(c), 'g', -1, 64) +
			"+" + strconv.FormatFloat(imag(c), 'g', -1, 64) + "i", nil
	case reflect.String:
		return v.String(), nil
	case reflect.Bool:
		if v.Bool() {
			return "true", nil
		}
		return "false", nil
	}

	return "", errors.New("invalid interface type")
}

// formatSlice attempts to format a slice into a comma separated
// joined string.
func formatSlice(slice interface{}) string {
	v := reflect.ValueOf(slice)

	stringSlice := make([]string, v.Len())
	var idx int
	for i := 0; i < v.Len(); i++ {
		f := v.Index(i)
		str, err := formatInterface(f.Interface())
		if err != nil {
			continue
		}
		stringSlice[idx] = str
		idx++
	}

	return strings.Join(stringSlice[:idx], ",")
}

// Normalise replaces new line characters etc with spaces.
func Normalise(input string) string {
	if input == "" {
		return ""
	}

	// strip HTML/JS from input.
	p := bluemonday.StrictPolicy()
	input = p.Sanitize(input)

	for _, r := range regexList {
		input = r.ReplaceAllString(input, " ")
	}

	input = invalidSpaceRegex.ReplaceAllString(input, "$1")
	return strings.TrimSpace(input)
}
