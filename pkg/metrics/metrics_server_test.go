package metrics

import "testing"

func Test_SetupMetricsCollection(t *testing.T) {
	tests := []struct {
		filename string
		wantErr  bool
	}{
		{
			filename: "testdata/valid_scrape.yaml",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			c, _ := FromYamlFile(tt.filename)
			if err := SetupMetricsCollection(c); (err != nil) != tt.wantErr {
				t.Errorf("SetupMetricsCollection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
