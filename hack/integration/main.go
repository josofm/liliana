package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

func main() {
	status := 0

	if err := run("docker", "compose", "-f", "docker-compose.yaml", "up", "-d", "postgres"); err != nil {
		status = exitCode(err)
	}

	if status == 0 {
		err := run(
			"docker", "compose", "-f", "docker-compose.yaml",
			"run", "--rm", "--build", "--no-deps",
			"liliana", "go", "test", "-race", "-timeout", "60s", "-tags", "integration", "./...",
		)
		if err != nil {
			status = exitCode(err)
		}
	}

	if err := run("docker", "compose", "-f", "docker-compose.yaml", "down", "--remove-orphans"); err != nil && status == 0 {
		status = exitCode(err)
	}

	os.Exit(status)
}

func run(name string, args ...string) error {
	fmt.Printf("+ %s %s\n", name, joinArgs(args))

	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func exitCode(err error) int {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}

	fmt.Fprintln(os.Stderr, err)
	return 1
}

func joinArgs(args []string) string {
	result := ""
	for i, arg := range args {
		if i > 0 {
			result += " "
		}
		result += arg
	}
	return result
}
