// A package to help with logging in GCP Cloud Functions
package cflog

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"cloud.google.com/go/logging/apiv2"
	_struct "github.com/golang/protobuf/ptypes/struct"
	"github.com/micro/protobuf/jsonpb"
	"google.golang.org/genproto/googleapis/api/monitoredres"
	"google.golang.org/genproto/googleapis/logging/type"
	loggingpb "google.golang.org/genproto/googleapis/logging/v2"
)

// Severity is a wrapper around an int for log severity
type Severity ltype.LogSeverity

// Using the log packages severity
// https://godoc.org/google.golang.org/genproto/googleapis/logging/type#LogSeverity
const (
	SeverityDefault   = Severity(ltype.LogSeverity_DEFAULT)
	SeverityDebug     = Severity(ltype.LogSeverity_DEBUG)
	SeverityInfo      = Severity(ltype.LogSeverity_INFO)
	SeverityNotice    = Severity(ltype.LogSeverity_NOTICE)
	SeverityWarning   = Severity(ltype.LogSeverity_WARNING)
	SeverityError     = Severity(ltype.LogSeverity_ERROR)
	SeverityCritical  = Severity(ltype.LogSeverity_CRITICAL)
	SeverityAlert     = Severity(ltype.LogSeverity_ALERT)
	SeverityEmergency = Severity(ltype.LogSeverity_EMERGENCY)
)

// Client holds a logging client and the resources needed for
type Client struct {
	client               *logging.Client
	logName              string
	logMonitoredResource *monitoredres.MonitoredResource
}

// NewClient creates a client for writing logs using environment variable
// https://cloud.google.com/functions/docs/env-var
func NewClient(ctx context.Context) (Client, error) {
	c := Client{}
	client, err := logging.NewClient(ctx)
	if err != nil {
		return c, err
	}

	c.client = client
	c.logName = fmt.Sprintf("projects/%s/logs/cloudfunctions.googleapis.com%scloud-functions", os.Getenv("GCP_PROJECT"), "%2F")
	c.logMonitoredResource = &monitoredres.MonitoredResource{
		Type: "cloud_function",
		Labels: map[string]string{
			"function_name": os.Getenv("FUNCTION_NAME"),
			"project_id":    os.Getenv("GCP_PROJECT"),
			"region":        os.Getenv("FUNCTION_REGION"),
		},
	}

	return c, nil
}

// Close will close the underlying client
func (c Client) Close() error {
	return c.client.Close()
}

func setEntryPayload(entry *loggingpb.LogEntry, in interface{}) error {
	var s string
	switch v := in.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		if v == nil {
			break
		}
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		s = string(data)
	}

	if strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}") {
		var payload _struct.Struct
		if err := jsonpb.UnmarshalString(s, &payload); err == nil {
			entry.Payload = &loggingpb.LogEntry_JsonPayload{JsonPayload: &payload}
			return nil
		}
	}

	entry.Payload = &loggingpb.LogEntry_TextPayload{TextPayload: s}
	return nil
}

// Log creates a log using the payload given
// Payload should be either a string or a struct that can marshal to JSON
func (c Client) Log(ctx context.Context, severity Severity, payload interface{}) error {
	// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry
	entry := &loggingpb.LogEntry{
		LogName:  c.logName,
		Resource: c.logMonitoredResource,
		Severity: ltype.LogSeverity(severity),
	}
	if err := setEntryPayload(entry, payload); err != nil {
		return err
	}

	req := &loggingpb.WriteLogEntriesRequest{Entries: []*loggingpb.LogEntry{entry}}
	if _, err := c.client.WriteLogEntries(ctx, req); err != nil {
		return err
	}
	return nil
}

var singleton Client

// Log uses an auto generated singleton client
func Log(ctx context.Context, severity Severity, payload interface{}) {
	if singleton.client == nil {
		var err error
		singleton, err = NewClient(context.Background())
		if err != nil {
			log.Printf("Could not create client to log payload '%q': %v", payload, err)
			return
		}
	}

	if err := singleton.Log(ctx, severity, payload); err != nil {
		log.Printf("Could not log payload '%q': %v", payload, err)
	}
}

// Debug calls Log with the severity set to Debug
func Debug(ctx context.Context, payload interface{}) { Log(ctx, SeverityDebug, payload) }

// Info calls Log with the severity set to Info
func Info(ctx context.Context, payload interface{}) { Log(ctx, SeverityInfo, payload) }

// Warn calls Log with the severity set to Warning
func Warn(ctx context.Context, payload interface{}) { Log(ctx, SeverityWarning, payload) }

// Error calls Log with the severity set to Error
func Error(ctx context.Context, payload interface{}) { Log(ctx, SeverityError, payload) }

// Critical calls Log with the severity set to Critical
func Critical(ctx context.Context, payload interface{}) { Log(ctx, SeverityCritical, payload) }
