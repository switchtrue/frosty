package job

import (
	"strconv"
	"strings"
	"time"

	"os/exec"

	"os"

	"fmt"

	"github.com/mleonard87/frosty/artifact"
	"github.com/mleonard87/frosty/config"
)

const (
	STATUS_SUCCESS = iota
	STATUS_FAILURE = iota
	BYTES_PER_SI   = 1000
)

var BINARY_SI_UNITS = [...]string{"B", " kB", " MB", " GB", " TB", " PB", " EB", " ZB"}

type JobStatus struct {
	Status            int
	StdOut            string
	StdErr            string
	Error             string
	StartTime         time.Time
	EndTime           time.Time
	JobConfig         config.JobConfig
	ArchiveCreated    bool
	ArchiveSize       int64
	TransferStartTime time.Time
	TransferEndTime   time.Time
	TransferError     string
}

func (js JobStatus) ElapsedTime() time.Duration {
	return js.EndTime.Sub(js.StartTime)
}

func (js JobStatus) ElapsedTransferTime() time.Duration {
	return js.TransferEndTime.Sub(js.TransferStartTime)
}

func (js JobStatus) IsSuccessful() bool {
	return js.Status == STATUS_SUCCESS
}

func (js JobStatus) GetArchiveNameDisplay() string {
	return GetArtifactArchiveFileName(js.JobConfig.Name)
}

func (js JobStatus) GetArchiveSizeDisplay() string {
	size := js.ArchiveSize
	for _, unit := range BINARY_SI_UNITS {
		if size < 1024 {
			return strconv.FormatInt(size, 10) + unit
		} else {
			size = size / BYTES_PER_SI
		}
	}
	return strconv.FormatInt(js.ArchiveSize, 10) + BINARY_SI_UNITS[0]
}

func Start(jobConfig config.JobConfig) JobStatus {
	js := JobStatus{}
	js.JobConfig = jobConfig
	js.Status = STATUS_SUCCESS
	js.StartTime = time.Now()

	err := RemoveJobDirectory(jobConfig.Name)
	if err != nil {
		js.Status = STATUS_FAILURE
		js.Error = err.Error()
		js.EndTime = time.Now()
		return js
	}

	jobDir, artifactDir, err := MakeJobDirectories(jobConfig.Name)
	if err != nil {
		js.Status = STATUS_FAILURE
		js.Error = err.Error()
		js.EndTime = time.Now()
		return js
	}

	env := os.Environ()
	env = append(env, fmt.Sprintf("FROSTY_JOB_DIR=%s", jobDir))
	env = append(env, fmt.Sprintf("FROSTY_JOB_ARTIFACTS_DIR=%s", artifactDir))
	cmd := exec.Command(jobConfig.Command)
	cmd.Env = env

	out, err := cmd.Output()
	if err != nil {
		js.Status = STATUS_FAILURE

		if ee, ok := err.(*exec.ExitError); ok {
			// Capture and trim any errors logged to stderr.
			js.StdErr = strings.TrimSpace(string(ee.Stderr))
		}

		js.Error = err.Error()
		js.EndTime = time.Now()
		js.StdOut = strings.TrimSpace(string(out[:]))

		return js
	}

	js.EndTime = time.Now()
	js.StdOut = strings.TrimSpace(string(out[:]))

	archiveTarget := GetArtifactArchiveTargetName(jobConfig.Name)
	js.ArchiveCreated, err = artifact.MakeArtifactArchive(artifactDir, archiveTarget)
	if err != nil {
		js.Status = STATUS_FAILURE
		js.Error = err.Error()

		return js
	}

	if js.ArchiveCreated {
		fileInfo, err := os.Stat(archiveTarget)
		if err != nil {
			js.Status = STATUS_FAILURE
			js.Error = err.Error()

			return js
		}
		js.ArchiveSize = fileInfo.Size()
	}

	return js
}
