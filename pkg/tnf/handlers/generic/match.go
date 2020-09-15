package generic

// Match follows the Container design pattern, and is used to store the arguments to a reel.Handler's ReelMatch
// function in a single data transfer object.
type Match struct {

	// Pattern is the pattern causing a match in reel.Handler ReelMatch.
	Pattern string `json:"pattern,omitempty" yaml:"pattern,omitempty"`

	// Before contains the text before the Match.
	Before string `json:"before,omitempty" yaml:"before,omitempty"`

	// Match is the matched string.
	Match string `json:"match,omitempty" yaml:"match,omitempty"`
}
