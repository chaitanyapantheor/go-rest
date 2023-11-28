package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/dimfeld/httptreemux/v5"
)

// Param returns the web call parameters from the request
func Param(r *http.Request, key string) string {
	m := httptreemux.ContextParams(r.Context())
	return m[key]
}

// Decode reads the body of an HTTP request looking for a JSON document. The
// body is decoded into the provided value
// If the provided value is a struct then it is checked for validation tags
func Decode(r *http.Request, val interface{}) error {
	if err := checkPayload(r); err != nil {
		return err
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(val); err != nil {
		// Checks if this is a bad key, or wrong value
		var re = regexp.MustCompile(`(?m)field ([A-Za-z-_\.]+) (of type [A-Za-z]+)`)
		matches := re.FindStringSubmatch(err.Error())
		if len(matches) == 3 {
			parts := strings.Split(matches[1], ".")
			err = fmt.Errorf("invalid json: %s must be %s", parts[len(parts)-1], matches[2])
			return NewRequestError(err, http.StatusBadRequest)
		}

		// Unknown Fields
		if strings.Contains(err.Error(), "unknown field") {
			str := strings.ReplaceAll(err.Error(), "\\", "")
			str = strings.ReplaceAll(str, "\"", "")
			str = strings.ReplaceAll(str, "unknown field", "unknown field:")
			return NewRequestError(fmt.Errorf(str), http.StatusBadRequest)
		}

		// Don't die on a decode failure
		return NewRequestError(err, http.StatusBadRequest)
	}

	return nil
}

func checkPayload(r *http.Request) error {

	// Need buffers to make sure we don't screw with the r.Body
	buf, _ := ioutil.ReadAll(r.Body)
	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
	rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))

	// Set the r.Body back to it's original state
	r.Body = rdr2

	// Check if empty for better error messages
	b, err := ioutil.ReadAll(rdr1)
	if err != nil {
		return err
	}

	// Empty Payload
	body := strings.TrimSpace(string(b))
	if body == "" {
		return NewRequestError(fmt.Errorf("json payload is empty"), http.StatusBadRequest)
	}

	// Missing first or last { brackets }
	if body[:1] != "{" || body[len(body)-1:] != "}" {
		return NewRequestError(fmt.Errorf("json missing opening or closing brackets"), http.StatusBadRequest)
	}

	return nil
}
