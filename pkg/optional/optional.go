package optional

import (
	"bytes"
	"encoding/json"
)

type Optional[T any] struct {
	Val  T
	Set  bool
	Null bool
}

func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	o.Set = true

	if bytes.Equal(data, []byte("null")) {
		o.Null = true
		return nil
	}

	if err := json.Unmarshal(data, &o.Val); err != nil {
		return err
	}

	return nil
}
