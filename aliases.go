package spec

// Aliases returns before, after, and context bound to it.Before, it.After, and describe.AsContext() -- the one-line declaration this account's specs start every file with.
func Aliases(describe G, it S) (before, after func(func()), context G) {
	return it.Before, it.After, describe.AsContext()
}
