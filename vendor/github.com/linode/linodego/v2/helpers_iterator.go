package linodego

import (
	"iter"
	"slices"
)

// mapIter returns a new iterator of the values in the given iterator transformed using the given transform function.
func mapIter[I, O any](values iter.Seq[I], transform func(I) O) iter.Seq[O] {
	return func(yield func(O) bool) {
		for value := range values {
			if !yield(transform(value)) {
				return
			}
		}
	}
}

// mapSlice returns a new slice of the values in the given slice transformed using the given transform function.
func mapSlice[I, O any](values []I, transform func(I) O) []O {
	return slices.Collect(mapIter(slices.Values(values), transform))
}
