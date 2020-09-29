package redhat

import (
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"strings"
	"time"
)

const (
	// ReleaseCommand is the Unix command used to check whether a container is based on Red Hat technologies.
	ReleaseCommand = "if [ -e /etc/redhat-release ]; then cat /etc/redhat-release; else echo \"Unknown Base Image\"; fi"
	// NotRedHatBasedRegex is the expected output for a container that is not based on Red Hat technologies.
	NotRedHatBasedRegex = `(?m)Unknown Base Image`
	// VersionRegex is regular expression expected for a container based on Red Hat technologies.
	VersionRegex = `(?m)Red Hat Enterprise Linux Server release (\d+\.\d+) \(\w+\)`
)

// Release is an implementation of tnf.Test used to determine whether a container is based on Red Hat technologies.
type Release struct {
	// result is the result of the test.
	result int
	// timeout is the timeout duration for the test.
	timeout time.Duration
	// args stores the command and arguments.
	args []string
	// release contains the contents of /etc/redhat-release if it exists, or "NOT Red Hat Based" if it does not exist.
	release string
	// isRedHatBased contains whether the container is based on Red Hat technologies.
	isRedHatBased bool
}

// Args returns the command line arguments for the test.
func (r *Release) Args() []string {
	return r.args
}

// Timeout returns the timeout for the test.
func (r *Release) Timeout() time.Duration {
	return r.timeout
}

// Result returns the test result.
func (r *Release) Result() int {
	return r.result
}

// ReelFirst returns a reel.Step which expects output from running the Args command.
func (r *Release) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{VersionRegex, NotRedHatBasedRegex},
		Timeout: r.timeout,
	}
}

// ReelMatch determines whether the container is based on Red Hat technologies through pattern matching logic.
func (r *Release) ReelMatch(pattern string, _ string, _ string) *reel.Step {
	if pattern == NotRedHatBasedRegex {
		r.result = tnf.FAILURE
		r.isRedHatBased = false
	} else if pattern == VersionRegex {
		// If the above conditional is not triggered, it can be deduced that we have matched the VersionRegex.
		r.result = tnf.SUCCESS
		r.isRedHatBased = true
	} else {
		r.result = tnf.ERROR
		r.isRedHatBased = false
	}
	return nil
}

// ReelTimeout does nothing;  no intervention is needed for a timeout.
func (r *Release) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no intervention is needed for EOF.
func (r *Release) ReelEOF() {
}

// NewRelease create a new Release tnf.Test.
func NewRelease(timeout time.Duration) *Release {
	return &Release{result: tnf.ERROR, timeout: timeout, args: strings.Split(ReleaseCommand, " ")}
}
