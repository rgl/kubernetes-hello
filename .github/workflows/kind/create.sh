#!/bin/bash
SCRIPT_PATH="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
source "$SCRIPT_PATH/env.sh"

echo "Creating $CLUSTER_NAME k8s..."
kind create cluster \
    --name="$CLUSTER_NAME" \
    --config="$SCRIPT_PATH/config.yml"
kubectl cluster-info

echo 'Creating the docker registry...'
# TODO create the registry inside the k8s cluster.
docker run \
    -d \
    --restart=unless-stopped \
    --name "$CLUSTER_NAME-registry" \
    --env REGISTRY_HTTP_ADDR=0.0.0.0:5001 \
    -p 5001:5001 \
    registry:2.8.3 \
    >/dev/null
while ! wget -q --spider http://localhost:5001/v2; do sleep 1; done;

echo 'Connecting the docker registry to the kind k8s network...'
# TODO isolate the network from other kind clusters with KIND_EXPERIMENTAL_DOCKER_NETWORK.
#      see https://github.com/kubernetes-sigs/kind/blob/v0.21.0/pkg/cluster/internal/providers/docker/network.go
docker network connect \
    kind \
    "$CLUSTER_NAME-registry"
