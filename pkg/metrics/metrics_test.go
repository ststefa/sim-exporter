package metrics

import (
	"io/ioutil"
	"math"
	"os"
	"reflect"
	"testing"
	"time"
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

/*
func Test_BuildConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		wantErr bool
	}{
		{
			name: "empty",
			content: []byte(
				"",
			),
			wantErr: true,
		},
		{
			name: "just-metrics",
			content: []byte(
				"metrics:\n",
			),
			wantErr: true,
		},
		{
			name: "empty-metric",
			content: []byte(
				"version: 1\n" +
					"metrics:\n" +
					"  my_metrics:\n",
			),
			wantErr: true,
		},
		{
			name: "invalid-metric-type",
			content: []byte(
				"version: 1\n" +
					"metrics:\n" +
					"  my_metric_a:\n" +
					"    type: foo\n",
			),
			wantErr: true,
		},
		{
			name: "surplus-item-label",
			content: []byte(
				"version: 1\n" +
					"metrics:\n" +
					"  my_metric_b2:\n" +
					"    type: gauge\n" +
					"    labels:\n" +
					"    - l1\n" +
					"    items:\n" +
					"    - value: 1\n" +
					"      labels:\n" +
					"        l1: v1\n" +
					"        l2: v2",
			),
			wantErr: true,
		},
		{
			name: "missing-item-label",
			content: []byte(
				"version: 1\n" +
					"metrics:\n" +
					"  my_metric_b3:\n" +
					"    type: gauge\n" +
					"    labels:\n" +
					"    - l1\n" +
					"    - l2\n" +
					"    items:\n" +
					"    - value: 1\n" +
					"      labels:\n" +
					"        l1: v1\n",
			),
			wantErr: true,
		},
		{
			name: "valid-multi-value-metric",
			content: []byte(
				"version: 1\n" +
					"metrics:\n" +
					"  my_metric_b:\n" +
					"    type: gauge\n" +
					"    labels:\n" +
					"    - l1\n" +
					"    items:\n" +
					"    - value: 100-200\n" +
					"      labels:\n" +
					"        l1: v1\n",
			),
			wantErr: false,
		},
		{
			name: "valid-metric-nolabel-gauge",
			content: []byte(
				"version: 1\n" +
					"metrics:\n" +
					"  my_metric_c:\n" +
					"    type: gauge\n" +
					"    items:\n" +
					"    - value: 1\n",
			),
			wantErr: false,
		},
		{
			name: "valid-metric-nolabel-counter",
			content: []byte(
				"version: 1\n" +
					"metrics:\n" +
					"  my_metric_d:\n" +
					"    type: counter\n" +
					"    items:\n" +
					"    - value: 1\n",
			),
			wantErr: false,
		},
		{
			name: "valid-metric-nolabel-summary",
			content: []byte(
				"version: 1\n" +
					"metrics:\n" +
					"  my_metric_e:\n" +
					"    type: summary\n" +
					"    items:\n" +
					"    - value: 1\n",
			),
			wantErr: false,
		},
		{
			name: "valid-metric-nolabel-histogram",
			content: []byte(
				"version: 1\n" +
					"metrics:\n" +
					"  my_metric_f:\n" +
					"    type: histogram\n" +
					"    items:\n" +
					"    - value: 1\n",
			),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := BuildConfiguration(tt.content)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("BuildConfiguration(%v) error = %v, wantErr %v", tt.name, err, tt.wantErr)
				}
			} else {
				if tt.wantErr {
					t.Errorf("BuildConfiguration(%v) no error but wantErr %v", tt.name, tt.wantErr)
				}
				err = SetupMetricsCollection(config)
				if err != nil {
					t.Error(err)
				}
			}
		})
	}
}
*/
func Test_FromYamlFile(t *testing.T) {
	tests := []struct {
		name    string
		content []string
		want    *Collection
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
				"version: \"1\"",
				"metrics:",
			},
			wantErr: true,
		},
		{
			name: "invalid-metric-type",
			content: []string{
				"version: \"1\"",
				"metrics:",
				"- name: a",
				"  type: foo",
			},
			wantErr: true,
		},
		{
			name: "surplus-item-label",
			content: []string{
				"version: \"1\"",
				"metrics:",
				"- name: b2",
				"  type: gauge",
				"  labels:",
				"  - l1",
				"  items:",
				"  - min: 1",
				"    max: 1",
				"    func: rand",
				"    interval: 1m",
				"    labels:",
				"      l1: v1",
				"      l2: v2",
			},
			wantErr: true,
		},
		{
			name: "missing-item-label",
			content: []string{
				"version: \"1\"",
				"metrics:",
				"- name: b3",
				"  type: gauge",
				"  labels:",
				"  - l1",
				"  - l2",
				"  items:",
				"  - min: 1",
				"    max: 1",
				"    func: rand",
				"    interval: 1m",
				"    labels:",
				"      l1: v1",
			},
			wantErr: true,
		},
		{
			name: "missing-func",
			content: []string{
				"version: \"1\"",
				"metrics:",
				"- name: b4",
				"  type: gauge",
				"  labels:",
				"  - l1",
				"  items:",
				"  - min: 100",
				"    max: 200",
				"    interval: 1m",
				"    labels:",
				"      l1: v1",
			},
			wantErr: true,
		},
		{
			name: "missing-interval",
			content: []string{
				"version: \"1\"",
				"metrics:",
				"- name: b4",
				"  type: gauge",
				"  labels:",
				"  - l1",
				"  items:",
				"  - min: 100",
				"    max: 200",
				"    func: rand",
				"    labels:",
				"      l1: v1",
			},
			wantErr: true,
		},
		{
			name: "valid-metric",
			content: []string{
				"version: \"1\"",
				"metrics:",
				"- name: b4",
				"  type: gauge",
				"  labels:",
				"  - l1",
				"  items:",
				"  - min: 100",
				"    max: 200",
				"    func: rand",
				"    interval: 1m",
				"    labels:",
				"      l1: v1",
			},
			//TODO make result comparison work
			//want: &(Collection{
			//	Version: "1",
			//	Metrics: []*Metric{
			//		&(Metric{
			//			Name:   "b4",
			//			Help:   "",
			//			Type:   "gauge",
			//			Labels: []string{"l1"},
			//			Items: []*MetricItem{&(MetricItem{
			//				Min:      100,
			//				Max:      200,
			//				Func:     "rand",
			//				Interval: 1m,
			//				parent:   nil,
			//				Labels:   map[string]string{"l1": "v1"},
			//			})},
			//			parent: nil,
			//		}),
			//	},
			//}),
			wantErr: false,
		},
		{
			name: "valid-metric-nolabel-gauge",
			content: []string{
				"version: \"1\"",
				"metrics:",
				"- name: c",
				"  type: gauge",
				"  items:",
				"  - min: 1",
				"    max: 1",
				"    func: rand",
				"    interval: 10m",
			},
			//want: &(Collection{
			//	Version: "1",
			//	Metrics: []*Metric{
			//		&(Metric{
			//			Name: "c",
			//			Type: "gauge",
			//			Items: []*MetricItem{&(MetricItem{
			//				Min:    1,
			//				Max:    1,
			//				Func:   "rand",
			//				parent: nil,
			//			})},
			//			parent: nil,
			//		}),
			//	},
			//}),
			wantErr: false,
		},
		{
			name: "valid-metric-nolabel-counter",
			content: []string{
				"version: \"1\"",
				"metrics:",
				"- name: d",
				"  type: counter",
				"  items:",
				"  - min: 1",
				"    max: 1",
				"    func: rand",
				"    interval: 10m",
			},
			wantErr: false,
		},
		{
			name: "valid-metric-nolabel-summary",
			content: []string{
				"version: \"1\"",
				"metrics:",
				"- name: e",
				"  type: summary",
				"  items:",
				"  - min: 1",
				"    max: 1",
				"    func: rand",
				"    interval: 10m",
			},
			wantErr: false,
		},
		{
			name: "valid-metric-nolabel-histogram",
			content: []string{
				"version: \"1\"",
				"metrics:",
				"- name: f",
				"  type: histogram",
				"  items:",
				"  - min: 1",
				"    max: 1",
				"    func: rand",
				"    interval: 1m",
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
			got, err := FromYamlFile(tempFile)
			os.Remove(tempFile)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("FromYamlFile() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else {
				if got != nil && tt.want != nil {
					if !reflect.DeepEqual(*got, *tt.want) {
						t.Errorf("FromYamlFile() = %+v, want %+v", *got, *tt.want)
					}
				}
			}
		})
	}
}

/*
func TestCollection_initialize(t *testing.T) {
	type fields struct {
		Version string
		Metrics map[string]Metric
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wrong metric type",
			fields: fields{
				Version: "1",
				Metrics: map[string]Metric{
					"a": {
						Name: "m-a",
						Help: "help m-a",
						Type: "gauge",
						Items: []MetricItem{
							{
								Value: 0,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "wrong metric type",
			fields: fields{
				Version: "1",
				Metrics: map[string]Metric{
					"a": {
						Name: "m-a",
						Help: "help m-a",
						Type: "foo",
						Items: []MetricItem{
							{
								Value: 0,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty metrics",
			fields: fields{
				Version: "1",
				Metrics: map[string]Metric{},
			},
			wantErr: true,
		},
		{
			name: "no version",
			fields: fields{
				Version: "",
			},
			wantErr: true,
		},
		{
			name: "",
			fields: fields{
				Version: "",
				Metrics: map[string]Metric{
					"a": {
						Name:     "",
						Help:     "",
						Type:     "",
						Labels:   nil,
						Min:      0,
						Max:      0,
						Func:     "",
						Interval: 0,
						Items: []MetricItem{
							{
								Value:  0,
								parent: nil,
								Labels: nil,
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Collection{
				Version: tt.fields.Version,
				Metrics: tt.fields.Metrics,
			}
			if err := c.initialize(); (err != nil) != tt.wantErr {
				t.Errorf("initialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
*/

/*
func TestCollection_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		yamlData []byte
		wantErr  bool
	}{
		{
			name: "empty metric",
			yamlData: []byte(
				"version: \"1\"\n" +
					"metrics:\n" +
					"  m1:\n",
			),
			wantErr: true,
		},
		{
			name: "empty metric",
			yamlData: []byte(
				"version: \"1\"\n" +
					"metrics:\n" +
					"  m1:\n",
			),
			wantErr: true,
		},
		{
			name: "zero metrics",
			yamlData: []byte(
				`version: "1"\n` +
					`metrics:\n`,
			),
			wantErr: true,
		},
		{
			name: "no metrics",
			yamlData: []byte(
				`version: "1"\n`,
			),
			wantErr: true,
		},
		{
			name: "no version",
			yamlData: []byte(
				`metrics:\n`,
			),
			wantErr: true,
		},
		{
			name:     "no yaml",
			yamlData: []byte(`test`),
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Collection{}
			if err := yaml.Unmarshal(tt.yamlData, &c); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
			fmt.Printf("c:%v\n", c)
		})
	}
}
*/

func TestCollection_GetMetric(t *testing.T) {
	type fields struct {
		Version string
		Metrics []*Metric
	}
	tests := []struct {
		name   string
		fields fields
		args   string
		want   *Metric
		ok     bool
	}{
		{
			name: "",
			fields: fields{
				Metrics: []*Metric{
					&(Metric{
						Name: "b",
					}),
				},
			},
			args: "a",
			want: nil,
			ok:   false,
		},
		{
			name: "",
			fields: fields{
				Metrics: []*Metric{
					&(Metric{
						Name: "a",
					}),
				},
			},
			args: "a",
			want: &(Metric{
				Name: "a",
			}),
			ok: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Collection{
				Version: tt.fields.Version,
				Metrics: tt.fields.Metrics,
			}
			got, got1 := c.GetMetric(tt.args)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMetric() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.ok {
				t.Errorf("GetMetric() ok = %v, want %v", got1, tt.ok)
			}
		})
	}
}

func Test_interval(t *testing.T) {
	type args struct {
		start    time.Time
		interval time.Duration
	}
	tests := []struct {
		name    string
		args    args
		want    time.Duration
		wantErr bool
	}{
		{
			name: "123456789s int 100s",
			args: args{
				start:    time.Now().Add(time.Duration(-123456789 * time.Second)),
				interval: time.Duration(100 * time.Second),
			},
			want: time.Duration(89 * time.Second),
		},
		{
			name: "10s int 9s",
			args: args{
				start:    time.Now().Add(time.Duration(-10 * time.Second)),
				interval: time.Duration(9 * time.Second),
			},
			want: time.Duration(1 * time.Second),
		},
		{
			name: "1m int 1m",
			args: args{
				start:    time.Now().Add(time.Duration(-1 * time.Minute)),
				interval: time.Duration(1 * time.Minute),
			},
			want: time.Duration(0 * time.Second),
		},
		{
			name: "1m int 100s",
			args: args{
				start:    time.Now().Add(time.Duration(-1 * time.Minute)),
				interval: time.Duration(100 * time.Second),
			},
			want: time.Duration(1 * time.Minute),
		},
		{
			name: "interval 0",
			args: args{
				start:    time.Now().Truncate(time.Duration(2 * time.Minute)),
				interval: 0,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := interval(tt.args.start, tt.args.interval)
			if (err != nil) != tt.wantErr {
				t.Errorf("interval() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Truncate(time.Duration(time.Second)) != tt.want {
				t.Errorf("interval() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetricItem_generateValue(t *testing.T) {
	type fields struct {
		Min      float64
		Max      float64
		Func     string
		Interval time.Duration
		Labels   map[string]string
		parent   *Metric
	}
	type args struct {
		start time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    float64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &MetricItem{
				Min:      tt.fields.Min,
				Max:      tt.fields.Max,
				Func:     tt.fields.Func,
				Interval: tt.fields.Interval,
				Labels:   tt.fields.Labels,
				parent:   tt.fields.parent,
			}
			got, err := i.generateValue(tt.args.start)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("generateValue() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetricItem_generateValue1(t *testing.T) {
	type fields struct {
		Min      float64
		Max      float64
		Func     string
		Interval time.Duration
	}
	type args struct {
		start time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    float64
		wantErr bool
	}{
		//{
		//	name: "invalid func",
		//	fields: fields{
		//		Min:      10,
		//		Max:      20,
		//		Func:     "foo",
		//		Interval: time.Duration(1*time.Minute + 1*time.Millisecond),
		//	},
		//	args:    args{start: time.Now().Add(time.Duration(-1 * time.Minute))},
		//	wantErr: true,
		//},
		{
			name: "valid desc intvl end",
			fields: fields{
				Min:      10,
				Max:      20,
				Func:     "desc",
				Interval: time.Duration(1*time.Minute + 1*time.Millisecond),
			},
			args:    args{start: time.Now().Add(time.Duration(-1 * time.Minute))},
			want:    10,
			wantErr: false,
		},
		{
			name: "valid desc intvl mid",
			fields: fields{
				Min:      10,
				Max:      20,
				Func:     "desc",
				Interval: time.Duration(1 * time.Minute),
			},
			args:    args{start: time.Now().Add(time.Duration(-30 * time.Second))},
			want:    15,
			wantErr: false,
		},
		{
			name: "valid desc intvl start",
			fields: fields{
				Min:      10,
				Max:      20,
				Func:     "desc",
				Interval: time.Duration(1 * time.Minute),
			},
			args:    args{start: time.Now().Add(time.Duration(-1 * time.Minute))},
			want:    20,
			wantErr: false,
		},
		{
			name: "valid asc intvl end",
			fields: fields{
				Min:      10,
				Max:      20,
				Func:     "asc",
				Interval: time.Duration(1*time.Minute + 1*time.Millisecond),
			},
			args:    args{start: time.Now().Add(time.Duration(-1 * time.Minute))},
			want:    20,
			wantErr: false,
		},
		{
			name: "valid asc intvl mid",
			fields: fields{
				Min:      10,
				Max:      20,
				Func:     "asc",
				Interval: time.Duration(1 * time.Minute),
			},
			args:    args{start: time.Now().Add(time.Duration(-30 * time.Second))},
			want:    15,
			wantErr: false,
		},
		{
			name: "valid asc intvl start",
			fields: fields{
				Min:      10,
				Max:      20,
				Func:     "asc",
				Interval: time.Duration(10 * time.Minute),
			},
			args:    args{start: time.Now().Add(time.Duration(-10 * time.Minute))},
			want:    10,
			wantErr: false,
		},
		{
			name: "valid sin intvl 3/4",
			fields: fields{
				Min:      10,
				Max:      20,
				Func:     "sin",
				Interval: time.Duration(100 * time.Second),
			},
			args:    args{start: time.Now().Add(time.Duration(-75 * time.Second))},
			want:    10,
			wantErr: false,
		},
		{
			name: "valid sin intvl 1/2",
			fields: fields{
				Min:      10,
				Max:      20,
				Func:     "sin",
				Interval: time.Duration(100 * time.Second),
			},
			args:    args{start: time.Now().Add(time.Duration(-50 * time.Second))},
			want:    15,
			wantErr: false,
		},
		{
			name: "valid sin intvl 1/4",
			fields: fields{
				Min:      10,
				Max:      20,
				Func:     "sin",
				Interval: time.Duration(100 * time.Second),
			},
			args:    args{start: time.Now().Add(time.Duration(-25 * time.Second))},
			want:    20,
			wantErr: false,
		},
		{
			name: "valid sin intvl start",
			fields: fields{
				Min:      10,
				Max:      20,
				Func:     "sin",
				Interval: time.Duration(1 * time.Minute),
			},
			args:    args{start: time.Now().Add(time.Duration(-1 * time.Minute))},
			want:    15,
			wantErr: false,
		},
		{
			name: "valid rand",
			fields: fields{
				Min:      10,
				Max:      20,
				Func:     "rand",
				Interval: time.Duration(1 * time.Minute),
			},
			args:    args{start: time.Now().Add(time.Duration(-10 * time.Minute))},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &MetricItem{
				Min:      tt.fields.Min,
				Max:      tt.fields.Max,
				Func:     tt.fields.Func,
				Interval: tt.fields.Interval,
			}
			got, err := i.generateValue(tt.args.start)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.fields.Func == "rand" {
				if got < tt.fields.Min || got > tt.fields.Max {
					t.Errorf("generateValue() got = %v, want %v<got<%v", got, tt.fields.Min, tt.fields.Max)
				}
			} else {
				if math.Round(got) != tt.want {
					t.Errorf("generateValue() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
