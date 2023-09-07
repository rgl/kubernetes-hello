# About

[![Build status](https://img.shields.io/github/actions/workflow/status/rgl/kubernetes-hello/main.yml?branch=master)](https://github.com/rgl/kubernetes-hello/actions/workflows/main.yml)
[![Docker pulls](https://img.shields.io/docker/pulls/ruilopes/kubernetes-hello)](https://hub.docker.com/repository/docker/ruilopes/kubernetes-hello)

Container that shows details about the environment its running on.

When running in Azure Kubernetes Service (AKS), it will also:

* List the Azure DNS Zones using the [Azure Workload Identity authentication](https://azure.github.io/azure-workload-identity/docs/) (see the [rgl/terraform-azure-aks-example repository](https://github.com/rgl/terraform-azure-aks-example)).

This is used in:

* [rgl/terraform-azure-aks-example](https://github.com/rgl/terraform-azure-aks-example)
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
