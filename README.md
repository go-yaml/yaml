[![Build Status](https://semaphoreci.com/api/v1/calico/go-yaml-3/branches/calico/shields_badge.svg)](https://semaphoreci.com/calico/go-yaml-3)

# YAML support for the Go language

This is a fork of `github.com/go-yaml/yaml` to provide modified YAML 
parsing for use with libcalico-go and calicoctl.  The modifications include:
  -  Swapping support of Float32 with Float64 (since calico does not use Float32)

