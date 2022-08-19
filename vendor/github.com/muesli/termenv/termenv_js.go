//go:build js
// +build js

package termenv

func colorProfile() Profile {
	return ANSI256
}

func foregroundColor() Color {
	// default gray
	return ANSIColor(7)
}

func backgroundColor() Color {
	// default black
	return ANSIColor(0)
}
