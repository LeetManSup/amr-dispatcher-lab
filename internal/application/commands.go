package application

type AddTaskInput struct {
	RequestID    string
	SourcePoint  string
	TargetPoint  string
	BusinessType string
	Priority     int
	CreatedAt    int
	Deadline     int
}
