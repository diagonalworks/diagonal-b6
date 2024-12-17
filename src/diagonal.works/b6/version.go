package b6

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"golang.org/x/mod/semver"
)

// ApiVersion is a semver 2.0.0 compliant version for the GRPC API exposed to
// the Python client library. The b6 backend may fail requests from a Python
// client library with a lower major version. We only support pre-release
// indicators of a, b or rc, since that's all Python allows.
// For consistency, we tie the version of a backend and client library build
// to this version, using AdvanceVersionFromGit.
const ApiVersion = "0.2.1"

// BackendVersion is a semver 2.0.0 compliant version for the backend binary,
// generated at build time by AdvanceVersionFromGit, and stamped into the
// binary by the linker. It won't be defined when b6 is included as a
// library.
var BackendVersion = ""

// Return a version for a backend and client library build, given the
// base API version, using a process similar to go modules. See
// https://go.dev/ref/mod#versions.
// If the current commit has been tagged with a valid semver, we
// use that, otherwise, we derive a semver compliant one from the
// api version and the git state my including the number of commits
// to main, and the latest commit hash. We end up with something like
// v0.0.2-4-gb57f6d1.
// If the latest commit is tagged with a release version, we ensure
// it matches the API version in the code, otherwise, we ensure the
// last tagged version doesn't exceed the current API version.
func AdvanceVersionFromGit() (string, error) {
	if !inGitRepo() {
		return ApiVersion, nil
	}
	binary, err := exec.LookPath("git")
	if err != nil {
		return "", fmt.Errorf("can't find git binary in PATH")
	}
	description, err := gitDescription(binary)
	if err != nil {
		return "", err
	}
	commits, err := gitCommits(binary)
	if err != nil {
		return "", err
	}
	return makeGoVersionFromGitOutput(ApiVersion, description, commits)
}

func gitDescription(binary string) (string, error) {
	cmd := exec.Command(binary, "describe", "--tags", "--match=v*", "--dirty=+modified")
	output, err := cmd.Output()
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("failed to get git description: %s", string(e.Stderr))
		} else {
			return "", fmt.Errorf("failed to get git description: %s", err)
		}
	}
	return strings.TrimSpace(string(output)), nil
}

func gitCommits(binary string) (string, error) {
	cmd := exec.Command(binary, "rev-list", "--count", "origin/main")
	output, err := cmd.Output()
	commits := strings.TrimSpace(string(output))
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("failed to count commits: %s", string(e.Stderr))
		} else {
			return "", fmt.Errorf("failed to count commits: %s", err)
		}
	}
	return commits, nil
}

func makeGoVersionFromGitOutput(api string, description string, commits string) (string, error) {
	v, err := makeVersionFromGitOutput(api, description, commits)
	if err != nil {
		return "", err
	}
	r := v.Base
	if v.Commits != "" {
		r = r + "." + v.Commits + "." + v.Hash
	}
	if v.Build != "" {
		r += v.Build
	}
	return r, nil
}

type version struct {
	Base    string
	Commits string
	Hash    string
	Build   string
}

func makeVersionFromGitOutput(api string, description string, commits string) (version, error) {
	if !semver.IsValid("v" + api) {
		return version{}, fmt.Errorf("api version %s isn't a valid semver", api)
	}
	if p := semver.Prerelease("v" + api); p != "" && p != "-a" && p != "-b" && p != "-rc" {
		return version{}, fmt.Errorf("api version %s uses a prelease indicator other than a, b or rc", api)
	}
	tag, exact := extractTagFromGitDescription(description)
	if !semver.IsValid(tag) {
		return version{}, fmt.Errorf("git tag %s isn't a valid semver", description)
	}
	build := semver.Build(description)
	if exact {
		if semver.Compare("v"+api, tag) != 0 {
			return version{}, fmt.Errorf("api version v%s doesn't match commit tagged %s", api, tag)
		}
		return version{Base: tag[1:], Build: build}, nil
	}
	if semver.Compare(tag, "v"+api) > 0 {
		return version{}, fmt.Errorf("commit tag version %s > api version v%s", tag, api)
	}
	hash := extractHashFromGitDescription(description)
	if p := semver.Prerelease("v" + api); p != "" {
		return version{Base: api, Commits: commits, Hash: hash, Build: build}, nil
	}
	patch, err := extractPatchFromSemver("v" + api)
	if err != nil {
		return version{}, err
	}
	return version{
		Base:    fmt.Sprintf("%s.%d-a", semver.MajorMinor("v" + api)[1:], patch+1),
		Commits: commits,
		Hash:    hash,
		Build:   build,
	}, nil
}

func extractTagFromGitDescription(description string) (string, bool) {
	if i := strings.LastIndex(description, "+"); i > 0 {
		description = description[0:i]
	}
	i := 0
	hyphens := 0
	for i = len(description) - 1; i >= 0; i-- {
		if description[i] == '-' {
			hyphens++
			if hyphens == 2 {
				break
			}
		}
	}
	if i < 0 {
		i = len(description) // nothing following the tag
	}
	return description[0:i], i == len(description)
}

func extractHashFromGitDescription(description string) string {
	if i := strings.LastIndex(description, "+"); i > 0 {
		description = description[0:i]
	}
	return description[strings.LastIndex(description, "-")+1:]
}

func extractPatchFromSemver(v string) (int, error) {
	suffix := v[len(semver.MajorMinor(v)):]
	if suffix == "" || suffix[0] != '.' {
		return 0, nil
	}
	suffix = suffix[1:] // Skip leading dot
	end := len(suffix)
	if i := strings.Index(suffix, "-"); i >= 0 {
		end = i
	}
	patch, err := strconv.Atoi(suffix[0:end])
	if err != nil {
		return 0, fmt.Errorf("can't extract patch from %s", v)
	}
	return patch, nil
}

// Return a version of AdvanceVersionFromGit that meets the
// PIP440 formatting rules.
// See: https://peps.python.org/pep-0440/
func AdvancePythonVersionFromGit() (string, error) {
	if !inGitRepo() {
		return ApiVersion, nil
	}
	binary, err := exec.LookPath("git")
	if err != nil {
		return "", fmt.Errorf("can't find git binary in PATH")
	}
	description, err := gitDescription(binary)
	if err != nil {
		return "", err
	}
	commits, err := gitCommits(binary)
	if err != nil {
		return "", err
	}
	return makePythonVersionFromGitOutput(ApiVersion, description, commits)
}

func makePythonVersionFromGitOutput(api string, description string, commits string) (string, error) {
	v, err := makeVersionFromGitOutput(api, description, commits)
	if err != nil {
		return "", err
	}
	r := v.Base
	if v.Commits != "" {
		r = strings.Replace(r, "-", "", 1)
		r = r + v.Commits + "+" + v.Hash
	}
	return r, nil
}

// Return true if the current directory is the root of a git repo
func inGitRepo() bool {
	s, err := os.Stat(".git")
	return err == nil && s.IsDir()
}
