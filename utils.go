package main

import (
	"fmt"
	"os/exec"
	"regexp"

	"github.com/google/shlex"
)

func execCommand(cmd string) error {
	parts, err := shlex.Split(cmd)
	if err != nil {
		return err
	}
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}
	_, err = exec.Command(parts[0], parts[1:]...).Output()
	return err
}

func getRegexpSubmatch(re *regexp.Regexp, b []byte, index int) ([]byte, error) {
	if re == nil {
		return nil, fmt.Errorf("invalid regex")
	} else if matches := re.FindSubmatch(b); len(matches)-1 >= index {
		return matches[index], nil
	}

	return nil, fmt.Errorf("no matches for the supplied capture group found")
}
