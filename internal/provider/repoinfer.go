package provider

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

var (
	httpsPattern = regexp.MustCompile(`github\.com[/:]([^/]+/[^/]+?)(?:\.git)?$`)
	sshPattern   = regexp.MustCompile(`git@github\.com:([^/]+/[^/]+?)(?:\.git)?$`)
)

func InferRepo() (string, error) {
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return "", ErrNoRepo
	}

	return ParseRemote(strings.TrimSpace(string(out)))
}

func ParseRemote(remote string) (string, error) {
	if m := sshPattern.FindStringSubmatch(remote); len(m) == 2 {
		return m[1], nil
	}

	if m := httpsPattern.FindStringSubmatch(remote); len(m) == 2 {
		return m[1], nil
	}

	return "", fmt.Errorf("cannot infer repo from remote %q: not a GitHub remote; pass --repo owner/repo", remote)
}
