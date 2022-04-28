package metrics

import (
	"reflect"
	"testing"
)

func Test_isInSlice(t *testing.T) {
	type args struct {
		searchString string
		slice        []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "empty in empty",
			args: args{
				searchString: "",
				slice:        []string{""},
			},
			want: true,
		},
		{
			name: "empty in value",
			args: args{
				searchString: "",
				slice:        []string{"a"},
			},
			want: false,
		},
		{
			name: "a in a",
			args: args{
				searchString: "a",
				slice:        []string{"a"},
			},
			want: true,
		},
		{
			name: "b in a",
			args: args{
				searchString: "b",
				slice:        []string{"a"},
			},
			want: false,
		},
		{
			name: "a in a,b",
			args: args{
				searchString: "a",
				slice:        []string{"a", "b"},
			},
			want: true,
		},
		{
			name: "a in b,a",
			args: args{
				searchString: "a",
				slice:        []string{"b", "a"},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isInSlice(tt.args.searchString, tt.args.slice); got != tt.want {
				t.Errorf("isInSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_stringSlicesEqual(t *testing.T) {
	type args struct {
		a []string
		b []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "a,b:a,b",
			args: args{
				a: []string{"a", "b"},
				b: []string{"a", "b"},
			},
			want: true,
		},
		{
			name: "a:a",
			args: args{
				a: []string{"a"},
				b: []string{"a"},
			},
			want: true,
		},
		{
			name: ":",
			args: args{
				a: []string{""},
				b: []string{""},
			},
			want: true,
		},
		{
			name: "a,b:b,a",
			args: args{
				a: []string{"a", "b"},
				b: []string{"b", "a"},
			},
			want: false,
		},
		{
			name: "a:a,b",
			args: args{
				a: []string{"a"},
				b: []string{"a", "b"},
			},
			want: false,
		},
		{
			name: "a,b:b",
			args: args{
				a: []string{"a", "b"},
				b: []string{"b"},
			},
			want: false,
		},
		{
			name: "a:b",
			args: args{
				a: []string{"a"},
				b: []string{"b"},
			},
			want: false,
		},
		{
			name: ":a",
			args: args{
				a: []string{""},
				b: []string{"a"},
			},
			want: false,
		},
		{
			name: "a:",
			args: args{
				a: []string{"a"},
				b: []string{""},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stringSlicesEqual(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("stringSlicesEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createMatchMap(t *testing.T) {
	tests := []struct {
		line string
		want map[string]string
	}{
		{
			line: `my_metric{l1="v1",l2="v2"} 1`,
			want: map[string]string{
				"name":   "my_metric",
				"labels": `l1="v1",l2="v2"`,
				"value":  "1",
			},
		},
		{
			line: `my_metric{l1="v1",l2="v2"} 1 234`,
			want: map[string]string{
				"name":   "my_metric",
				"labels": `l1="v1",l2="v2"`,
				"value":  "1",
			},
		},
		{
			line: `my_metric { l1 = "v1" , l2 = "v2" } 1`,
			want: map[string]string{
				"name":   "my_metric",
				"labels": ` l1 = "v1" , l2 = "v2" `,
				"value":  "1",
			},
		},
		{
			line: `my_metric {l1="v1",l2="v2"} 1`,
			want: map[string]string{
				"name":   "my_metric",
				"labels": `l1="v1",l2="v2"`,
				"value":  "1",
			},
		},
		{
			line: `my_metric{l1="v1",l2="v2"} 1`,
			want: map[string]string{
				"name":   "my_metric",
				"labels": `l1="v1",l2="v2"`,
				"value":  "1",
			},
		},
		{
			line: "my_metric{} 1",
			want: map[string]string{
				"name":   "my_metric",
				"labels": "",
				"value":  "1",
			},
		},
		{
			line: "my_metric 1 234",
			want: map[string]string{
				"name":   "my_metric",
				"labels": "",
				"value":  "1",
			},
		},
		{
			line: "my_metric 1",
			want: map[string]string{
				"name":   "my_metric",
				"labels": "",
				"value":  "1",
			},
		},
		{
			line: "my_metric{}",
			want: map[string]string{},
		},
		{
			line: "my_metric",
			want: map[string]string{},
		},
		{
			line: "",
			want: map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			if got := createMatchMap(regexpMetricItem, tt.line); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createMatchMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
