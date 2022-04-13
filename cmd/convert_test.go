package cmd

import (
	"reflect"
	"testing"
)

func Test_readLines(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *[]string
		wantErr bool
	}{
		{
			name:    "no-such-file",
			args:    args{path: "no-such-file"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty-file",
			args:    args{path: "testdata/empty-file.txt"},
			want:    &[]string{},
			wantErr: false,
		},
		{
			name: "regular-file",
			args: args{path: "testdata/regular-file.txt"},
			want: &[]string{
				" Leading space",
				"trailing space ",
				"",
				" both space ",
				"no space",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readLines(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("readLines() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readLines() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildConfig(t *testing.T) {
	type args struct {
		scrapeLines *[]string
	}
	tests := []struct {
		name    string
		args    args
		want    *Configuration
		wantErr bool
	}{
		{
			name: "Must start with HELP, not TYPE",
			args: args{scrapeLines: &[]string{
				`# TYPE libvirt_domain_block_stats_allocation gauge`,
				`libvirt_domain_block_stats_allocation{bus="ide",cache="writeback",domain="instance-aaaaa"} 1`},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Must start with HELP, not metric",
			args: args{scrapeLines: &[]string{
				`libvirt_domain_block_meta{bus="ide",cache="writeback",domain="instance-aaaaa"} 1`,
				`libvirt_domain_block_meta{bus="ide",cache="writeback",domain="instance-bbbbb"} 1`},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "TYPE must have same name like HELP",
			args: args{scrapeLines: &[]string{
				`# HELP libvirt_domain_block_meta Block device metadata info. Device name, source file, serial.`,
				`# TYPE some_other_metric gauge`,
				`libvirt_domain_block_meta{bus="ide",cache="writeback",domain="instance-aaaaa"} 1`},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Metrics with name other than announced in HELP/TYPE are dropped",
			args: args{scrapeLines: &[]string{
				`# HELP my_metric This is a metric`,
				`# TYPE my_metric gauge`,
				`my_metric{foo="lion",instance="aaa"} 1`,
				`my_metric{foo="lion",instance="bbb"} 2`,
				`some_other_metric{foo="lion",instance="aaa"} 3`},
			},
			want: &Configuration{
				Version: "1",
				Metrics: map[string]ConfigurationMetric{
					"my_metric": {
						Name:   "my_metric",
						Help:   "This is a metric",
						Type:   "gauge",
						Labels: []string{"foo", "instance"},
						Items: []ConfigurationMetricItem{
							{
								Value: "1",
								Labels: map[string]string{
									"foo": "lion", "instance": "aaa",
								},
							},
							{
								Value: "2",
								Labels: map[string]string{
									"foo": "lion", "instance": "bbb",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Metrics without value are dropped",
			args: args{scrapeLines: &[]string{
				`# HELP my_metric This is a metric`,
				`# TYPE my_metric gauge`,
				`my_metric{foo="lion",instance="aaa"} 1`,
				`my_metric{foo="lion",instance="bbb"}`},
			},
			want: &Configuration{
				Version: "1",
				Metrics: map[string]ConfigurationMetric{
					"my_metric": {
						Name:   "my_metric",
						Help:   "This is a metric",
						Type:   "gauge",
						Labels: []string{"foo", "instance"},
						Items: []ConfigurationMetricItem{
							{
								Value: "1",
								Labels: map[string]string{
									"foo": "lion", "instance": "aaa",
								},
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
			got, err := buildConfig(tt.args.scrapeLines)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_stripQuotes(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Empty in, empty out",
			args: args{s: ""},
			want: "",
		},
		{
			name: "Single quote",
			args: args{s: `"`},
			want: "",
		},
		{
			name: "Only opening quote",
			args: args{s: `"test`},
			want: "test",
		},
		{
			name: "Only closing quote",
			args: args{s: `test"`},
			want: "test",
		},
		{
			name: "Enclosing quotes",
			args: args{s: `"test"`},
			want: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripQuotes(tt.args.s); got != tt.want {
				t.Errorf("stripQuotes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_generateValueRange(t *testing.T) {
	type args struct {
		value     string
		deviation int
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "1 dev 10",
			args:    args{value: "1", deviation: 10},
			want:    "1",
			wantErr: false,
		},
		{
			name:    "10 dev 10",
			args:    args{value: "10", deviation: 10},
			want:    "9-11",
			wantErr: false,
		},
		{
			name:    "Empty in, empty out",
			args:    args{value: "", deviation: 0},
			want:    "0",
			wantErr: true,
		},
		{
			name:    "Not a number",
			args:    args{value: "abc", deviation: 0},
			want:    "0",
			wantErr: true,
		},
		{
			name:    "1",
			args:    args{value: "1", deviation: 0},
			want:    "1",
			wantErr: false,
		},
		{
			name:    "10000",
			args:    args{value: "10000", deviation: 0},
			want:    "1.000e+04",
			wantErr: false,
		},
		{
			name:    "10000 dev 50",
			args:    args{value: "10000", deviation: 30},
			want:    "7000-13000",
			wantErr: false,
		},
		{
			name:    "100000 dev 50",
			args:    args{value: "100000", deviation: 30},
			want:    "7.000e+04-1.300e+05",
			wantErr: false,
		},
		{
			name:    "1000 dev 10",
			args:    args{value: "10000", deviation: 10},
			want:    "9000-11000",
			wantErr: false,
		},
		{
			name:    "1000 dev 0",
			args:    args{value: "1000", deviation: 0},
			want:    "1000",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateValueRange(tt.args.value, tt.args.deviation)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("generateValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
