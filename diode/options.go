package diode

type options struct {
	rep  Reporter
	safe bool
	copy bool
}

// Option configures how we set up the diode.
type Option interface {
	apply(*options)
}

// optionFunc wraps a function that modifies options into an implementation of
// the Option interface.
type optionFunc struct {
	f func(*options)
}

func (of *optionFunc) apply(o *options) {
	of.f(o)
}

// WithReporter returns an Option which sets a Reporter for the diode to use for
// alerts and warnings.
func WithReporter(r Reporter) Option {
	return &optionFunc{
		f: func(o *options) {
			o.rep = r
		},
	}
}

// WithManyWriters returns an Option which tells the diode to use a many-to-one
// configuration that is safe for many writers (on go-routines B-n), and a
// single reader (on go-routine A). Note that this option affects performance.
func WithManyWriters() Option {
	return &optionFunc{
		f: func(o *options) {
			o.safe = true
		},
	}
}

// WithCopy returns an Option which tells the diode to make copies of data
// rather than reusing pointers passed into Set. Note that this option affects
// performance.
func WithCopy() Option {
	return &optionFunc{
		f: func(o *options) {
			o.copy = true
		},
	}
}
