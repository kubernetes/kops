package protokube

import (
	"encoding/json"
	"fmt"
)

func DebugString(o interface{}) string {
	b, err := json.Marshal(o)
	if err != nil {
		return fmt.Sprintf("error marshaling %T: %v", o, err)
	}
	return string(b)
}
