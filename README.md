### prom_rest_exporter

[![CircleCI](https://circleci.com/bb/mentalvary/prom_rest_exporter.svg?style=svg)](https://circleci.com/bb/mentalvary/prom_rest_exporter)

prom_rest_exporter translates arbitrary REST endpoints to metrics for [Prometheus](https://prometheus.io/).

It uses the excellent [jq](https://github.com/stedolan/jq) to transform JSON responses to numeric metric values.

prom_rest_exporter runs as a process exposing one or more /metrics endpoints for Prometheus.

## Development

Building prom_rest_exporter requires compiled
[jq](https://github.com/stedolan/jq) libraries.

### Fetching jq sources and build tools

```bash
git clone https://github.com/stedolan/jq.git jq-master
cd jq-master
git submodule update --init
```
Note: 1.6 release doesn't work because of bug when compiling

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

- documentation
- jq result freeing