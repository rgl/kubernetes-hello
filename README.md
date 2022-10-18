# About

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
