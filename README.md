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
# TYPE my_metric gauge
my_metric{flavor="m1.medium",instance_name="server1"} 123
my_metric{flavor="m1.large",instance_name="server2"} 456
# HELP population This metric shows even more awesome values
# TYPE population gauge
population{planet="earth"} 7.418e+9
population{planet="mars"} 0
```

Every metric is introduced by a HELP/TYPE header followed by one or more lines which are prefixed with that metric name. Each line is referred to as a "metric item". They have the same name but differ by their set of labels (the key/value pairs enclosed in "{}"). That is, two metric items of a given metric must never have identical label sets. The line ends with the value of the metric item.

The convert command will turn this into a simulator configuration.

```sh
$ sim-exporter convert -o scrape.yaml scrape.txt
Wrote config to scrape.yaml
$ cat scrape.yaml
version: "1"
metrics:
- name: my_metric
  help: This metric shows awesome values
  type: gauge
  labels:
  - flavor
  - instance_name
  items:
  - min: 90.22476595603
    max: 155.77523404397
    func: sin
    interval: 1h47m5s
    labels:
      flavor: m1.medium
      instance_name: server1
  - min: 351.8987625539297
    max: 560.1012374460703
    func: rand
    interval: 45m10s
    labels:
      flavor: m1.large
      instance_name: server2
- name: population
  help: This metric shows even more awesome values
  type: gauge
  labels:
  - planet
  items:
  - min: 5.446888360363824e+09
    max: 9.389111639636177e+09
    func: asc
    interval: 1h19m24s
    labels:
      planet: earth
  - min: 0
    max: 0
    func: asc
    interval: 1h34m19s
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
# TYPE my_metric gauge
my_metric{flavor="m1.large",instance_name="server2"} 503.7733071579815
my_metric{flavor="m1.medium",instance_name="server1"} 123.00000220622213
# HELP population This metric shows even more awesome values
# TYPE population gauge
population{planet="earth"} 5.446888466456302e+09
population{planet="mars"} 0
$ sleep 10
$ curl localhost:1234/showme
... (lots of internal prometheus/golang stuff again) ...
# HELP my_metric This metric shows awesome values
# TYPE my_metric gauge
my_metric{flavor="m1.large",instance_name="server2"} 447.28264243993624
my_metric{flavor="m1.medium",instance_name="server1"} 123.64107424600228
# HELP population This metric shows even more awesome values
# TYPE population gauge
population{planet="earth"} 5.463440479214053e+09
population{planet="mars"} 0
```

## Code

The simulator configuration is represented by a `Collection`. It consists of a list of `Metric` objects.

The metric object mainly carries the name and type of the metric. The type is a prometheus vector type like `counter` or `gauge`. Secondly, the metric contains a list of `MetricItem` entries.

Each metric item contains the parameters required for the value calculation which happens over time. The main factors are the `min` and `max` values as well as `func`(tion) and `interval`.

Collections can be created by either unmarshaling them from a yaml file or by creating them programmatically.

## Functions

Each metric item has a configured function and interval. They are used to allow for a deterministic way to change values over time (as apposed to changing them randomly). New values for all metrics are calculated on every refresh (see `serve` command). The values change according to the function stretched over the interval.

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

The project contains a simple helm chart which makes it easy to drop the simulator into a kubernetes (aka k8s) >=1.19 environment. Multiple configuration files can be mounted as a k8s `ConfigMap`. Supply your own input by changing `.Values.configs`. One of the configurations is then chosen with `.Values.activeConfig` and served over `http://*:8080/metrics>` by default.

The chart can optionally create an ingress in case you need to make the simulator reachable from outside the prometheus cluster. However this is a poorly tested path which is not deemed excessively relevant.

In case your kubernetes has a Prometheus Operator installed, you can also enable automatic scrape configuration using a ServiceMonitor.

### Helm Deployment Example

The example assumes that you registered the GEC artifactory with the name "gec". Modify as appropriate. Also, make sure it is properly updated using `helm repo update` to get the latest version.

First, extract the values of the chart to a local file

```sh
$ helm show values gec/sim-exporter > myvalues.yaml
$
```

Next, edit `myvalues.yaml` to your preference. I recommend deleting everything that is not changed to keep it minimal. You will usually want to keep just `configs, activeConfig` and maybe `refreshTime`. The result might e.g. look like so:

```yaml
configs:
  mymetrics.yaml: |-
    version: v1
    metrics:
    - name: happy-wave
      type: gauge
      items:
      - min: 0
        max: 10
        func: sin
        interval: 3m
    - name: flippy-saw
      type: gauge
      items:
      - min: -10
        max: 20
        func: asc
        interval: 2m
activeConfig: mymetrics.yaml
refreshTime: 5s
```

...or, if your kubernetes has a Prometheus Operator, additionally:

```yaml
serviceMonitor:
  enabled: true
```

Make sure that your configs are valid. You might check that by creating them as a separate file first and validate them using the `check` command (see **Usage** above).

Deploy the chart with your values

```sh
$ helm upgrade --install mysim -f myvalues.yaml gec/sim-exporter
...
```

The chart produces useful output that shows how to access the exporter through the k8s `Service`. In case of using Prometheus Operator, your metrics should additionally be immediately scraped and visible.

## TODO: Notes to Self

- Functions should be factored out so they become easier to extend (package/interface `mutator`?)
- Additional func "rect": v=i.Min in first half of interval, i.Max in second half
- Additional func "saw": linearly increase until middle of interval, then linearly decrease
- It might be better to treat small values as an int. E.g. for an "up" value it makes no sense that it deviates after the decimal point. At the same time, there might be values for which it does make sense. How to decide without configuring explicitly? Perhaps regex a la "percent" rule?
