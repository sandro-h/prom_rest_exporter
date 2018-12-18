### prom_rest_exporter

[![CircleCI](https://circleci.com/bb/mentalvary/prom_rest_exporter.svg?style=svg)](https://circleci.com/bb/mentalvary/prom_rest_exporter)

## intro

goal is a tool that translate arbitrary application rest endpoints to a /metrics endpoint
understandable by prometheus.

learning goal is:

* (re)learn golang
* make early use of CI

## jq in golang

jq c api:  
https://github.com/stedolan/jq/wiki/C-API:-libjq

existing golang bindings:  
https://github.com/ashb/jqrepl

c bindings in golang:  
https://golang.org/cmd/cgo/

### Compiling jq

```
git clone https://github.com/stedolan/jq.git jq-master
cd jq-master
git submodule update --init
```
Note: 1.6 release doesn't work because of bug when compiling

packages:

```
sudo apt-get install autoconf make libtool flex bison gcc-mingw-w64-x86-64
```

#### compile for linux

```
autoreconf -fi
./configure --with-oniguruma=builtin --prefix=$PWD/build/linux/usr/local
make -j8
make install
# Remove so files to statically link
rm -f build/linux/usr/local/lib/*.so*
```

#### cross-compile for windows

https://github.com/stedolan/jq/wiki/Cross-compilation

```
autoreconf -fi
./configure
make distclean
# Run it twice if first time you get fatal error: compile.h
CPPFLAGS=-I$PWD/src scripts/crosscompile win64 --disable-shared --enable-static --enable-all-static --target=win64-x86_64 --host=x86_64-w64-mingw32 --with-oniguruma=builtin
```

## prometheus metrics

https://github.com/prometheus/docs/blob/master/content/docs/instrumenting/exposition_formats.md

## todos
