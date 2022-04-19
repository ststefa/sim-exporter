# Prometheus Exporter Simulator

This project implements a configurable prometheus exporter which can generate arbitrary metrics. It can be used to test prometheus scraping in scenarios where the real metrics are difficult to acquire.

Acknowledgements go to <https://github.com/webdevops/simulation-exporter> for inspiration and idea.

## Usage

The simulator reads metric definitions from a yaml configuration file and serves them as a scrapable prometheus page. The metric values are modified in intervals.

The simulator features the following commands

### convert

Parses a prometheus scrape output and converts it to a configuration yaml. This enables to quickly setup a simulator for metrics which were scraped from any arbitrary endpoint at any point in time.

A prometheus scrape output looks roughly like so:

```text
# HELP a_metric This metric shows awesome values
# TYPE a_metric counter
a_metric{flavor="m1.medium",instance_name="server1"} 123
a_metric{flavor="m1.large",instance_name="server2"} 456
# HELP population This metric shows even more awesome values
# TYPE population gauge
population{planet="earth"} 7e+9
population{planet="mars"} 0
...
```

### check

Validate a configuration yaml. If validation succeeds then it should be safely usable. Except for the cases where it's not ;).

### serve

Serve metrics from a configuration yaml as scrapable prometheus metrics

## Helm Chart

for kubernetes >=1.19