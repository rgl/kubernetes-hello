#!/bin/bash
SCRIPT_PATH="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
source "$SCRIPT_PATH/.github/workflows/kind/env.sh"

echo 'Building the container image...'
DOCKER_BUILDKIT=1 docker build -t localhost:5001/kubernetes-hello .

echo 'Pushing the container image...'
docker push localhost:5001/kubernetes-hello

echo 'Listing the remote container image tags...'
#wget -qO- http://localhost:5001/v2/_catalog | jq
wget -qO- http://localhost:5001/v2/kubernetes-hello/tags/list | jq

# delete the existing pod.
kubectl get pods --context "kind-$CLUSTER_NAME" -l app=kubernetes-hello -o name | while read pod_name; do
    echo "Deleting existing pod $pod_name..."
    kubectl delete --context "kind-$CLUSTER_NAME" "$pod_name"
done

echo 'Deploying the application...'
sed -E 's,(\simage:).+,\1 localhost:5001/kubernetes-hello:latest,g' \
    resources.yml \
    | kubectl apply \
        --context "kind-$CLUSTER_NAME" \
        -f -

echo 'Waiting for the application to be running...'
kubectl rollout status \
    --context "kind-$CLUSTER_NAME" \
    --timeout 3m \
    daemonset/kubernetes-hello
