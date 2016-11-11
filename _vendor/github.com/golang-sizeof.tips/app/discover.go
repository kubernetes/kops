package app

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gophergala/golang-sizeof.tips/internal/parser"
)

const exampleCode = `// Sample code
struct {
	a string
	b bool
	c string
}
`

func discoverHandler(w http.ResponseWriter, r *http.Request) {
	code := parseCodeRequestParam(r.FormValue("t"))
	if code == "" {
		code = exampleCode
	}

	toRender := &struct {
		Code   string
		Result *viewData
		Error  string
	}{Code: code}

	result, err := parser.ParseCode(code)
	if err != nil {
		toRender.Error = err.Error()
	} else {
		toRender.Result = createViewData(result)
	}

	templates["index"].ExecuteTemplate(w, "base", toRender)
}

func parseCodeRequestParam(param string) string {
	param = strings.TrimSpace(param)
	bytes, err := base64.URLEncoding.DecodeString(param)
	if err != nil {
		return ""
	}
	return string(bytes)
}

type chunk struct {
	Cells     []bool
	IsPadding bool
}

func newChunk(offset, size, align uint64) *chunk {
	ch := &chunk{Cells: make([]bool, align)}
	for i := uint64(0); i < size; i++ {
		ch.Cells[offset+i] = true
	}
	return ch
}

func (ch *chunk) asPadding() *chunk {
	ch.IsPadding = true
	return ch
}

type row struct {
	Chunks []*chunk
	Name   string
}

type viewData struct {
	*parser.TypeInfo
	Details []*row
}

func (data *viewData) prepareFields(
	fields []*parser.TypeInfo,
	offset uint64,
	topLevel bool,
) uint64 {
	for _, field := range fields {
		switch {
		case field.IsArray:
			fallthrough
		case field.IsFixed:
			if field.Sizeof == 0 {
				data.Details = append(data.Details, &row{
					Name:   field.Name,
					Chunks: []*chunk{newChunk(0, 0, data.Alignof)},
				})
				continue
			}
			size := uint64(0)
			n := field.Sizeof / field.Alignof
			chunks := make([]*chunk, 0, n)
			for i := uint64(0); i < n; i++ {
				size += field.Alignof
				if offset+size <= data.Alignof {
					continue
				}
				chunkSize := size - field.Alignof
				if offset+chunkSize < data.Alignof {
					paddingSize := data.Alignof - (offset + chunkSize)
					data.Details = append(data.Details, &row{
						Name: "padding",
						Chunks: []*chunk{newChunk(
							offset, paddingSize, data.Alignof,
						).asPadding()},
					})
					offset = (offset + paddingSize) % data.Alignof
				}
				if chunkSize > 0 {
					chunks = append(chunks, newChunk(
						offset, chunkSize, data.Alignof,
					))
					offset = (offset + chunkSize) % data.Alignof
				}
				offset = 0
				size = field.Alignof % data.Alignof
				if field.Alignof == data.Alignof {
					chunks = append(chunks, newChunk(
						offset, field.Alignof, data.Alignof,
					))
				}
			}
			if size > 0 {
				chunks = append(chunks, newChunk(
					offset, size, data.Alignof,
				))
				offset = (offset + size) % data.Alignof
			}
			if len(chunks) > 0 {
				data.Details = append(data.Details, &row{
					Name:   field.Name,
					Chunks: chunks,
				})
			}
		case field.IsStruct:
			if len(field.Fields) < 1 {
				data.Details = append(data.Details, &row{
					Name:   field.Name,
					Chunks: []*chunk{newChunk(0, 0, data.Alignof)},
				})
				continue
			}
			offset = data.prepareFields(field.Fields, offset, false)
		}
	}
	if topLevel && offset > 0 && offset < data.Alignof {
		data.Details = append(data.Details, &row{
			Name: "padding",
			Chunks: []*chunk{newChunk(
				offset, data.Alignof-offset, data.Alignof,
			).asPadding()},
		})
	}
	return offset
}

func createViewData(typ *parser.TypeInfo) (data *viewData) {
	data = &viewData{TypeInfo: typ}
	if !typ.IsStruct {
		return
	}
	data.Details = make([]*row, 0, len(typ.Fields))
	data.prepareFields(typ.Fields, 0, true)
	return
}
