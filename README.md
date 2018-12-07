### prom_rest_exporter

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