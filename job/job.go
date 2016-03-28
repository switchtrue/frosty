package job

import (
	"strings"
	"time"

	"os/exec"

	"os"

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

func (js JobStatus) ElapsedTime() time.Duration {
	return js.EndTime.Sub(js.StartTime)
}

func (js JobStatus) IsSuccessful() bool {
	return js.Status == STATUS_SUCCESS
}

func Start(jobConfig config.JobConfig) JobStatus {

	jobDir, artefactDir := MakeJobDirectories(jobConfig.Name)

	os.Setenv("FROSTY_JOB_DIR", jobDir)
	os.Setenv("FROSTY_JOB_ARTIFACTS_DIR", artefactDir)

	jc := JobStatus{}

	jc.JobConfig = jobConfig
	jc.Status = STATUS_SUCCESS
	jc.StartTime = time.Now()

	out, err := exec.Command(jobConfig.Command).Output()
	if err != nil {
		jc.Status = STATUS_FAILURE
		jc.Error = strings.TrimSpace(err.Error())
	}

	jc.EndTime = time.Now()
	jc.Output = strings.TrimSpace(string(out[:]))

	return jc
}
