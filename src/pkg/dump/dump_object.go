// Package dump provide a debug dump of objects.
package dump

import (
	"encoding/json"
	"fmt"
)

// Object produces a debug string for an object.
func Object(val any) string {
	b, err := json.MarshalIndent(val, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"err": "%s"}`, err.Error())
	}

	return string(b)
}
