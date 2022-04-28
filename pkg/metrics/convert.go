package metrics

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// https://golangdocs.com/golang-read-file-line-by-line
func readLines(path string) (*[]string, error) {
	readFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	result := make([]string, 0)

	for fileScanner.Scan() {
		result = append(result, fileScanner.Text())
	}
	readFile.Close()

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

func convertScrapeToConfig(scrapeLines *[]string, maxdeviation int, function string, interval string, honorpct string) (*Collection, error) {

	c := Collection{
		Version: "1",
	}

	// Defines order in which metric properties must appear
	expectHelp := true
	expectType := false
	expectMetric := false

	// Skip predefined prometheus-internal metrics
	skipMetric := false

	var metricName string

	for lineno, line := range *scrapeLines {
		log.Debugf("%d: %v\n", lineno, line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "# HELP") {
			// line e.g. '# HELP libvirt_domain_interface_meta Interfaces metadata. Source bridge, target device, interface uuid'
			if !expectHelp {
				return nil, fmt.Errorf("line %v: Unexpected 'HELP' (no 'TYPE' since last 'HELP')", lineno)
			}

			expectHelp = false
			expectType = true

			fields := strings.Split(line, " ")
			metricName = fields[2]
			if strings.HasPrefix(metricName, "go_") || strings.HasPrefix(metricName, "process_") || strings.HasPrefix(metricName, "promhttp_") {
				skipMetric = true
				log.Infof("line %vff: Skipping prometheus-internal metric %q", lineno, metricName)
			} else {
				skipMetric = false
			}
			if skipMetric {
				continue
			}
			mHelp := strings.Join(fields[3:], " ")

			m := Metric{
				Name: metricName,
				Help: mHelp,
			}
			err := c.AddMetric(m)
			if err != nil {
				return nil, fmt.Errorf("line %v: %v", lineno, err)
			}
		} else if strings.HasPrefix(line, "# TYPE") {
			// line e.g. '# TYPE libvirt_domain_block_meta gauge'
			if !expectType {
				return nil, fmt.Errorf("line %v: Unexpected 'TYPE' (not preceded by 'HELP')", lineno)
			}

			expectType = false
			expectMetric = true
			expectHelp = true

			fields := strings.Split(line, " ")

			if fields[2] != metricName {
				return nil, fmt.Errorf("line %v: Out-of-order line '%v' (expecting 'TYPE %v ...')", lineno, line, metricName)
			}
			if skipMetric {
				continue
			}
			mType := fields[3]
			if m, ok := c.GetMetric(metricName); ok {
				m.Type = mType
			} else {
				return nil, fmt.Errorf("line %v: Metric %q unknown", lineno, metricName)
			}
		} else {
			// line e.g. 'libvirt_domain_block_meta{bus="scsi",cache="writeback",discard="unmap",disk_type="network",domain="instance-0000bfa6",driver_type="raw",flavor="m1.small",instance_name="zabbix-prod",project_name="C00061-Nexible",project_uuid="077224dcfd454436987147de7d86fa89",root_type="image",root_uuid="fd8ad5aa-6b33-4198-a05d-8be42fc0f20e",serial="",source_file="ephemeral-vms/ed1ce34e-0200-4f2e-a0cc-2b60216e1362_disk",target_device="sda",user_name="OSieben@nexible.de",user_uuid="02a474d230b04749b882362909c79502",uuid="ed1ce34e-0200-4f2e-a0cc-2b60216e1362"} 1'
			if !expectMetric {
				return nil, fmt.Errorf("line %v: Unexpected '%v' (not preceded by 'HELP')", lineno, line)
			}

			expectType = false
			expectHelp = true
			expectMetric = true

			if skipMetric {
				continue
			}
			matchMap := createMatchMap(regexpMetricItem, line)
			if len(matchMap) == 0 {
				log.Infof("line %v: Skipping unparsable metric item '%v' (must match regex '%v')", lineno, line, regexpMetricItem)
				continue
			}
			if matchMap["name"] != metricName {
				log.Infof("line %v: Skipping out-of-order line '%v' (expecting '%v ...')", lineno, line, metricName)
				continue
			}

			m, ok := c.GetMetric(metricName)
			if !ok {
				return nil, fmt.Errorf("line %v: Metric %q unknown", lineno, metricName)
			}

			value, err := strconv.ParseFloat(matchMap["value"], 64)
			if err != nil {
				log.Infof("line %v: Skipping metric item %q (cannot parse value %q)", lineno, line, matchMap["value"])
				continue
			}

			// correctness asserted in ScrapefileToCollection(...)
			f, _ := randomFunc(function)
			d, _ := randomDuration(interval)

			var min, max float64
			if isPercent(metricName, honorpct) {
				min = math.Max(0, value-float64(maxdeviation))
				max = math.Min(100, value+float64(maxdeviation))
			} else {
				min, max = randomRange(value, maxdeviation)
			}

			item := MetricItem{
				Min:      min,
				Max:      max,
				Func:     f,
				Interval: d,
			}

			if matchMap["labels"] != "" {
				labels := make(map[string]string)
				for _, labelKv := range strings.Split(matchMap["labels"], ",") {
					kv := strings.Split(labelKv, "=")
					labels[stripQuotes(kv[0])] = stripQuotes(kv[1])
				}
				item.Labels = labels
			}
			if err := m.AddItem(item); err != nil {
				return nil, fmt.Errorf("line %v: %v", lineno, err)
			}
		}
	}
	return &c, nil
}

func isPercent(metricName string, honorpct string) bool {
	s := strings.Split(honorpct, ",")
	for _, sub := range s {
		if strings.Contains(metricName, sub) {
			return true
		}
	}
	return false
}

func randomFunc(function string) (string, error) {
	funcSlice := strings.Split(function, ",")
	if len(funcSlice) < 1 {
		return "", fmt.Errorf("specify one or more functions")
	}
	for _, s := range funcSlice {
		if !isInSlice(s, validFunctions) {
			return "", fmt.Errorf("unknown function %q", s)
		}
	}
	return funcSlice[rand.Intn(len(funcSlice))], nil
}

func randomDuration(interval string) (time.Duration, error) {
	var result time.Duration
	times := strings.Split(interval, "-")
	if len(times) != 2 {
		return result, fmt.Errorf("interval must be specified as <duration>-<duration>")
	}
	from, err := time.ParseDuration(times[0])
	if err != nil {
		return result, err
	}
	if from < time.Duration(15*time.Second) {
		return result, fmt.Errorf("minimum duration %q too small. Must be >= 15s", times[0])
	}
	to, err := time.ParseDuration(times[1])
	if err != nil {
		return result, err
	}

	if from > to {
		return result, fmt.Errorf("minimum duration greater than maximum duration")
	}

	d := float64(from.Seconds()) + (float64(to.Seconds())-float64(from.Seconds()))*rand.Float64()
	result = time.Duration(int64(d)) * time.Second
	return result, nil
}

func randomRange(value float64, maxDeviation int) (min float64, max float64) {
	pct := rand.Float64() * float64(maxDeviation)
	min = value - (value * pct / 100)
	max = value + (value * pct / 100)
	return
}

func ScrapefileToCollection(filename string, maxdeviation int, function string, interval string, honorpct string) (*Collection, error) {

	// Assert correctness of input parameters
	_, err := randomFunc(function)
	if err != nil {
		return nil, err
	}
	_, err = randomDuration(interval)
	if err != nil {
		return nil, err
	}

	scrapeLines, err := readLines(filename)
	if err != nil {
		return nil, err
	}

	collection, err := convertScrapeToConfig(scrapeLines, maxdeviation, function, interval, honorpct)
	if err != nil {
		return nil, err
	}

	return collection, nil
}

// based on https://gist.github.com/lelandbatey/a5c957b537bed39d1d6fb202c3b8de06
// Unfinished idea to manipulate complex data types using reflection
/*
func setField(item interface{}, fieldName string, value interface{}) error {
	v := reflect.ValueOf(item).Elem()
	if !v.CanAddr() {
		return fmt.Errorf("cannot assign to the item passed, item must be a pointer in order to assign")
	}
	fieldNames := map[string]int{}
	for i := 0; i < v.NumField(); i++ {
		typeField := v.Type().Field(i)
		fieldNames[typeField.Name] = i
	}

	fieldNum, ok := fieldNames[fieldName]
	if !ok {
		return fmt.Errorf("field %s does not exist within the provided item", fieldName)
	}
	fieldVal := v.Field(fieldNum)
	fieldVal.Set(reflect.ValueOf(value))
	return nil
}
*/
