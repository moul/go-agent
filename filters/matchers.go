package filters

// Matcher is the interface shared by all matchers.
type Matcher interface {
	Matches(x interface{}) bool
}
