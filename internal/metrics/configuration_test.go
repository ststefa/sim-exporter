package metrics

import (
	"database/sql"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"gopkg.in/guregu/null.v4"
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

func Test_LoadAndValidateConfiguration(t *testing.T) {
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
			name: "valid-multi-value-metric",
			content: []string{
				"metrics:",
				"  my_metric_b:",
				"    type: gauge",
				"    labels:",
				"    - l1",
				"    items:",
				"    - value: 100-200",
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
			config, err := LoadAndValidateConfiguration(tempFile)
			os.Remove(tempFile)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("LoadAndValidateConfiguration(%v) error = %v, wantErr %v", tt.name, err, tt.wantErr)
				}
			} else {
				if tt.wantErr {
					t.Errorf("LoadAndValidateConfiguration(%v) no error but wantErr %v", tt.name, tt.wantErr)
				}
				err = SetupMetricsCollection(config)
				if err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestConfigurationMetricItem_parseFloatFromString(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name    string
		args    args
		want    null.Float
		wantErr bool
	}{
		{
			name: "123456789",
			args: args{
				value: "123456789",
			},
			want:    null.Float{sql.NullFloat64{123456789, true}},
			wantErr: false,
		},
		{
			name: "empty",
			args: args{
				value: "",
			},
			want:    null.Float{sql.NullFloat64{0, false}},
			wantErr: true,
		},
		{
			name: "not-a-number",
			args: args{
				value: "not-a-number",
			},
			want:    null.Float{sql.NullFloat64{0, false}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFloatFromString(tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConfigurationMetricItem.parseFloatFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ConfigurationMetricItem.parseFloatFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigurationMetricItem_generateValue(t *testing.T) {
	type fields struct {
		Value     string
		value     null.Float
		rangeFrom null.Float
		rangeTo   null.Float
		Labels    map[string]string
	}
	tests := []struct {
		name   string
		fields fields
		want   null.Float
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ConfigurationMetricItem{
				Value:     tt.fields.Value,
				value:     tt.fields.value,
				rangeFrom: tt.fields.rangeFrom,
				rangeTo:   tt.fields.rangeTo,
				Labels:    tt.fields.Labels,
			}
			if got := m.generateValue(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConfigurationMetricItem.generateValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
