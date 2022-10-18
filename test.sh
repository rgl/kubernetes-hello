#!/bin/bash
SCRIPT_PATH="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
source "$SCRIPT_PATH/.github/workflows/kind/env.sh"

kubectl exec \
    --context "kind-$CLUSTER_NAME" \
    --quiet \
    --stdin \
    --tty \
    "$(
        kubectl get pods \
            --context "kind-$CLUSTER_NAME" \
            -l app=kubernetes-hello \
            -o name
    )" \
    -- \
    wget -qO- http://localhost:8000 \
    >index.html
