package parser

import (
	"testing"
	"unsafe"
)

func TestTypeParsing(t *testing.T) {
	cases := map[string]uint64{
		`struct{}`:  uint64(unsafe.Sizeof(struct{}{})),
		`[0]string`: uint64(unsafe.Sizeof([0]string{})),
		`struct{a struct{}; b bool}`: uint64(unsafe.Sizeof(struct {
			a struct{}
			b bool
		}{})),
		`struct{b bool; u int32}`: uint64(unsafe.Sizeof(struct {
			b bool
			u int32
		}{})),
		`struct{a bool; b bool; u int32}`: uint64(unsafe.Sizeof(struct {
			a bool
			b bool
			u int32
		}{})),
		`struct{u int32; a bool; b bool}`: uint64(unsafe.Sizeof(struct {
			u int32
			a bool
			b bool
		}{})),
		`struct{a bool; u int32; b bool}`: uint64(unsafe.Sizeof(struct {
			a bool
			u int32
			b bool
		}{})),
		`struct{a bool; s string; b bool}`: uint64(unsafe.Sizeof(struct {
			a bool
			s string
			b bool
		}{})),
		`struct{a [3]bool; b int16; c [6]bool}`: uint64(unsafe.Sizeof(struct {
			a [3]bool
			b int16
			c [6]bool
		}{})),
		`struct{a [3]bool; b struct{a int16; c struct{a string}}; c [6]bool}`: uint64(unsafe.Sizeof(struct {
			a [3]bool
			b struct {
				a int16
				c struct{ a string }
			}
			c [6]bool
		}{})),
		`struct{a uint64; b complex128; c bool; d [6]bool; n struct{a bool}}`: uint64(unsafe.Sizeof(struct {
			a uint64
			b complex128
			c bool
			d [6]bool
			n struct{ a bool }
		}{})),
		`struct{a uint64; b complex128; c bool; d [6]bool; n struct{a [2]bool}}`: uint64(unsafe.Sizeof(struct {
			a uint64
			b complex128
			c bool
			d [6]bool
			n struct{ a [2]bool }
		}{})),
		`struct{a uint64; b complex128; c bool; d [6]bool; n struct{a [9]bool}}`: uint64(unsafe.Sizeof(struct {
			a uint64
			b complex128
			c bool
			d [6]bool
			n struct{ a [9]bool }
		}{})),
		`struct{a uint64; f [2]uint32; ff uint16; dd struct{a [3]int32}}`: uint64(unsafe.Sizeof(struct {
			a  uint64
			f  [2]uint32
			ff uint16
			dd struct{ a [3]int32 }
		}{})),
	}
	for code, size := range cases {
		typ, err := ParseCode(code)
		if err != nil {
			t.Fatalf(
				"failed to parse code '%s', reason -> %s",
				code, err.Error(),
			)
		}
		if typ.Sizeof != size {
			t.Errorf(
				"invalid sizeof('%s')\n\texpected: %d\n\tactual: %d",
				code, size, typ.Sizeof,
			)
		}
	}
}
