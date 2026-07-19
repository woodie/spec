package spec

// Var holds a value shared between a Before hook and the it blocks that read it.
// Declaring one fresh inside a G body gets automatic per-spec freshness for
// free from spec's own re-evaluation model; what Var adds on top is a clear
// panic on Get before Set, instead of a silent zero value.
type Var[T any] struct {
	name string
	val  T
	set  bool
}

// NewVar declares a Var; name appears in the panic message if Get runs before Set.
func NewVar[T any](name string) *Var[T] {
	return &Var[T]{name: name}
}

// Set stores a value, usually called from Before.
func (v *Var[T]) Set(val T) {
	v.val = val
	v.set = true
}

// Get returns the stored value, panicking if called before Set in this spec.
func (v *Var[T]) Get() T {
	if !v.set {
		panic("spec: " + v.name + " read before Set")
	}
	return v.val
}
