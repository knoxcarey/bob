# Simple Beacon-of-Beacons (BoB)

This is a simple implementation of a beacon-of-beacons: a service that
exposes the GA4GH beacon API and forwards the request on to a list of
other beacon systems. This implementation is intended to serve as a
test bench for experimenting with new security and federation
features. This is a work-in-progress, so caveat emptor.


## Running

This system is implemented in the Go programming language. Install
into your `$GOPATH` and build in the standard way.
If you run the program with the `-h` switch,
you'll get some information about command-line parameters that can be
set.

To issue a query against the BoB, use a command-line tool like `curl`
for now (until there is a visual interface):

```
curl "http://localhost:8080/?chromosome=1&start=156105028&alternateBases=C&referenceBases=T&assemblyId=GRCh37"
```

should return

```
[{"name":"ICGC","status":200,"responses":{"ICGC":"null"}},{"name":"Cosmic","status":200,"responses":{"Cosmic":"null"}}]
```

and

```
curl "http://localhost:8080/?chromosome=13&start=32900706&alternateBases=T&assemblyId=GRCh37"
```

should return

```
[{"name":"ICGC","status":200,"responses":{"ICGC":"true"}},{"name":"Cosmic","status":200,"responses":{"Cosmic":"true"}}]
```


## Configuring
The configuration files are (by default) in the `config` subdirectory. 

Configuration for each beacon is through a JSON document in the
`config/beacons` subdirectory. There are a couple of examples there to
get started. The `name`, `version`, and `endpoint` fields are required
for each beacon. The `datasetIds` field contains a list of data sets
to be queried for the beacon. The field `queryMap` is used to map the
standard names of query fields (and assemblies) to
implementation-specific names. For example, the COSMIC beacon uses the
string `chrom` instead of the standard `chromosome`. On the other
hand, the ICGC beacon is completely consistent with the standard API,
so no adjustments are necessary. Note that these mapping functions may
disappear in future versions as the standard APIs are more uniformly
observed.


## Roadmap

There are two immediate next steps with this BoB system: (a) add a
web-based user interface that allows the queries to be entered through
a web page and (b) integrate OpenID Connect/OAUTH 2.0 for
authentication and authorization. 