package b6

import (
	"fmt"
	"os/exec"
	"os/user"
	"strconv"
	"strings"

	"golang.org/x/mod/semver"
)

// ApiVersion is a semver 2.0.0 compliant version for the GRPC API exposed to
// the Python client library. The Python client library may fail requests to
// at b6 backend with a major version number that's different to the one it
// was built with.
// For consistency, we tie the version of software releases to this version,
// generating the version number itself from the b6-api binary using the
// functions that process the git output below. Version generation roughly
// follows the scheme described here for go modules:
// https://go.dev/ref/mod#versions
// We enforce that the repository can't be tagged with a version higher
// than specified here at build time.
const ApiVersion = "0.0.1-pre"

// Return a version for the software release, based on the api version
// and git status. Adds build metadata if withMeta is true. Returns
// a semver like 0.0.1-alpha.317.gb57f6d1.
func MakeVersionFromGit(withMeta bool) (string, error) {
	binary, err := exec.LookPath("git")
	if err != nil {
		return "", fmt.Errorf("can't find git binary in PATH")
	}
	var cmd *exec.Cmd
	if withMeta {
		cmd = exec.Command(binary, "describe", "--tags", "--match=v*", "--dirty=+modified")
	} else {
		cmd = exec.Command(binary, "describe", "--tags", "--match=v*")
	}
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git description: %s", err)
	}
	description := strings.TrimSpace(string(output))
	if withMeta {
		if u, err := user.Current(); err == nil {
			if i := strings.LastIndex(description, "+"); i >= 0 {
				description = description + "-" + u.Username
			} else {
				description = description + "+" + u.Username
			}
		}
	}
	cmd = exec.Command(binary, "rev-list", "--count", "main")
	output, err = cmd.Output()
	commits := strings.TrimSpace(string(output))
	if err != nil {
		return "", fmt.Errorf("failed to count commits: %s", err)
	}
	return makeVersionFromGitOutput(ApiVersion, description, string(commits))
}

func makeVersionFromGitOutput(api string, description string, commits string) (string, error) {
	if !semver.IsValid("v" + api) {
		return "", fmt.Errorf("api version %s isn't a valid semver", api)
	}
	tag, exact := extractTagFromGitDescription(description)
	if !semver.IsValid(tag) {
		return "", fmt.Errorf("git tag %s isn't a valid semver", description)
	}
	build := semver.Build(description)
	if exact {
		if semver.Compare("v"+api, tag) != 0 {
			return "", fmt.Errorf("api version v%s doesn't match commit tagged %s", api, tag)
		}
		return tag[1:] + build, nil
	}
	if semver.Compare(tag, "v"+api) > 0 {
		return "", fmt.Errorf("commit tag version %s > api version v%s", tag, api)
	}
	hash := extractHashFromGitDescription(description)
	if p := semver.Prerelease("v" + api); p != "" {
		return fmt.Sprintf("%s.%s.%s%s", api, commits, hash, build), nil
	}
	patch, err := extractPatchFromSemver("v" + api)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s.%d-pre.%s.%s%s", semver.MajorMinor("v" + api)[1:], patch+1, commits, hash, build), nil
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
