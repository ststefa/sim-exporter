package cmd

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/spf13/cobra"
)

var (
	outfile_help        = "Where to write the output"
	outfile      string = "/dev/stdout"

	maxdeviation_help     = "How many percent to deviate from converted value at most"
	maxdeviation      int = 10

	convertCmd = &cobra.Command{
		Use:   "convert <prometheus-scrape-file>",
		Short: "Parse prometheus-style scrape file and create simulator yaml config",
		Long:  "Parses data in prometheus scrape format read from <prometheus-scrape-file> and turns it into a yaml structure suitable as input for the simulator",
		Args:  cobra.ExactArgs(1),
		Run:   doConvert,
	}

	// How a prometheus metric line must look like
	regexpMetricItem = regexp.MustCompile(`^(?P<name>\w+)(?:|{(?P<labels>[^}]*)})\s+(?P<value>[^\s]*).*$`)
)

func init() {
	convertCmd.Flags().StringVar(&outfile, "outfile", outfile, outfile_help)
	convertCmd.Flags().IntVar(&maxdeviation, "maxdeviation", maxdeviation, maxdeviation_help)

	rootCmd.AddCommand(convertCmd)
}

// https://golangdocs.com/golang-read-file-line-by-line
func readLines(path string) (*[]string, error) {
	readFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	//var result []string
	result := make([]string, 0)

	for fileScanner.Scan() {
		result = append(result, fileScanner.Text())
	}
	readFile.Close()

	//if result==nil{
	//	return string
	//}
	return &result, nil
}

func stripQuotes(s string) string {
	if len(s) > 0 && s[0] == '"' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '"' {
		s = s[:len(s)-1]
	}
	return s
}

func generateValueRange(value string, deviation int) (string, error) {
	formatString := func(val float64) string {
		if val < 10000 {
			return "%.0f"
		} else {
			return "%.3e"
		}
	}
	floatVal, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return "0", err
	}
	if deviation == 0 {
		return fmt.Sprintf(formatString(floatVal), floatVal), nil
	} else {
		lower := floatVal - (floatVal * float64(deviation) / 100)
		upper := floatVal + (floatVal * float64(deviation) / 100)
		if fmt.Sprintf(formatString(lower), lower) == fmt.Sprintf(formatString(upper), upper) {
			return fmt.Sprintf(formatString(floatVal), floatVal), nil
		} else {
			// upper and lower values should be formatted identically regardless of "10000-rule"
			format := fmt.Sprintf("%s-%s", formatString(lower), formatString(lower))
			return fmt.Sprintf(format, lower, upper), nil
		}
	}

}

func buildConfig(scrapeLines *[]string) (*Configuration, error) {

	config := Configuration{
		Version: "1",
		Metrics: map[string]ConfigurationMetric{},
	}

	// Defines order in which metric properties must appear
	expectHelp := true
	expectType := false
	expectMetric := false
	// Skip predefined prometheus-internal metrics
	skipMetric := false

	var metric ConfigurationMetric
	var metricName string

	for lineno, line := range *scrapeLines {
		log.Debugf("%d: %v\n", lineno, line)
		if strings.HasPrefix(line, "# HELP") {
			// line e.g. '# HELP libvirt_domain_interface_meta Interfaces metadata. Source bridge, target device, interface uuid'
			if !expectHelp {
				return nil, &SimulationError{fmt.Sprintf("Line %v: Unexpected 'HELP' (no 'TYPE' since last 'HELP')", lineno)}
			}
			expectHelp = false
			expectType = true
			fields := strings.Split(line, " ")
			metricName = fields[2]
			if strings.HasPrefix(metricName, "go_") || strings.HasPrefix(metricName, "process_") || strings.HasPrefix(metricName, "promhttp_") {
				skipMetric = true
				log.Infof("Line %vff: Skipping prometheus-internal metric %q", lineno, metricName)
			} else {
				skipMetric = false
			}
			if !skipMetric {

				mHelp := strings.Join(fields[3:], " ")
				metric = ConfigurationMetric{
					Name: metricName,
					Help: mHelp,
				}
				config.Metrics[metricName] = metric
			}
		} else if strings.HasPrefix(line, "# TYPE") {
			// line e.g. '# TYPE libvirt_domain_block_meta gauge'
			if !expectType {
				return nil, &SimulationError{fmt.Sprintf("Line %v: Unexpected 'TYPE' (not preceded by 'HELP')", lineno)}
			}
			expectType = false
			expectMetric = true
			fields := strings.Split(line, " ")

			if fields[2] != metricName {
				return nil, &SimulationError{fmt.Sprintf("Line %v: Out-of-order line %q (expecting 'TYPE %v ...')", lineno, line, metricName)}
			}
			if !skipMetric {
				mType := fields[3]
				// this can probably be done using by-reference invocation but I'm still too stupid ;)
				tmpMetric := config.Metrics[metricName]
				tmpMetric.Type = mType
				config.Metrics[metricName] = tmpMetric
				//if err := setField(config.Metrics[mName], "Type", mType); err != nil {
				//	return nil, &SimulationError{fmt.Sprintf("%v", err)}
				//}
			}
		} else {
			// line e.g. 'libvirt_domain_block_meta{bus="scsi",cache="writeback",discard="unmap",disk_type="network",domain="instance-0000bfa6",driver_type="raw",flavor="m1.small",instance_name="zabbix-prod",project_name="C00061-Nexible",project_uuid="077224dcfd454436987147de7d86fa89",root_type="image",root_uuid="fd8ad5aa-6b33-4198-a05d-8be42fc0f20e",serial="",source_file="ephemeral-vms/ed1ce34e-0200-4f2e-a0cc-2b60216e1362_disk",target_device="sda",user_name="OSieben@nexible.de",user_uuid="02a474d230b04749b882362909c79502",uuid="ed1ce34e-0200-4f2e-a0cc-2b60216e1362"} 1'
			if !expectMetric {
				return nil, &SimulationError{fmt.Sprintf("Line %v: Unexpected %q (not preceded by 'HELP')", lineno, line)}
			}
			expectType = false
			expectHelp = true

			if !skipMetric {
				matchMap := createMatchMap(regexpMetricItem, &line)
				if len(matchMap) >= 1 {
					if matchMap["name"] != metricName {
						log.Infof("Line %v: Ignoring out-of-order line %q (expecting '%v ...')", lineno, line, metricName)
						continue
					}

					tmpMetric := config.Metrics[metricName]
					var item ConfigurationMetricItem

					randInt, err := rand.Int(rand.Reader, big.NewInt(int64(maxdeviation)))
					if err != nil {
						return nil, &SimulationError{fmt.Sprintf("Line %v: Cannot generate random number", lineno)}
					}
					value, err := generateValueRange(matchMap["value"], int(randInt.Int64()))
					if err != nil {
						log.Infof("Line %v: Skipping metric item %q (cannot parse value %q)", lineno, line, matchMap["value"])
						continue
					} else {
						item.Value = value
					}

					if matchMap["labels"] != "" {
						var labelNames []string
						labels := make(map[string]string)
						for _, labelKv := range strings.Split(matchMap["labels"], ",") {
							kvList := strings.Split(labelKv, "=")
							labelNames = append(labelNames, stripQuotes(kvList[0]))
							labels[stripQuotes(kvList[0])] = stripQuotes(kvList[1])
						}
						tmpMetric.Labels = labelNames
						item.Labels = labels
					}
					tmpMetric.Items = append(tmpMetric.Items, item)
					config.Metrics[metricName] = tmpMetric
				} else {
					log.Infof("Line %v: Skipping unparsable metric item %q (must match regex %q)", lineno, line, regexpMetricItem)
					continue
				}
			}
		}
	}
	return &config, nil
}

// Any undesired but handled outcome is signaled by panicking with SimulationError
func doConvert(cmd *cobra.Command, args []string) {
	//var config Configuration

	scrapeLines, err := readLines(args[0])
	if err != nil {
		panic(&SimulationError{err.Error()})
	}

	config, err := buildConfig(scrapeLines)
	if err != nil {
		panic(&SimulationError{err.Error()})
	}

	yamlData, err := yaml.Marshal(config)
	if err != nil {
		panic(&SimulationError{err.Error()})
	}

	err = ioutil.WriteFile(outfile, yamlData, 0644)
	if err != nil {
		panic(&SimulationError{err.Error()})
	} else {
		fmt.Printf("Created config file %v\n", outfile)
	}
}
