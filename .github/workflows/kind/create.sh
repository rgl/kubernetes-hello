#!/bin/bash
SCRIPT_PATH="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
source "$SCRIPT_PATH/env.sh"

# see https://hub.docker.com/_/registry
# see https://github.com/distribution/distribution/releases
# renovate: datasource=docker depName=registry
registry_image_version='3.0.0'

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
    --volume "$SCRIPT_PATH/registry-config.yml:/etc/distribution/config.yml:ro" \
    --env OTEL_SDK_DISABLED=true \
    --env OTEL_TRACES_EXPORTER=none \
    --env OTEL_METRICS_EXPORTER=none \
    --env OTEL_LOGS_EXPORTER=none \
    -p 5001:5001 \
    "registry:$registry_image_version" \
    >/dev/null
while ! wget -q --spider http://localhost:5001/v2/; do sleep 1; done;

echo 'Connecting the docker registry to the kind k8s network...'
# TODO isolate the network from other kind clusters with KIND_EXPERIMENTAL_DOCKER_NETWORK.
#      see https://github.com/kubernetes-sigs/kind/blob/v0.31.0/pkg/cluster/internal/providers/docker/network.go
docker network connect \
    kind \
    "$CLUSTER_NAME-registry"
