# About

[![Build status](https://img.shields.io/github/workflow/status/rgl/kubernetes-hello/main)](https://github.com/rgl/kubernetes-hello/actions?query=workflow%3Amain)
[![Docker pulls](https://img.shields.io/docker/pulls/ruilopes/kubernetes-hello)](https://hub.docker.com/repository/docker/ruilopes/kubernetes-hello)

Container that shows details about the environment its running on.

This is used in:

* [rgl/rancher-single-node-ubuntu-vagrant](https://github.com/rgl/rancher-single-node-ubuntu-vagrant)

# Usage

Install docker.

Create the local test infrastructure:

```bash
./.github/workflows/kind/create.sh
```

Build and test:

```bash
./build.sh && ./test.sh && xdg-open index.html
```

Destroy the local test infrastructure:

```bash
./.github/workflows/kind/destroy.sh
```
