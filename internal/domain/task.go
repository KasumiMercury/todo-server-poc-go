package domain

type Task struct {
	id    string
	title string
}

func NewTask(id, title string) *Task {
	return &Task{
		id:    id,
		title: title,
	}
}

func (t *Task) ID() string {
	return t.id
}

func (t *Task) Title() string {
	return t.title
}
