package entity

type Optional[T any] struct {
	Val  T
	Set  bool
	Null bool
}
