package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type FrostyConfig struct {
	WorkingDirectory string `json:"workingDirectory"`
	Jobs             []Job  `json:"jobs"`
}

type Job struct {
	Name                 string   `json:"name"`
	EnvironmentVariables []string `json:"envVars"`
	Script               []string `json:"script"`
	Archives             []string `json:"archives"`
}

func Read(filePath string) {
	configFile, e := ioutil.ReadFile(filePath)
	if e != nil {
		fmt.Printf("Could not open configuration file: %q.\n%s\n", filePath, e)
		os.Exit(1)
	}
	fmt.Printf("%s\n", string(configFile))

	var frostyConfig FrostyConfig
	json.Unmarshal(configFile, &frostyConfig)
	fmt.Printf("Results: %v\n", frostyConfig)
}
