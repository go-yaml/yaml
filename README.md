# YAML support for the Go language

This is a fork of `github.com/go-yaml/yaml` to provide modified YAML 
parsing for use with libcalico-go and calicoctl.  The modifications include:
  -  Swapping support of Float32 with Float64 (since calico does not use Float32)

