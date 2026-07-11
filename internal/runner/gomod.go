package runner

import (
	"os"
	"os/exec"
)

func GoModTidy(projectDir string) error {

	cmd := exec.Command("go", "mod", "tidy")

	cmd.Dir = projectDir

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
