package push

import (
	"testing"

	"github.com/cylonchau/pantheon/pkg/version"
	"github.com/stretchr/testify/assert"
)

func TestPushOptions_Complete(t *testing.T) {
	tests := []struct {
		name       string
		labels     []string
		wantLabels map[string]string
		wantErr    bool
	}{
		{
			name:   "Valid labels",
			labels: []string{"env=prod", "region=us"},
			wantLabels: map[string]string{
				"env":    "prod",
				"region": "us",
			},
			wantErr: false,
		},
		{
			name:    "Invalid label format",
			labels:  []string{"env"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewPushOptions()
			o.Labels = tt.labels

			// We can't easily mock network/os calls in this simple struct without interface injection,
			// so we mainly verify the parsing logic and that keys exist.
			err := o.Complete(nil)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			for k, v := range tt.wantLabels {
				assert.Equal(t, v, o.ParsedLabels[k])
			}

			// Verify mandatory labels exist
			assert.NotEmpty(t, o.ParsedLabels["hostname"])
			assert.Equal(t, version.Version, o.ParsedLabels["cli_version"])
			// client_ip might be empty if network fails, but key shouldn't crash
		})
	}
}

func TestPushOptions_Validate(t *testing.T) {
	type fields struct {
		JobName string
		Address string
		Metrics []string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "Valid options",
			fields:  fields{JobName: "job", Address: "http://localhost", Metrics: []string{"a=1"}},
			wantErr: false,
		},
		{
			name:    "Missing Address",
			fields:  fields{JobName: "job", Address: "", Metrics: []string{"a=1"}},
			wantErr: true,
		},
		{
			name:    "Missing JobName",
			fields:  fields{JobName: "", Address: "http://localhost", Metrics: []string{"a=1"}},
			wantErr: true,
		},
		{
			name:    "Missing Metrics",
			fields:  fields{JobName: "job", Address: "http://localhost", Metrics: []string{}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &PushOptions{
				JobName: tt.fields.JobName,
				Address: tt.fields.Address,
				Metrics: tt.fields.Metrics,
			}
			if err := o.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("PushOptions.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_parseMetricString(t *testing.T) {
	tests := []struct {
		input     string
		wantName  string
		wantValue float64
		wantType  string
		wantErr   bool
	}{
		{"cpu=0.5", "cpu", 0.5, "gauge", false},
		{"hits=100:counter", "hits", 100, "counter", false},
		{"temp=-10.5:gauge", "temp", -10.5, "gauge", false},
		{"invalid", "", 0, "", true},
		{"cpu=abc", "cpu", 0, "gauge", true},
		{"metric=10:unknown", "metric", 10, "unknown", false}, // Valid parsing, Validate checks type later or Run
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			name, val, typ, err := parseMetricString(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantName, name)
			assert.Equal(t, tt.wantValue, val)
			assert.Equal(t, tt.wantType, typ)
		})
	}
}
