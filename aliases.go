package spec

import "testing"

// Aliases returns before, after, and context bound to it.Before, it.After, and describe.AsContext().
func Aliases(describe G, it S) (before, after func(func()), context G) {
	return it.Before, it.After, describe.AsContext()
}

// RunAliased wraps Run, passing before, after, and context (see Aliases) directly as parameters to f -- no per-file alias declaration needed anywhere.
func RunAliased(t *testing.T, text string, f func(t *testing.T, describe, context G, it S, before, after func(func())), opts ...Option) bool {
	t.Helper()
	return Run(t, text, func(t *testing.T, describe G, it S) {
		before, after, context := Aliases(describe, it)
		f(t, describe, context, it, before, after)
	}, opts...)
}
