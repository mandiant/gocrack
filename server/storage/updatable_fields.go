package storage

// ModifiableTaskRequest defines the fields in `Task` that are allowed to be modified
type ModifiableTaskRequest struct {
	AssignedToHost    *string
	AssignedToDevices *CLDevices
	Status            *TaskStatus
	TaskDuration      *int
}

// UserModifyRequest contains the fields in `User` that are allowed to be modified
type UserModifyRequest struct {
	Password    *string
	UserIsAdmin *bool
	Email       *string
}
