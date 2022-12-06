package diode

// Reporter is used to report alerts and warnings.
type Reporter interface {
	Alert(dropped uint64) // Some values were overwritten.
	Warn(msg string)      // A warning message was generated.
}

// Assert reporter implements Reporter.
var _ Reporter = reporter{}

// reporter is a struct that satisfies the Reporter interface, but
// doesn't actually do anything. Just in case no Reporter is set.
type reporter struct{}

func (r reporter) Alert(dropped uint64) {}
func (r reporter) Warn(msg string)      {}
