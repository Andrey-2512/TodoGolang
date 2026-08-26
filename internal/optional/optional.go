package optional

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
)

type Optional[T any] struct {
	Val  T
	Set  bool
	Null bool
}

func (o *Optional[T]) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	o.Set = true

	if dec.PeekKind() == 'n' {
		if _, err := dec.ReadToken(); err != nil {
			return err
		}
		o.Null = true
		return nil
	}

	if err := json.UnmarshalDecode(dec, &o.Val); err != nil {
		return err
	}

	return nil
}
