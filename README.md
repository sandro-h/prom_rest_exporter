### prom_rest_exporter

[![CircleCI](https://circleci.com/bb/mentalvary/prom_rest_exporter.svg?style=svg)](https://circleci.com/bb/mentalvary/prom_rest_exporter)

prom_rest_exporter translates arbitrary REST endpoints to metrics for [Prometheus](https://prometheus.io/).

It uses the excellent [jq](https://github.com/stedolan/jq) to transform JSON responses to numeric metric values.

prom_rest_exporter runs as a process exposing one or more `/metrics` endpoints for Prometheus.

## Installation

Download one of the released binaries.

Run the binary to start prom_rest_exporter.
It expects a configuration file, see below.

## Configuration

To configure what REST endpoints should be exported and how,
create a `prom_rest_exporter.yml` YAML file.

You can also name it differently and use `--config <path/to/config.yml>` to load it.

Simple config example:
```yaml
endpoints:
    # Run on 0.0.0.0:9011/metrics
  - port: 9011
    targets:
        # Get data from this REST endpoint
      - url: https://reqres.in/api/users
        metrics:
          - name: user_count
            description: Number of users
            type: gauge
            # jq program to extract a numeric value from the REST response
            selector: "[.data[].last_name] | length"
```

See [config.md](config.md) for more detailed information.

## Logging

prom_rest_exporter will write log output to `prom_rest_exporter.log` in the working directory.

Log details can be increased with the `--debug` and `--trace` command-line flags.

## Error handling

Failures during metric collection, such as REST endpoint downtimes or non-matching metric selectors, are logged but the collection continues for the remaining targets and metrics.

Any metrics that could not be extracted due to errors are simply skipped. Thus, Prometheus will show a "No data" gap when errors occur, and appropriate alerts can be set up for this.


## Development

Dependencies are managed with [dep](https://github.com/golang/dep).
Fetch all necessary dependencies with:
```bash
dep ensure
```

### jq dependency

Building prom_rest_exporter requires compiled
[jq](https://github.com/stedolan/jq) libraries.

```bash
git clone https://github.com/stedolan/jq.git jq-master
cd jq-master
git checkout jq-1.6
git submodule update --init
```

Required build tools:

```bash
sudo apt-get install autoconf make libtool flex bison gcc-mingw-w64-x86-64
```

#### Compiling for linux

```bash
autoreconf -fi
./configure --with-oniguruma=builtin --prefix=$PWD/build/linux/usr/local
make -j8
make install
# Remove so files to statically link
rm -f build/linux/usr/local/lib/*.so*
```

#### Cross-compiling for Windows

Cf. https://github.com/stedolan/jq/wiki/Cross-compilation

```bash
autoreconf -fi
./configure
make distclean
mkdir build
# Run it twice if first time you get "fatal error: compile.h"
CPPFLAGS=-I$PWD/src scripts/crosscompile win64 \
--disable-shared \
--enable-static \
--enable-all-static \
--target=win64-x86_64 \
--host=x86_64-w64-mingw32 \
--with-oniguruma=builtin
```

## Todos

- meta metrics per target
  - response time
  - errors