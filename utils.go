package main

import (
	"fmt"
	"os/exec"
	"regexp"

	"github.com/google/shlex"
)

func retryBackOff(t func() error, retries int) {
	for range retries {
		if err := t(); err == nil {
			return
		}
	}
}

func execCommand(cmd string) ([]byte, error) {
	parts, err := shlex.Split(cmd)
	if err != nil {
		return nil, err
	}

	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	return exec.Command(parts[0], parts[1:]...).Output()
}

func getRegexpSubmatch(re *regexp.Regexp, b []byte, index int) ([]byte, error) {
	if re == nil {
		return nil, fmt.Errorf("invalid regex")
	} else if matches := re.FindSubmatch(b); len(matches)-1 >= index {
		return matches[index], nil
	}

	return nil, fmt.Errorf("no matches for the supplied capture group found")
}
