package job

import (
	"log"
	"strconv"
	"strings"
	"time"

	"os/exec"

	"os"

	"github.com/mleonard87/frosty/artifact"
	"github.com/mleonard87/frosty/config"
)

const (
	STATUS_SUCCESS = iota
	STATUS_FAILURE = iota
	BYTES_PER_SI   = 1000
)

var BINARY_SI_UNITS = [...]string{"b", "kb", "mb", "gb", "tb", "pb", "eb", "zb"}

type JobStatus struct {
	Status            int
	Output            string
	Error             string
	StartTime         time.Time
	EndTime           time.Time
	JobConfig         config.JobConfig
	ArchiveCreated    bool
	ArchiveSize       int64
	TransferStartTime time.Time
	TransferEndTime   time.Time
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
	RemoveJobDirectory(jobConfig.Name)
	jobDir, artifactDir := MakeJobDirectories(jobConfig.Name)

	os.Setenv("FROSTY_JOB_DIR", jobDir)
	os.Setenv("FROSTY_JOB_ARTIFACTS_DIR", artifactDir)

	js := JobStatus{}

	js.JobConfig = jobConfig
	js.Status = STATUS_SUCCESS
	js.StartTime = time.Now()

	out, err := exec.Command(jobConfig.Command).Output()
	if err != nil {
		js.Status = STATUS_FAILURE
		js.Error = strings.TrimSpace(err.Error())
	}

	js.EndTime = time.Now()
	js.Output = strings.TrimSpace(string(out[:]))

	archiveTarget := GetArtifactArchiveTargetName(jobConfig.Name)
	js.ArchiveCreated = artifact.MakeArtifactArchive(artifactDir, archiveTarget)

	if js.ArchiveCreated {
		fileInfo, err := os.Stat(archiveTarget)
		if err != nil {
			log.Fatal(err)
		}
		js.ArchiveSize = fileInfo.Size()
	}

	return js
}
