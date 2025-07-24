package task

type Task struct {
	id     string
	title  string
	userId string // Note: This represents the userId from JWT sub claim, which may change in the future
}

func NewTask(id, title, userId string) *Task {
	return &Task{
		id:     id,
		title:  title,
		userId: userId,
	}
}

func (t *Task) ID() string {
	return t.id
}

func (t *Task) Title() string {
	return t.title
}

func (t *Task) UserID() string {
	return t.userId
}
