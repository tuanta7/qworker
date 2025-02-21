package pgrepo

type SchedulerRepository struct{}

func NewSchedulerRepository() *SchedulerRepository {
	return &SchedulerRepository{}
}

func (r *SchedulerRepository) LoadConfiguration() {
	// Create a job
}
