package workmgr

import "github.com/fireeye/gocrack/shared"

// Config defines all the configuration settings for the Work/Job Manager
type Config struct {
	CheckForInactiveWorkers *shared.HumanDuration `yaml:"inactive_workers_check_interval,omitempty"`
	StopTasksAfter          *shared.HumanDuration `yaml:"stop_tasks_after_interval,omitempty"`
}
