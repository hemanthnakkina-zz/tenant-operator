## Tenant Controller

This is a CompositeController that defines namespaces and resourcequotas
connected to a Custom Resource Tenant.

### Prerequisites

* Install [Metacontroller](https://github.com/GoogleCloudPlatform/metacontroller)
* Install [ArgoWorkflow](https://github.com/argoproj/argo)

### Create a serviceaccount with RBAC

In this example, cluster-admin role is binded to the serviceaccount.
In production, create a role binding with the RBAC rules required.
In this PoC, serviceaccount requires complete RBAC permissions on namespaces, resourcequotas resourse.

```sh
kubectl -n argo create serviceaccount tenant-admin
kubectl create clusterrolebinding tenant-admin --clusterrole=cluster-admin --serviceaccount=argo:tenant-admin
```

### Deploy the controller

```sh
kubectl create configmap tenant-controller -n metacontroller --from-file=sync.py
kubectl create configmap tenant-argo-template -n metacontroller --from-file=template.j2
kubectl apply -f crd.yaml
kubectl apply -f controller.yaml
kubectl apply -f webhook.yaml
```

### Create a Tenant

```sh
kubectl apply -f tenant.yaml
```

Verify creation of namespace, resourcequota

kubectl get ns
kubectl get resourcequotas --all-namesapces
kubectl get workflows --all-namespaces
kubectl get tenants

### Clean up

```sh
kubectl delete -f webhook.yaml
kubectl delete -f controller.yaml
kubectl delete -f crd.yaml
kubectl delete configmap -n metacontroller tenant-controller
kubectl delete configmap -n metacontroller tenant-argo-template
```
