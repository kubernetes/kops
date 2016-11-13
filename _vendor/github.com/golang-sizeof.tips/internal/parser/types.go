package parser

import (
	"errors"
	"fmt"
	. "go/ast"
	. "go/parser"
	"go/token"
	"reflect"
	"strconv"
	"unsafe"
)

var BasicSizes = map[string]uint64{
	"bool":       uint64(unsafe.Sizeof(true)),
	"int8":       uint64(unsafe.Sizeof(int8(0))),
	"uint8":      uint64(unsafe.Sizeof(uint8(0))),
	"byte":       uint64(unsafe.Sizeof(byte(0))),
	"int16":      uint64(unsafe.Sizeof(int16(0))),
	"uint16":     uint64(unsafe.Sizeof(uint16(0))),
	"int32":      uint64(unsafe.Sizeof(int32(0))),
	"uint32":     uint64(unsafe.Sizeof(uint32(0))),
	"rune":       uint64(unsafe.Sizeof(rune(0))),
	"float32":    uint64(unsafe.Sizeof(float32(0))),
	"int":        uint64(unsafe.Sizeof(int(0))),
	"uint":       uint64(unsafe.Sizeof(uint(0))),
	"int64":      uint64(unsafe.Sizeof(int64(0))),
	"uint64":     uint64(unsafe.Sizeof(uint64(0))),
	"float64":    uint64(unsafe.Sizeof(float64(0))),
	"uintptr":    uint64(unsafe.Sizeof(uintptr(0))),
	"complex64":  uint64(unsafe.Sizeof(complex64(0))),
	"complex128": uint64(unsafe.Sizeof(complex128(0))),
	"string":     uint64(unsafe.Sizeof("")),
}

var FixedSizes = map[string]uint64{
	"ptr":   uint64(unsafe.Sizeof(&struct{}{})),
	"map":   uint64(unsafe.Sizeof(map[bool]bool{})),
	"slice": uint64(unsafe.Sizeof([]struct{}{})),
	"chan":  uint64(unsafe.Sizeof(make(chan struct{}))),
	"func":  uint64(unsafe.Sizeof(func() {})),
}

var (
	errInvalidArrayLength = errors.New("invalid length in array definition")
	errInvalidType        = errors.New("invalid type expression")
)

type TypeInfo struct {
	Sizeof   uint64
	Alignof  uint64
	Name     string
	IsFixed  bool
	IsArray  bool
	IsStruct bool
	Fields   []*TypeInfo
}

func parseType(n Node) (*TypeInfo, error) {
	switch node := n.(type) {
	case *Ident:
		size, exists := BasicSizes[node.Name]
		if !exists {
			return nil, fmt.Errorf("unknown type '%s'", node.Name)
		}
		return &TypeInfo{
			Sizeof:  size,
			Alignof: min(size, BasicSizes["uintptr"]),
			Name:    node.Name,
			IsFixed: true,
		}, nil
	case *StarExpr: // todo: maybe more deep checking?
		return &TypeInfo{
			Sizeof:  FixedSizes["ptr"],
			Alignof: min(FixedSizes["ptr"], BasicSizes["uintptr"]),
			Name:    "pointer",
			IsFixed: true,
		}, nil
	case *MapType:
		return &TypeInfo{
			Sizeof:  FixedSizes["map"],
			Alignof: min(FixedSizes["map"], BasicSizes["uintptr"]),
			Name:    "map",
			IsFixed: true,
		}, nil
	case *ChanType:
		return &TypeInfo{
			Sizeof:  FixedSizes["chan"],
			Alignof: min(FixedSizes["chan"], BasicSizes["uintptr"]),
			Name:    "channel",
			IsFixed: true,
		}, nil
	case *FuncLit:
		return &TypeInfo{
			Sizeof:  FixedSizes["func"],
			Alignof: min(FixedSizes["func"], BasicSizes["uintptr"]),
			Name:    "function",
			IsFixed: true,
		}, nil
	case *FuncType:
		return &TypeInfo{
			Sizeof:  FixedSizes["func"],
			Alignof: min(FixedSizes["func"], BasicSizes["uintptr"]),
			Name:    "function",
			IsFixed: true,
		}, nil
	case *ArrayType:
		if node.Len == nil {
			return &TypeInfo{
				Sizeof:  FixedSizes["slice"],
				Alignof: min(FixedSizes["slice"], BasicSizes["uintptr"]),
				Name:    "slice",
				IsFixed: true,
			}, nil
		}
		len, ok := node.Len.(*BasicLit)
		if !ok || len.Kind != token.INT {
			return nil, errInvalidArrayLength
		}
		num, err := strconv.ParseUint(len.Value, 10, 64)
		if err != nil {
			return nil, errInvalidArrayLength
		}
		typ, err := parseType(node.Elt)
		if err != nil {
			return nil, err
		}
		return &TypeInfo{
			Sizeof:  num * typ.Sizeof,
			Alignof: typ.Alignof,
			Name:    "array",
			IsArray: true,
		}, nil
	case *StructType:
		strct := &TypeInfo{
			Alignof:  1, // empty struct has unsafe.Alignof() == 1
			Name:     "struct",
			IsStruct: true,
		}
		if len(node.Fields.List) < 1 {
			return strct, nil
		}
		strct.Fields = make([]*TypeInfo, len(node.Fields.List))
		for i, field := range node.Fields.List {
			typ, err := parseType(field.Type)
			if err != nil {
				return nil, err
			}
			if len(field.Names) > 0 {
				typ.Name = field.Names[0].Name + " " + typ.Name
			}
			if typ.Alignof > strct.Alignof {
				strct.Alignof = typ.Alignof
			}
			strct.Fields[i] = typ
		}
		num, size := uint64(0), uint64(0)
		for _, typ := range strct.Fields {
			if typ.Sizeof == 0 {
				continue
			}
			n := typ.Sizeof / typ.Alignof
			for i := uint64(0); i < n; i++ {
				size += typ.Alignof
				if size <= strct.Alignof {
					continue
				}
				size = typ.Alignof % strct.Alignof
				num++
				if typ.Alignof == strct.Alignof {
					num++
				}
			}
		}
		if size > 0 {
			num++
		}
		strct.Sizeof = num * strct.Alignof
		return strct, nil
	default:
		//return nil, errInvalidType
		return nil, fmt.Errorf("%v", reflect.TypeOf(n))
	}
}

func min(x, y uint64) uint64 {
	if x < y {
		return x
	}
	return y
}

func ParseCode(code string) (*TypeInfo, error) {
	expr, err := ParseExpr(code)
	if err != nil {
		return nil, fmt.Errorf("syntax error: %s", err.Error())
	}
	typ, err := parseType(expr)
	if err != nil {
		return nil, fmt.Errorf("type error: %s", err.Error())
	}
	return typ, nil
}
