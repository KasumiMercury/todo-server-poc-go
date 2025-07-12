package domain

type Task struct {
	id   string
	name string
}

func NewTask(id, name string) *Task {
	return &Task{
		id:   id,
		name: name,
	}
}

func (t *Task) ID() string {
	return t.id
}

func (t *Task) Name() string {
	return t.name
}
