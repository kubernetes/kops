package codegen

import "io"

type Generator interface {
	Init(parser *GoParser) error
	WriteHeader(w io.Writer) error
	WriteType(w io.Writer, typeName string) error
}
