package job

import (
	"time"

	"os/exec"

	"github.com/mleonard87/frosty/config"
)

const (
	STATUS_SUCCESS = iota
	STATUS_FAILURE = iota
)

type JobStatus struct {
	Status    int
	Output    string
	Error     string
	StartTime time.Time
	EndTime   time.Time
	JobConfig config.JobConfig
}

func (js *JobStatus) ElapsedTime() time.Duration {
	return js.EndTime.Sub(js.StartTime)
}

func Start(jobConfig config.JobConfig) JobStatus {
	jc := JobStatus{}

	jc.JobConfig = jobConfig
	jc.Status = STATUS_SUCCESS
	jc.StartTime = time.Now()

	out, err := exec.Command(jobConfig.Command).Output()
	if err != nil {
		jc.Status = STATUS_FAILURE
		jc.Error = err.Error()
	}

	jc.EndTime = time.Now()
	jc.Output = string(out[:])

	return jc
}
