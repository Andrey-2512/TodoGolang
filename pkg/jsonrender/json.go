package jsonrender

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func JSONResponse(response any, w http.ResponseWriter, statusCode int) {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(response); err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"detail": "Internal Error"}`, 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	buf.WriteTo(w)

}

func DecodeBody(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	return json.NewDecoder(r.Body).Decode(dst)
}
