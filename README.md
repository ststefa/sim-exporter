# Prometheus Exporter Simulator

This project implements a configurable prometheus exporter which can generate arbitrary metrics. It can be used to test prometheus scraping in scenarios where the real metrics are difficult to acquire.

Acknowledgements go to <https://github.com/webdevops/simulation-exporter> for inspiration and idea.

## Usage

The simulator reads metric definitions from a yaml configuration file and serves them as a scrapable prometheus page. The metric values are modified in intervals.

The simulator features the following commands

### convert

Parses a prometheus scrape output and converts it to a configuration yaml. This enables to quickly setup a simulator for metrics which were scraped from any arbitrary endpoint at any point in time.

A prometheus scrape output looks roughly like so:

```sh
$ cat scrape.txt
# HELP my_metric This metric shows awesome values
# TYPE my_metric counter
my_metric{flavor="m1.medium",instance_name="server1"} 123
my_metric{flavor="m1.large",instance_name="server2"} 456
# HELP population This metric shows even more awesome values
# TYPE population gauge
population{planet="earth"} 7.418e+9
population{planet="mars"} 0
```

The convert command will turn this into a simulator configuration.

```sh
$ sim-exporter convert --maxdeviation 50 scrape.txt
version: "1"
metrics:
  my_metric:
    name: my_metric
    help: This metric shows awesome values
    type: counter
    labels:
    - flavor
    - instance_name
    items:
    - value: 103-143
      labels:
        flavor: m1.medium
        instance_name: server1
    - value: 274-638
      labels:
        flavor: m1.large
        instance_name: server2
  population:
    name: population
    help: This metric shows even more awesome values
    type: gauge
    labels:
    - planet
    items:
    - value: 6.750e+09-8.086e+09
      labels:
        planet: earth
    - value: "0"
      labels:
        planet: mars
```

The values will be randomly chosen based ob the supplied parameters. They can be either constant (e.g. 123) or a range (e.g. 12-34). Ranged values change over time in the simulator. Constants do not.

### check

Validate a configuration yaml. If validation succeeds then it should be safely usable as a simulator input. Except for the cases where it's not ;).

### serve

Serve metrics from a configuration yaml as scrapable prometheus metrics on the specified port and path. The values will be mutated according to their range in the specified refresh interval.

```sh
$ sim-exporter serve --port 1234 --path /showme --refresh 60s scrape.yaml
INFO[0000] Serving metrics on *:1234/showme
```

## Helm Chart

The project contains a simple helm chart which makes it easy to drop the simulator into a kubernetes >=1.19 environment. The configuration file is mounted as a ConfigMap. Supply your own input by changing