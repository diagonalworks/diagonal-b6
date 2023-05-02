package b6

import (
	"testing"
)

func TestMakeVersionFromGitOutput(t *testing.T) {
	tests := []struct {
		ApiVersion     string
		GitDescription string
		Commits        string
		Expected       string
		OK             bool
	}{
		// A repository at a commit with a version tag
		{"0.0.2", "v0.0.2", "317", "0.0.2", true},
		// A repository at a commit with a version tag and build metadata
		{"0.0.2", "v0.0.2+modified", "317", "0.0.2+modified", true},
		{"0.0.2", "v0.0.2+modified-andrew", "317", "0.0.2+modified-andrew", true},
		// A repository modified since the last version tag, when the version
		// in the code has been incremented
		{"0.0.1-alpha", "v0.0.0-4-gb57f6d1+modified", "317", "0.0.1-alpha.317.gb57f6d1+modified", true},
		{"0.0.1-alpha", "v0.0.0-4-gb57f6d1+modified-andrew", "317", "0.0.1-alpha.317.gb57f6d1+modified-andrew", true},
		// A repository modified since the last version tag, when the version
		// in the code hasn't yet been incremented
		{"0.0.0", "v0.0.0-4-gb57f6d1+modified", "317", "0.0.1-pre.317.gb57f6d1+modified", true},
		// The version from git tag can't be ahead of the version in the code
		{"0.0.1", "v0.0.2-4-gb57f6d1+modified", "317", "", false},
		// The version from an unmodified git tag must match the version in the code
		{"0.0.3", "v0.0.2", "317", "", false},
		// The tag from git describe must be a valid semver
		{"0.0.3", "vx.0.3-4-gb57f6d1+modified", "317", "", false},
	}
	for _, c := range tests {
		version, err := makeVersionFromGitOutput(c.ApiVersion, c.GitDescription, c.Commits)
		if err != nil && c.OK {
			t.Errorf("Expected no error, found: %s for api version %s and git description %s", err, c.ApiVersion, c.GitDescription)
		} else if err == nil && !c.OK {
			t.Errorf("Expected error for api version %s and git description %s, found none", c.ApiVersion, c.GitDescription)
		} else if version != c.Expected {
			t.Errorf("Expected %q, found %q", c.Expected, version)
		}
	}
}
