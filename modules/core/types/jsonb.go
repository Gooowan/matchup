package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type JSONB map[string]any

func (j *JSONB) Scan(value any) error {
	if value == nil {
		*j = make(JSONB)
		return nil
	}

	switch v := value.(type) {
	case []byte:
		// Try to unmarshal directly first
		if err := json.Unmarshal(v, j); err != nil {
			// If it fails, try to handle double-encoded JSON
			var str string
			if json.Unmarshal(v, &str) == nil {
				return json.Unmarshal([]byte(str), j)
			}
			return err
		}
		return nil
	case string:
		// Try to unmarshal directly first
		if err := json.Unmarshal([]byte(v), j); err != nil {
			// If it fails, try to handle double-encoded JSON
			var str string
			if json.Unmarshal([]byte(v), &str) == nil {
				return json.Unmarshal([]byte(str), j)
			}
			return err
		}
		return nil
	default:
		return fmt.Errorf("cannot scan %T into JSONB", value)
	}
}

func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
} 