package proxy

import (
	"bytes"
	"testing"
)

func TestFilterMetrics(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		dropPattern string
		expected    string
	}{
		{
			name: "Drop matching metrics and metadata",
			input: `# HELP go_gc_duration_seconds A summary of the pause...
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0.5"} 0.00012
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",status="200"} 1053
`,
			dropPattern: "go_.*",
			expected: `# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",status="200"} 1053
`,
		},
		{
			name: "Empty drop pattern returns original",
			input: `# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",status="200"} 1053
`,
			dropPattern: "",
			expected: `# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",status="200"} 1053
`,
		},
		{
			name: "Keep comments that are not HELP or TYPE",
			input: `# Some generic comment that should be kept
# HELP go_info Information about go runtime
# TYPE go_info gauge
go_info{version="go1.21"} 1
`,
			dropPattern: "go_info",
			expected: `# Some generic comment that should be kept
`,
		},
		{
			name: "Metric without brackets",
			input: `# HELP test_metric A metric without brackets
# TYPE test_metric gauge
test_metric 42
# HELP keep_metric A metric to keep
# TYPE keep_metric gauge
keep_metric 100
`,
			dropPattern: "test_metric",
			expected: `# HELP keep_metric A metric to keep
# TYPE keep_metric gauge
keep_metric 100
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := filterMetrics([]byte(tt.input), tt.dropPattern)
			if !bytes.Equal(output, []byte(tt.expected)) {
				t.Errorf("Test %q failed.\nExpected:\n%s\nGot:\n%s", tt.name, tt.expected, string(output))
			}
		})
	}
}
