# Prometheus Exporter Simulator

This project implements a configurable prometheus exporter which can generate arbitrary metrics. It can be used to test prometheus scraping in scenarios where the real metrics are difficult to acquire.

Acknowledgements go to <https://github.com/webdevops/simulation-exporter> for inspiration and idea. Some lines of the original code are sure to be found here and I'm grateful for them. However the feature set is extended to a degree where I consider it a different project rather than a fork.

## Usage

The simulator reads metric definitions from a yaml configuration file and serves them as a scrapable prometheus page.

The configuration file can be written by hand, which is appropriate for simple setups. It can also be converted from real prometheus scrape output which allows to quickly setup a mock simulation of any arbitrary metric source.

As usual, the code can be invoked in several ways:

### Go-Run Locally

```sh
$ git clone <repo>
$ cd <repo>
$ go run . version
0.0.0
```

### Build and run locally

```sh
$ git clone <repo>
$ cd <repo>
$ make build
$ build/sim-exporter version
0.0.0
```

### Use GEC image, run locally

```sh
$ docker run --rm artifactory.intern.gec.io/docker-release-local/sim-exporter version
0.0.0
```

To work with local files you can optionally bind-mount your directory into the container:

```sh
$ cat myconf.yaml
version: "1"
metrics:
  population:
    name: population
    help: This is it
    type: gauge
    labels:
    - planet
    items:
    - min: 6.750e+09
      max: 8.086e+09
      func: rand
      interval: 1h
      labels:
        planet: earth
    - min: 0
      max: 0
      func: rand
      interval: 1h
      labels:
        planet: mars
$ docker run --rm -v $(pwd):/foo artifactory.intern.gec.io/docker-release-local/sim-exporter check /foo/myconf.yaml
/foo/myconf.yaml validated successfully
```

### Run on k8s

First, configure your local kubectl and helm to connect to your destination cluster.

Once done:

```sh
$ git clone <repo>
$ cd <repo>/deployment/chart
# Potentially review and change values.yaml
$ helm upgrade --install my-simulator-release .
```

## Commands

The simulator features the following commands. All commands have a (hopefully) useful help feature (try -h).

### convert

Parses a prometheus scrape output and converts it to a configuration yaml. This enables to quickly setup a simulator for metrics which were scraped from any arbitrary prometheus endpoint at any point in time.

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

Every metric is introduced by a HELP/TYPE header followed by one or more lines which are prefixed with that metric name. Each line is referred to as a "metric item". They have the same name but differ by their set of labels (the key/value pairs enclosed in "{}"). That is, two metric items must never have identical label sets. The line ends with the value of the metric item.

The convert command will turn this into a simulator configuration.

```sh
$ sim-exporter convert -o scrape.yaml scrape.txt
Wrote config to scrape.yaml
$ cat scrape.yaml
version: "1"
metrics:
- name: my_metric
  help: This metric shows awesome values
  type: counter
  labels:
  - flavor
  - instance_name
  items:
  - min: 104.30887825564639
    max: 141.6911217443536
    func: rand
    interval: 1h4m23s
    labels:
      flavor: m1.medium
      instance_name: server1
  - min: 257.31808355013527
    max: 654.6819164498647
    func: asc
    interval: 18m55s
    labels:
      flavor: m1.large
      instance_name: server2
- name: population
  help: This metric shows even more awesome values
  type: gauge
  labels:
  - planet
  items:
  - min: 4.325312545201143e+09
    max: 1.0510687454798857e+10
    func: desc
    interval: 28m0s
    labels:
      planet: earth
  - min: 0
    max: 0
    func: desc
    interval: 1h58m47s
    labels:
      planet: mars
```

Upon conversion, the deviation is calculated as a random percentage towards the scrape value with a maximum of `maxdeviation`. E.g. if the value in the scrape is 10 and maxdeviation is set to 50, then the converted min and max values might be anything from 10-10 (deviation 0) to 5-15 (deviation 50). Deviation is always applied symmetrically towards the scrape value.

The conversion process has basically no idea about the nature of the values it deviates. This can lead to undesired results, especially for percentage values which might be converted to less than 0 or more than 100. To work around this problem, a list of substrings can be specified for metric names to identify them as percentages. Percentage values will

- ... have the deviation applied as an absolute random value instead of a random percentage
- ... never be converted to less than zero or more than 100

Sure this approach cannot solve all cases but at least it solves mine ;).

In case you want to fine-tune the simulation you can of course manually change the converted file and specify values, intervals and functions that make most sense to you.

### check

Validate a configuration yaml. If validation succeeds then it should be safely usable as a simulator input.

```sh
$ sim-exporter check scrape.yaml
scrape.yaml validated successfully
```

### serve

Serve metrics from a configuration yaml as scrapable prometheus metrics on the specified port and path. The values will be mutated according to their min and max values by the configured function and repeating in the specified interval. New values will be calculated in the specified refresh interval.

```sh
$ sim-exporter serve --port 1234 --path /showme --refresh 10s scrape.yaml &
INFO[0000] Serving metrics on *:1234/showme
$ curl localhost:1234/showme
... (lots of internal prometheus/golang stuff) ...
# HELP my_metric This metric shows awesome values
# TYPE my_metric counter
my_metric{flavor="m1.large",instance_name="server2"} 415.9462395366754
my_metric{flavor="m1.medium",instance_name="server1"} 115.7983799012793
# HELP population This metric shows even more awesome values
# TYPE population gauge
population{planet="earth"} 7.418002421701642e+09
population{planet="mars"} 0
$ sleep 10
$ curl localhost:1234/showme
... (lots of internal prometheus/golang stuff again) ...
# HELP my_metric This metric shows awesome values
# TYPE my_metric counter
my_metric{flavor="m1.large",instance_name="server2"} 1267.1453595383523
my_metric{flavor="m1.medium",instance_name="server1"} 365.813331199372
# HELP population This metric shows even more awesome values
# TYPE population gauge
population{planet="earth"} 8.623539364871435e+09
population{planet="mars"} 0

```

## Code

The simulator configuration is represented by a `Collection`. It consists of a list of `Metric` objects.

The metric object mainly carries the name and type of the metric. The type is a prometheus vector type like `counter` or `gauge`. Secondly, the metric contains a list of `MetricItem`s.

Each metric item contains the parameters required for the value calculation which happens over time. The main factors are the `min` and `max` values as well as `func`tion and `interval`.

Collections can be created by either marshaling them from a yaml file or by creating them programmatically.

## Functions

Each metric item has a configured function and interval. They are used to allow for a deterministic way to mutate values over time (as apposed to random mutation). New values for all metrics are calculated on every refresh (see `serve` command). The values change according to the function stretched over the interval.

Implemented functions are:

### rand

Randomly changes the value between `min` and `max`

### asc

Starts the interval at `min` and linearly increases until `max` at the end of an interval.

### desc

Just the opposite, starts the interval at `max` and linearly decreases until `min` at the end of an interval.

### sin

Starts the interval in the middle (`(min+max)/2`) and does a full sine wave with the amplitude of `max-min` stretched over the interval.

## Helm Chart

The project contains a simple helm chart which makes it easy to drop the simulator into a kubernetes >=1.19 environment. Multiple configuration files can be mounted as a ConfigMap. Supply your own input by changing `.Values.configFiles`. One of the configurations is then chosen with `.Values.activeConfig` and served over `http://*:8080/metrics>`.

The chart can optionally create an ingress in case you need to make the simulator reachable from outside the prometheus cluster. However this is a poorly tested path which is not deemed excessively relevant.

## TODO: Notes to Self

- Functions should be factored out so they become easier to extend (package/interface `mutator`?)
