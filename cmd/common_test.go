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
func TestReadConfig(t *testing.T) {
	_, err := readAndValidateConfig("testdata/config_valid.yaml")
	require.NoError(t, err)
}

func TestReadConfigMissing(t *testing.T) {
	_, err := readAndValidateConfig("noSuchFile")
	var targetError *os.PathError
	require.ErrorAsf(t, err, &targetError, "Should return '%T', but returned '%s'", targetError, err)
}

func TestReadConfigUnparsable(t *testing.T) {
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

func TestReadConfigInvalid(t *testing.T) {
	genConf := Config{
		Metrics: []Metric{
			{
				Name:       "ma",
				Func:       "x",
				Fuzzy:      -1,
				LowerLimit: 100,
				UpperLimit: 10,
			},
		},
		InstanceTypes: []InstanceType{
			{
				Name: "ia",
				MetricRefs: []MetricRef{
					{
						Ref: "a",
					},
				},
			},
		},
		Fleet: []Fleet{
			{
				Kind: "ib",
				Num:  1,
			},
		},
	}
	yamlData, err := yaml.Marshal(&genConf)
	require.NoError(t, err, "Error while marshaling")
	fileName, err := generateTempConfig(string(yamlData))
	require.NoError(t, err, "Cannot create tempfile")
	defer os.Remove(fileName)
	_, err = readAndValidateConfig(fileName)
	var targetError *ValidationError
	require.ErrorAsf(t, err, &targetError, "Should return '%T', but returned '%s'", targetError, err)
}
