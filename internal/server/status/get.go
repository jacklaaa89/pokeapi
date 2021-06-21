package status

import (
	"net/http"
)

// Get http.HandlerFunc which handles /status
// this literally just returns a http.StatusOK
func Get(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }
