package cmd

import (
	"io/ioutil"
	"os"
	"testing"
)

func generateTempConfig(content []string) (fileName string, err error) {

	f, err := ioutil.TempFile(os.TempDir(), "sim-exporter-config-*.yaml")
	if err != nil {
		return "", err
	}
	for _, line := range content {
		f.WriteString(line + "\n")
	}

	return f.Name(), nil
}

func Test_loadAndValidateConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		content []string
		wantErr bool
	}{
		{
			name: "empty",
			content: []string{
				"",
			},
			wantErr: true,
		},
		{
			name: "just-metrics",
			content: []string{
				"metrics:",
			},
			wantErr: true,
		},
		{
			name: "empty-metric",
			content: []string{
				"metrics:",
				"  my_metrics:",
			},
			wantErr: true,
		},
		{
			name: "invalid-metric-type",
			content: []string{
				"metrics:",
				"  my_metric_a:",
				"    type: foo",
			},
			wantErr: true,
		},
		{
			name: "surplus-item-label",
			content: []string{
				"metrics:",
				"  my_metric_b2:",
				"    type: gauge",
				"    labels:",
				"    - l1",
				"    items:",
				"    - value: 1",
				"      labels:",
				"        l1: v1",
				"        l2: v2",
			},
			wantErr: true,
		},
		{
			name: "missing-item-label",
			content: []string{
				"metrics:",
				"  my_metric_b3:",
				"    type: gauge",
				"    labels:",
				"    - l1",
				"    - l2",
				"    items:",
				"    - value: 1",
				"      labels:",
				"        l1: v1",
			},
			wantErr: true,
		},
		{
			name: "valid-metric",
			content: []string{
				"metrics:",
				"  my_metric_b:",
				"    type: gauge",
				"    labels:",
				"    - l1",
				"    items:",
				"    - value: 1",
				"      labels:",
				"        l1: v1",
			},
			wantErr: false,
		},
		{
			name: "valid-metric-nolabel-gauge",
			content: []string{
				"metrics:",
				"  my_metric_c:",
				"    type: gauge",
				"    items:",
				"    - value: 1",
			},
			wantErr: false,
		},
		{
			name: "valid-metric-nolabel-counter",
			content: []string{
				"metrics:",
				"  my_metric_d:",
				"    type: counter",
				"    items:",
				"    - value: 1",
			},
			wantErr: false,
		},
		{
			name: "valid-metric-nolabel-summary",
			content: []string{
				"metrics:",
				"  my_metric_e:",
				"    type: summary",
				"    items:",
				"    - value: 1",
			},
			wantErr: false,
		},
		{
			name: "valid-metric-nolabel-histogram",
			content: []string{
				"metrics:",
				"  my_metric_f:",
				"    type: histogram",
				"    items:",
				"    - value: 1",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempFile, err := generateTempConfig(tt.content)
			if err != nil {
				t.Error(err)
			}
			config, err := loadAndValidateConfiguration(tempFile)
			os.Remove(tempFile)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("loadAndValidateConfiguration(%v) error = %v, wantErr %v", tt.name, err, tt.wantErr)
				}
			} else {
				setupMetricsCollection(config)
				if err != nil {
					t.Error(err)
				}
			}
		})
	}
}
