package npc

import (
	"encoding/json"
)

func js(v interface{}) string {
	a, _ := json.Marshal(v)
	return string(a)
}
