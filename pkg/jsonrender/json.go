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
		w.WriteHeader(500)
		_, _ = w.Write([]byte(`{"detail": "Internal Error"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = buf.WriteTo(w)

}

func DecodeBody(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	return json.NewDecoder(r.Body).Decode(dst)
}
