package lipgloss

import (
	"image/color"
	"sync"

	"github.com/muesli/termenv"
)

var (
	colorProfile         termenv.Profile
	getColorProfile      sync.Once
	explicitColorProfile bool

	hasDarkBackground       bool
	getBackgroundColor      sync.Once
	explicitBackgroundColor bool

	colorProfileMtx sync.RWMutex
)

// ColorProfile returns the detected termenv color profile. It will perform the
// actual check only once.
func ColorProfile() termenv.Profile {
	colorProfileMtx.RLock()
	defer colorProfileMtx.RUnlock()

	if !explicitColorProfile {
		getColorProfile.Do(func() {
			colorProfile = termenv.EnvColorProfile()
		})
	}
	return colorProfile
}

// SetColorProfile sets the color profile on a package-wide context. This
// function exists mostly for testing purposes so that you can assure you're
// testing against a specific profile.
//
// Outside of testing you likely won't want to use this function as
// ColorProfile() will detect and cache the terminal's color capabilities
// and choose the best available profile.
//
// Available color profiles are:
//
// termenv.Ascii (no color, 1-bit)
// termenv.ANSI (16 colors, 4-bit)
// termenv.ANSI256 (256 colors, 8-bit)
// termenv.TrueColor (16,777,216 colors, 24-bit)
//
// This function is thread-safe.
func SetColorProfile(p termenv.Profile) {
	colorProfileMtx.Lock()
	defer colorProfileMtx.Unlock()

	colorProfile = p
	explicitColorProfile = true
}

// HasDarkBackground returns whether or not the terminal has a dark background.
func HasDarkBackground() bool {
	colorProfileMtx.RLock()
	defer colorProfileMtx.RUnlock()

	if !explicitBackgroundColor {
		getBackgroundColor.Do(func() {
			hasDarkBackground = termenv.HasDarkBackground()
		})
	}

	return hasDarkBackground
}

// SetHasDarkBackground sets the value of the background color detection on a
// package-wide context. This function exists mostly for testing purposes so
// that you can assure you're testing against a specific background color
// setting.
//
// Outside of testing you likely won't want to use this function as
// HasDarkBackground() will detect and cache the terminal's current background
// color setting.
//
// This function is thread-safe.
func SetHasDarkBackground(b bool) {
	colorProfileMtx.Lock()
	defer colorProfileMtx.Unlock()

	hasDarkBackground = b
	explicitBackgroundColor = true
}

// TerminalColor is a color intended to be rendered in the terminal. It
// satisfies the Go color.Color interface.
type TerminalColor interface {
	value() string
	color() termenv.Color
	RGBA() (r, g, b, a uint32)
}

// NoColor is used to specify the absence of color styling. When this is active
// foreground colors will be rendered with the terminal's default text color,
// and background colors will not be drawn at all.
//
// Example usage:
//
//	var style = someStyle.Copy().Background(lipgloss.NoColor{})
type NoColor struct{}

func (n NoColor) value() string {
	return ""
}

func (n NoColor) color() termenv.Color {
	return ColorProfile().Color("")
}

// RGBA returns the RGBA value of this color. Because we have to return
// something, despite this color being the absence of color, we're returning
// black with 100% opacity.
//
// Red: 0x0, Green: 0x0, Blue: 0x0, Alpha: 0xFFFF.
func (n NoColor) RGBA() (r, g, b, a uint32) {
	return 0x0, 0x0, 0x0, 0xFFFF
}

var noColor = NoColor{}

// Color specifies a color by hex or ANSI value. For example:
//
//	ansiColor := lipgloss.Color("21")
//	hexColor := lipgloss.Color("#0000ff")
type Color string

func (c Color) value() string {
	return string(c)
}

func (c Color) color() termenv.Color {
	return ColorProfile().Color(string(c))
}

// RGBA returns the RGBA value of this color. This satisfies the Go Color
// interface. Note that on error we return black with 100% opacity, or:
//
// Red: 0x0, Green: 0x0, Blue: 0x0, Alpha: 0xFF.
func (c Color) RGBA() (r, g, b, a uint32) {
	return hexToColor(c.value()).RGBA()
}

// AdaptiveColor provides color options for light and dark backgrounds. The
// appropriate color will be returned at runtime based on the darkness of the
// terminal background color.
//
// Example usage:
//
//	color := lipgloss.AdaptiveColor{Light: "#0000ff", Dark: "#000099"}
type AdaptiveColor struct {
	Light string
	Dark  string
}

func (ac AdaptiveColor) value() string {
	if HasDarkBackground() {
		return ac.Dark
	}
	return ac.Light
}

func (ac AdaptiveColor) color() termenv.Color {
	return ColorProfile().Color(ac.value())
}

// RGBA returns the RGBA value of this color. This satisfies the Go Color
// interface. Note that on error we return black with 100% opacity, or:
//
// Red: 0x0, Green: 0x0, Blue: 0x0, Alpha: 0xFF.
func (ac AdaptiveColor) RGBA() (r, g, b, a uint32) {
	cf := hexToColor(ac.value())
	return cf.RGBA()
}

// CompleteColor specifies exact values for truecolor, ANSI256, and ANSI color
// profiles. Automatic color degredation will not be performed.
type CompleteColor struct {
	TrueColor string
	ANSI256   string
	ANSI      string
}

func (c CompleteColor) value() string {
	switch ColorProfile() {
	case termenv.TrueColor:
		return c.TrueColor
	case termenv.ANSI256:
		return c.ANSI256
	case termenv.ANSI:
		return c.ANSI
	default:
		return ""
	}
}

func (c CompleteColor) color() termenv.Color {
	return colorProfile.Color(c.value())
}

// RGBA returns the RGBA value of this color. This satisfies the Go Color
// interface. Note that on error we return black with 100% opacity, or:
//
// Red: 0x0, Green: 0x0, Blue: 0x0, Alpha: 0xFFFF
func (c CompleteColor) RGBA() (r, g, b, a uint32) {
	return hexToColor(c.value()).RGBA()
}

// CompleteColor specifies exact values for truecolor, ANSI256, and ANSI color
// profiles, with separate options for light and dark backgrounds. Automatic
// color degredation will not be performed.
type CompleteAdaptiveColor struct {
	Light CompleteColor
	Dark  CompleteColor
}

func (cac CompleteAdaptiveColor) value() string {
	if HasDarkBackground() {
		return cac.Dark.value()
	}
	return cac.Light.value()
}

func (cac CompleteAdaptiveColor) color() termenv.Color {
	return ColorProfile().Color(cac.value())
}

// RGBA returns the RGBA value of this color. This satisfies the Go Color
// interface. Note that on error we return black with 100% opacity, or:
//
// Red: 0x0, Green: 0x0, Blue: 0x0, Alpha: 0xFFFF
func (cac CompleteAdaptiveColor) RGBA() (r, g, b, a uint32) {
	return hexToColor(cac.value()).RGBA()
}

// hexToColor translates a hex color string (#RRGGBB or #RGB) into a color.RGB,
// which satisfies the color.Color interface. If an invalid string is passed
// black with 100% opacity will be returned: or, in hex format, 0x000000FF.
func hexToColor(hex string) (c color.RGBA) {
	c.A = 0xFF

	if hex == "" || hex[0] != '#' {
		return c
	}

	const (
		fullFormat  = 7 // #RRGGBB
		shortFormat = 4 // #RGB
	)

	switch len(hex) {
	case fullFormat:
		const offset = 4
		c.R = hexToByte(hex[1])<<offset + hexToByte(hex[2])
		c.G = hexToByte(hex[3])<<offset + hexToByte(hex[4])
		c.B = hexToByte(hex[5])<<offset + hexToByte(hex[6])
	case shortFormat:
		const offset = 0x11
		c.R = hexToByte(hex[1]) * offset
		c.G = hexToByte(hex[2]) * offset
		c.B = hexToByte(hex[3]) * offset
	}

	return c
}

func hexToByte(b byte) byte {
	const offset = 10
	switch {
	case b >= '0' && b <= '9':
		return b - '0'
	case b >= 'a' && b <= 'f':
		return b - 'a' + offset
	case b >= 'A' && b <= 'F':
		return b - 'A' + offset
	}
	// Invalid, but just return 0.
	return 0
}
