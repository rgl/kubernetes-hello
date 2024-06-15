# About

[![Build status](https://img.shields.io/github/actions/workflow/status/rgl/kubernetes-hello/main.yml?branch=master)](https://github.com/rgl/kubernetes-hello/actions/workflows/main.yml)
[![Docker pulls](https://img.shields.io/docker/pulls/ruilopes/kubernetes-hello)](https://hub.docker.com/repository/docker/ruilopes/kubernetes-hello)

Container that shows details about the environment its running on.

It will:

* Show the request method, url, and headers.
* Show the client and server address.
* Show the container environment variables.
* Show the container tokens, secrets, and configs (config maps).
* Show the container pod name and namespace.
* Show the containers running inside the container pod.
* Show the container memory limits.
* Show the container cgroups.
* Configure [the Go runtime](https://pkg.go.dev/runtime) to honour the defined kubernetes container memory and cpu resource limits with the `GOMEMLIMIT` and `GOMAXPROCS` environment variables.
* Expose as a Kubernetes `LoadBalancer` `Service`.
  * Note that this results in the creation of an [EC2 Classic Load Balancer (CLB)](https://docs.aws.amazon.com/elasticloadbalancing/latest/classic/introduction.html).
* Use [Role and RoleBinding](https://kubernetes.io/docs/reference/access-authn-authz/rbac/).
* Use [ConfigMap](https://kubernetes.io/docs/concepts/configuration/configmap/).
* Use [Secret](https://kubernetes.io/docs/concepts/configuration/secret/).
* Use [ServiceAccount](https://kubernetes.io/docs/concepts/security/service-accounts/).
* Use [Service Account token volume projection (a JSON Web Token and OpenID Connect (OIDC) ID Token)](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#serviceaccount-token-volume-projection) for the `https://example.com` audience.

When running in Azure Kubernetes Service (AKS), it will also:

* List the DNS Zones hosted in [Azure DNS](https://azure.microsoft.com/en-us/products/dns) using the [Azure Workload Identity authentication](https://azure.github.io/azure-workload-identity/docs/) (see the [rgl/terraform-azure-aks-example repository](https://github.com/rgl/terraform-azure-aks-example)).

When running in Amazon Elastic Kubernetes Service (EKS), it will also:

* List the DNS Zones hosted in [AWS Route 53](https://aws.amazon.com/route53/) using the [EKS Pod Identities authentication](https://docs.aws.amazon.com/eks/latest/userguide/pod-identities.html) (see the [rgl/terramate-aws-eks-example repository](https://github.com/rgl/terramate-aws-eks-example)).

This is used in:

* [rgl/terraform-azure-aks-example](https://github.com/rgl/terraform-azure-aks-example)
* [rgl/terramate-aws-eks-example](https://github.com/rgl/terramate-aws-eks-example)
* [rgl/rancher-single-node-ubuntu-vagrant](https://github.com/rgl/rancher-single-node-ubuntu-vagrant)

# Usage

Install docker, kind, and kubectl.

Create the local test infrastructure:

```bash
./.github/workflows/kind/create.sh
```

Access the test infrastructure kind Kubernetes cluster:

```bash
export KUBECONFIG="$PWD/kubeconfig.yml"
kubectl get nodes -o wide
```

Build and test:

```bash
./build.sh && ./test.sh && xdg-open index.html
```

Destroy the local test infrastructure:

```bash
./.github/workflows/kind/destroy.sh
```
