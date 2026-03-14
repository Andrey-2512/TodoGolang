package jsonrender

import (
	"encoding/json"
	"net/http"
)

func JSONResponse(response any, w http.ResponseWriter, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func DecodeBody(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	return json.NewDecoder(r.Body).Decode(dst)
}
