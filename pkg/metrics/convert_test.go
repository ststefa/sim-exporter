package metrics

import (
	"reflect"
	"strings"
	"testing"
	"time"
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
			name: "no-such-file",
			args: args{
				path: "no-such-file",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "empty-file",
			args: args{
				path: "testdata/empty-file.txt",
			},
			want:    &[]string{},
			wantErr: false,
		},
		{
			name: "regular-file",
			args: args{
				path: "testdata/regular-file.txt",
			},
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

func Test_convertScrapeToConfig(t *testing.T) {
	type args struct {
		scrapeLines *[]string
	}
	tests := []struct {
		name    string
		args    args
		want    *Collection
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
			name: "Metrics with name other than announced in HELP or TYPE are dropped",
			args: args{scrapeLines: &[]string{
				`# HELP my_metric This is a metric`,
				`# TYPE my_metric gauge`,
				`my_metric{foo="lion",instance="aaa"} 1`,
				`my_metric{foo="lion",instance="bbb"} 2`,
				`some_other_metric{foo="lion",instance="aaa"} 3`},
			},
			//TODO make result comparison work
			//want: &Collection{
			//	Version: "1",
			//	Metrics: []*Metric{
			//		{
			//			Name:   "my_metric",
			//			Help:   "This is a metric",
			//			Type:   "gauge",
			//			Labels: []string{"foo", "instance"},
			//			Items: []*MetricItem{
			//				{
			//					Min:  1,
			//					Max:  1,
			//					Func: "rand",
			//					Labels: map[string]string{
			//						"foo": "lion", "instance": "aaa",
			//					},
			//				},
			//				{
			//					Min:  2,
			//					Max:  2,
			//					Func: "rand",
			//					Labels: map[string]string{
			//						"foo": "lion", "instance": "bbb",
			//					},
			//				},
			//			},
			//		},
			//	},
			//},
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
			//want: &Collection{
			//	Version: "1",
			//	Metrics: []*Metric{
			//		{
			//			Name:   "my_metric",
			//			Help:   "This is a metric",
			//			Type:   "gauge",
			//			Labels: []string{"foo", "instance"},
			//			Items: []*MetricItem{
			//				{
			//					Min:  1,
			//					Max:  1,
			//					Func: "rand",
			//					Labels: map[string]string{
			//						"foo": "lion", "instance": "aaa",
			//					},
			//				},
			//			},
			//		},
			//	},
			//},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertScrapeToConfig(tt.args.scrapeLines, 10, "rand", "1s-1s", "percent")
			if err != nil {
				if !tt.wantErr {
					t.Errorf("convertScrapeToConfig() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if tt.want != nil {
				if !reflect.DeepEqual(*got, *tt.want) {
					t.Errorf("convertScrapeToConfig() = %+v, want %+v", *got, *tt.want)
				}
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

//func Test_ConvertScrapefileToYaml(t *testing.T) {
//	type args struct {
//		filename  string
//		deviation int
//	}
//	tests := []struct {
//		name    string
//		args    args
//		want    []byte
//		wantErr bool
//	}{
//		{
//			name:    "1 dev 10",
//			args:    args{
//				filename: "1",
//				deviation: 10,
//			},
//			want:    "1",
//			wantErr: false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got, err := ScrapefileToCollection(tt.args.filename, tt.args.deviation)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("generateValueRange() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if got != tt.want {
//				t.Errorf("generateValueRange() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

func Test_randomFunc(t *testing.T) {
	tests := []struct {
		name     string
		function string
		want     string
		wantErr  bool
	}{
		{
			name:     "rand,asc,desc,sin",
			function: "rand,asc,desc,sin",
			wantErr:  false,
		},
		{
			name:     "sin,foo",
			function: "sin,foo",
			wantErr:  true,
		},
		{
			name:     "rand,desc,",
			function: "rand,desc,",
			wantErr:  true,
		},
		{
			name:     "sin",
			function: "sin",
			want:     "sin",
			wantErr:  false,
		},
		{
			name:     "rand",
			function: "rand",
			want:     "rand",
			wantErr:  false,
		},
		{
			name:     "rand,",
			function: "rand,",
			wantErr:  true,
		},
		{
			name:     "unknown func",
			function: "foo",
			wantErr:  true,
		},
		{
			name:     "empty",
			function: "",
			want:     "",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := randomFunc(tt.function)
			if (err != nil) != tt.wantErr {
				t.Errorf("randomFunc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != "" && got != tt.want {
				t.Errorf("randomFunc() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_randomDuration(t *testing.T) {
	tests := []struct {
		interval string
		wantErr  bool
	}{
		{
			interval: "20s-2h",
			wantErr:  false,
		},
		{
			interval: "20m-2h",
			wantErr:  false,
		},
		{
			interval: "20s-30s",
			wantErr:  false,
		},
		{
			interval: "10000m-7d",
			wantErr:  true,
		},
		{
			interval: "10s-30s",
			wantErr:  true,
		},
		{
			interval: "100m-1h",
			wantErr:  true,
		},
		{
			interval: "20s-30",
			wantErr:  true,
		},
		{
			interval: "20s-10s",
			wantErr:  true,
		},
		{
			interval: "0s-20s",
			wantErr:  true,
		},
		{
			interval: "s-s",
			wantErr:  true,
		},
		{
			interval: "1-2",
			wantErr:  true,
		},
		{
			interval: "1-",
			wantErr:  true,
		},
		{
			interval: "-",
			wantErr:  true,
		},
		{
			interval: "-1",
			wantErr:  true,
		},
		{
			interval: "",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.interval, func(t *testing.T) {
			got, err := randomDuration(tt.interval)
			if (err != nil) != tt.wantErr {
				t.Errorf("randomDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if err == nil {
				boundaries := strings.Split(tt.interval, "-")
				min, _ := time.ParseDuration(boundaries[0])
				max, _ := time.ParseDuration(boundaries[1])

				if got < min {
					t.Errorf("randomDuration() got = %v, want min %v", got, min)
				}
				if got > max {
					t.Errorf("randomDuration() got = %v, want max %v", got, max)
				}
			}
		})
	}
}

func Test_randomRange(t *testing.T) {
	type args struct {
		value        float64
		maxDeviation int
	}
	tests := []struct {
		name    string
		args    args
		wantMin float64
		wantMax float64
	}{
		{
			name: "10-100",
			args: args{
				value:        10,
				maxDeviation: 100,
			},
			wantMin: 0,
			wantMax: 20,
		},
		{
			name: "10-10",
			args: args{
				value:        10,
				maxDeviation: 10,
			},
			wantMin: 9,
			wantMax: 11,
		},
		{
			name: "42-0",
			args: args{
				value:        42,
				maxDeviation: 0,
			},
			wantMin: 42,
			wantMax: 42,
		},
		{
			name: "0-0",
			args: args{
				value:        0,
				maxDeviation: 0,
			},
			wantMin: 0,
			wantMax: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMin, gotMax := randomRange(tt.args.value, tt.args.maxDeviation)
			if gotMin < tt.wantMin {
				t.Errorf("randomRange() gotMin = %v, want %v", gotMin, tt.wantMin)
			}
			if gotMax > tt.wantMax {
				t.Errorf("randomRange() gotMax = %v, want %v", gotMax, tt.wantMax)
			}
		})
	}
}

func Test_isPercent(t *testing.T) {
	type args struct {
		metricName string
		honorpct   string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "4",
			args: args{
				metricName: "brabaz",
				honorpct:   "foo,bar",
			},
			want: false,
		},
		{
			name: "3",
			args: args{
				metricName: "brabaz",
				honorpct:   "foo,bar,baz",
			},
			want: true,
		},
		{
			name: "2",
			args: args{
				metricName: "meabaran",
				honorpct:   "foo,bar",
			},
			want: true,
		},
		{
			name: "1",
			args: args{
				metricName: "foo_met",
				honorpct:   "foo",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPercent(tt.args.metricName, tt.args.honorpct); got != tt.want {
				t.Errorf("isPercent() = %v, want %v", got, tt.want)
			}
		})
	}
}
