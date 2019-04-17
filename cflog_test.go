package cflog

import (
	"fmt"
	"testing"

	loggingpb "google.golang.org/genproto/googleapis/logging/v2"
)

func TestSetEntrypayload(t *testing.T) {
	textType := "*logging.LogEntry_TextPayload"
	jsonType := "*logging.LogEntry_JsonPayload"

	tests := []struct {
		name         string
		input        interface{}
		expectedType string
	}{
		{name: "string", input: "str", expectedType: textType},
		{name: "empty string", input: "", expectedType: textType},
		{name: "[]byte", input: []byte("bytes"), expectedType: textType},
		{name: "nil", input: nil, expectedType: textType},
		{name: "struct", input: struct{ M string }{M: "in"}, expectedType: jsonType},
		{name: "JSON string", input: `{"m": "m"}`, expectedType: jsonType},
		{name: "bad JSON string", input: `{m": "m"}`, expectedType: textType},
		{name: "empty struct", input: struct{}{}, expectedType: jsonType},
		{name: "uninitialized struct", input: struct{ M string }{}, expectedType: jsonType},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := &loggingpb.LogEntry{}
			if err := setEntryPayload(entry, test.input); err != nil {
				t.Fatal("Set error", err)
			}

			pType := fmt.Sprintf("%T", entry.Payload)
			if pType != test.expectedType {
				t.Fatal("Unexpected type", pType)
			}
		})
	}

}
