# OLM Install

The following steps can be used to install the operator using the [Operator Lifecycle Manager](https://github.com/operator-framework/operator-lifecycle-manager) on any Kubernetes 
environment.

## Cluster Setup

This guide uses [minikube](https://minikube.sigs.k8s.io/) to deploy a Kubernetes cluster locally, follow the 
instructions for your platform to install. If you already have a Kubernetes cluster ready to go, skip to 
the [OLM](#operator-lifecycle-manager) section.

Run minikube with a dedicated profile. Adjust the system resources as needed for your platform. 

```bash
minikube start -p dapr
```

## Operator Lifecycle Manager

Install the OLM components manually. If you already have OLM installed, skip to the [Operator](#operator-install) section.

Either

- install OLM from here: https://github.com/operator-framework/operator-lifecycle-manager/releases

or

- install using the `operator-sdk` command
```bash
operator-sdk olm install
```

Verify that OLM is installed. There should be two new namespaces, `olm` and `operators` created as a result.

```bash
kubectl get ns
```

```
NAME              STATUS   AGE
kube-system       Active   7d1h
default           Active   7d1h
kube-public       Active   7d1h
kube-node-lease   Active   7d1h
operators         Active   94s
olm               Active   94s
```

Verify that the OLM Pods are running in the `olm` namespace.

```bash
kubectl get pods -n olm
```

```
NAME                                READY   STATUS    RESTARTS   AGE
catalog-operator-569cd6998d-h5cbp   1/1     Running   0          39s
olm-operator-6fbbcd8c8b-qzv47       1/1     Running   0          39s
operatorhubio-catalog-m7qxq         1/1     Running   0          31s
packageserver-6cb8b48df4-wp89m      1/1     Running   0          30s
packageserver-6cb8b48df4-ww62h      1/1     Running   0          30s
```

That's it, OLM should be installed and availble to manage the Dapr Operator.

## Operator Install

The [dapr-kubernetes-operator](https://github.com/dapr/kubernetes-operator/) provides a pre-made kustomization file to deploy the Dapr Kubernetes Operator with OLM:

```bash
kubectl apply -k https://github.com/dapr/kubernetes-operator//config/samples/olm
```

This command should:

- Create a `dapr-system` namespace
- Create a `CatalogSource` in the `olm` namespace
  ```bash
  ➜ kubectl get catalogsources -n olm
  NAME                    DISPLAY               TYPE   PUBLISHER        AGE
  daprio-catalog          dapr.io catalog       grpc   dapr.io          11m
  operatorhubio-catalog   Community Operators   grpc   OperatorHub.io   18m
  ```
- Create an `OperatorGroup` in the `dapr-system` namespace
  ```bash
  ➜ kubectl get operatorgroups -n dapr-system
  NAME            AGE
  dapr-operator   12m
  ```
- Create a new `Subscription` for theDapr Kubernetes Operator in the new `dapr-system` namespace.
  ```bash
  ➜ kubectl get subscriptions.operators.coreos.com -n dapr-system
  NAME                 PACKAGE                    SOURCE           CHANNEL
  dapr-control-plane   dapr-kubernetes-operator   daprio-catalog   alpha
  ```
  The subscription should result in an `InstallPlan` being created in the `dapr-system` namespace which finally result in the `dapr-control-plane` Pod running
  ```bash
  ➜ kubectl get pods -n dapr-system
  NAME                                  READY   STATUS    RESTARTS   AGE
  dapr-control-plane-66866765b9-nzb6t   1/1     Running   0          13m
  ```

## Usage 

Once the operator is installed and running, new `DaprControlPlane` resources can be created. 

## Cleanup 

You can clean up the operator resources by running the following commands.

```bash
kubectl delete -k https://github.com/dapr/kubernetes-operator//config/samples/olm
```
