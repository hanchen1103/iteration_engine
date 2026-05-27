package testkit

import "encoding/json"

func RawJSON(value string) json.RawMessage {
	return json.RawMessage(value)
}
