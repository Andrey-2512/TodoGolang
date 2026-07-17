package optionalconv

import (
	"todo/domain/entity"
	"todo/pkg/optional"
)

func FromJSONToEntity[T any](opt optional.Optional[T]) entity.Optional[T] {
	return entity.Optional[T]{
		Val:  opt.Val,
		Set:  opt.Set,
		Null: opt.Null,
	}
}
