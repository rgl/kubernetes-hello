#!/bin/bash
set -euo pipefail

CLUSTER_NAME='kubernetes-hello'
export KUBECONFIG="$PWD/kubeconfig.yml"
