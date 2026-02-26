package jsonutil

import (
	"encoding/json"
	"io"
	"net/http"
)

func JSONResponse(response any, w http.ResponseWriter, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func JSONDecoder(r io.Reader, dec any) error {
	err := json.NewDecoder(r).Decode(dec)

	if err != nil {
		return err
	}
	return nil
}

func DecodeBody(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	return JSONDecoder(r.Body, dst)
}
