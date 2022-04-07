package cmd

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func generateTempConfig(content string) (fileName string, err error) {

	f, err := ioutil.TempFile(os.TempDir(), "sim-exporter-config-*.yaml")
	if err != nil {
		return "", errors.Wrap(err, "Cannot create temp file")
	}
	f.WriteString(content)

	return f.Name(), nil
}

func TestReadAndValidateConfig(t *testing.T) {
	_, err := readAndValidateConfig("testdata/config_valid.yaml")
	require.NoError(t, err)
}

func TestReadAndValidateConfigMissing(t *testing.T) {
	_, err := readAndValidateConfig("noSuchFile")
	var targetError *os.PathError
	require.ErrorAsf(t, err, &targetError, "Should return '%T', but returned '%s'", targetError, err)
}

func TestReadAndValidateConfigUnparsable(t *testing.T) {
	var (
		err      error
		fileName string
	)
	fileName, err = generateTempConfig("this is not yaml")
	require.NoError(t, err, "Cannot create tempfile")
	_, err = readAndValidateConfig(fileName)
	var targetError *yaml.TypeError
	require.ErrorAsf(t, err, &targetError, "Should return '%T', but returned '%s'", targetError, err)
}

// Lengthy test of invalid inputs (don't call it overengineering, it's just diligent ;))
func TestReadAndValidateConfigInvalid(t *testing.T) {

	var invalidConfigs = make(map[string]Config)
	invalidConfigs["Empty"] = Config{}
	invalidConfigs["NoMetrics"] = Config{
		InstanceTypes: []InstanceType{
			{
				Name: "ia",
				MetricRefs: []MetricRef{
					{
						Ref: "ma",
					},
				},
			},
		},
		Fleet: []Fleet{
			{
				Kind: "ia",
				Num:  1,
			},
		},
	}
	invalidConfigs["NoInstanceType"] = Config{
		Metrics: []Metric{
			{
				Name:       "ma",
				Func:       "x",
				Fuzzy:      0,
				LowerLimit: 0,
				UpperLimit: 100,
			},
		},
		Fleet: []Fleet{
			{
				Kind: "ia",
				Num:  1,
			},
		},
	}
	invalidConfigs["NoFleet"] = Config{
		Metrics: []Metric{
			{
				Name:       "ma",
				Func:       "x",
				Fuzzy:      0,
				LowerLimit: 0,
				UpperLimit: 10,
			},
		},
		InstanceTypes: []InstanceType{
			{
				Name: "ia",
				MetricRefs: []MetricRef{
					{
						Ref: "ma",
					},
				},
			},
		},
	}
	invalidConfigs["FuzzyTooSmall"] = Config{
		Metrics: []Metric{
			{
				Name:       "ma",
				Func:       "x",
				Fuzzy:      -1,
				LowerLimit: 0,
				UpperLimit: 10,
			},
		},
		InstanceTypes: []InstanceType{
			{
				Name: "ia",
				MetricRefs: []MetricRef{
					{
						Ref: "ma",
					},
				},
			},
		},
		Fleet: []Fleet{
			{
				Kind: "ia",
				Num:  1,
			},
		},
	}
	invalidConfigs["LowerEqualUpper"] = Config{
		Metrics: []Metric{
			{
				Name:       "ma",
				Func:       "x",
				Fuzzy:      0,
				LowerLimit: 0,
				UpperLimit: 0,
			},
		},
		InstanceTypes: []InstanceType{
			{
				Name: "ia",
				MetricRefs: []MetricRef{
					{
						Ref: "ma",
					},
				},
			},
		},
		Fleet: []Fleet{
			{
				Kind: "ia",
				Num:  1,
			},
		},
	}
	invalidConfigs["LowerAboveUpper"] = Config{
		Metrics: []Metric{
			{
				Name:       "ma",
				Func:       "x",
				Fuzzy:      0,
				LowerLimit: 10,
				UpperLimit: 0,
			},
		},
		InstanceTypes: []InstanceType{
			{
				Name: "ia",
				MetricRefs: []MetricRef{
					{
						Ref: "ma",
					},
				},
			},
		},
		Fleet: []Fleet{
			{
				Kind: "ia",
				Num:  1,
			},
		},
	}
	invalidConfigs["InvalidMetricRef"] = Config{
		Metrics: []Metric{
			{
				Name:       "ma",
				Func:       "x",
				Fuzzy:      0,
				LowerLimit: 0,
				UpperLimit: 10,
			},
		},
		InstanceTypes: []InstanceType{
			{
				Name: "ia",
				MetricRefs: []MetricRef{
					{
						Ref: "no-exist",
					},
				},
			},
		},
		Fleet: []Fleet{
			{
				Kind: "ia",
				Num:  1,
			},
		},
	}
	invalidConfigs["InvalidTypeRef"] = Config{
		Metrics: []Metric{
			{
				Name:       "ma",
				Func:       "x",
				Fuzzy:      0,
				LowerLimit: 0,
				UpperLimit: 10,
			},
		},
		InstanceTypes: []InstanceType{
			{
				Name: "ia",
				MetricRefs: []MetricRef{
					{
						Ref: "ma",
					},
				},
			},
		},
		Fleet: []Fleet{
			{
				Kind: "no-exist",
				Num:  1,
			},
		},
	}
	invalidConfigs["NumTooSmall"] = Config{
		Metrics: []Metric{
			{
				Name:       "ma",
				Func:       "x",
				Fuzzy:      0,
				LowerLimit: 0,
				UpperLimit: 10,
			},
		},
		InstanceTypes: []InstanceType{
			{
				Name: "ia",
				MetricRefs: []MetricRef{
					{
						Ref: "ma",
					},
				},
			},
		},
		Fleet: []Fleet{
			{
				Kind: "ia",
				Num:  -1,
			},
		},
	}

	for name, config := range invalidConfigs {
		yamlData, err := yaml.Marshal(&config)
		require.NoError(t, err, "Error while marshaling")
		fileName, err := generateTempConfig(string(yamlData))
		require.NoError(t, err, "Cannot create tempfile")
		defer os.Remove(fileName)
		_, err = readAndValidateConfig(fileName)
		var targetError *ValidationError
		require.ErrorAsf(t, err, &targetError, "'%v' should return '%T', but returned '%s'", name, targetError, err)
	}
}
