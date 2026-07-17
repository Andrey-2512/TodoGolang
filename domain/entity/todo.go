package entity

type Task struct {
	Id          int
	Title       *string
	Description *string
	UserId      int
}

type PatchTask struct {
	Id          int
	Title       Optional[string]
	Description Optional[string]
	UserId      int
}
