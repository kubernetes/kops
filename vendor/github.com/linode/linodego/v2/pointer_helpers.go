package linodego

/*
Pointer takes a value of any type T and returns a pointer to that value.
Go does not allow directly creating pointers to literals, so Pointer enables
abstraction away the pointer logic.

Example:

		booted := true

		createOpts := linodego.InstanceCreateOptions{
			Booted: &booted,
		}

		can be replaced with

		createOpts := linodego.InstanceCreateOptions{
			Booted: linodego.Pointer(true),
		}
*/

func Pointer[T any](value T) *T {
	return &value
}

// DoublePointer creates a double pointer to a value of type T.
//
// This is useful for APIs that distinguish between null and omitted fields.
//
// Example:
//
//	// For a field that should be non-null value in the API payload:
//	value := linodego.DoublePointer(42) // Returns **int pointing a *int pointer pointing to 42
//
//	// For a field that should be null in the API payload, use `DoublePointerNull` function instead:
//	nullValue := linodego.DoublePointerNull[int]() // Returns **int that is nil
//
//	// For a field that should not be included in the API payload, simply not include it in the struct.
func DoublePointer[T any](value T) **T {
	valuePtr := &value
	return &valuePtr
}

// DoublePointerNull creates a double pointer pointing to a nil pointer of type T,
// indicating that the field should be null in the API payload.
//
// This is useful for APIs that distinguish between null and omitted fields.
func DoublePointerNull[T any]() **T {
	return Pointer[*T](nil)
}
