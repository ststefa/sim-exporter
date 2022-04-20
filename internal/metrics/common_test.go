package metrics

import (
	"reflect"
	"regexp"
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isInSlice(tt.args.searchString, tt.args.slice); got != tt.want {
				t.Errorf("isInSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isNotInSlice(t *testing.T) {
	type args struct {
		searchString string
		slice        []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNotInSlice(tt.args.searchString, tt.args.slice); got != tt.want {
				t.Errorf("isNotInSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createMatchMap(t *testing.T) {
	type args struct {
		regexp *regexp.Regexp
		line   *string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createMatchMap(tt.args.regexp, tt.args.line); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createMatchMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
