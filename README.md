# Dapr Kubernetes Operator

The `Dapr Kubernetes Operator` manages the lifecycle for Dapr and its components. 
The operator's goal is to automate the tasks required when operating a Dapr running in [Kubernetes Mode](https://docs.dapr.io/operations/hosting/kubernetes/).

## Installation

The `Dapr Kubernetes Operator` was created with the intention of running through the [Operator Lifecycle Manager][olm_home], 
specifically on [OpenShift 4][openshift_home]. This is where the operator shines most, as it leverages the powerful 
features built into the latest version of OpenShift.

That being said, the operator can be installed and provide the same functionality on any Kubernetes cluster. The 
following methods are provided for installing the operator.

### OpenShift

The operator is published as part of the built-in Community Operators in the Operator Hub on OpenShift 4. See the 
[OpenShift Install Guide][install_openshift] for more information on installing on the OpenShift platform.

### Operator Lifecycle Manager

Using the Operator Lifecycle Manager to install and manage the `Dapr Kubernetes Operator` is the preferred method. The operator 
is published to [operatorhub.io][operatorhub_link]. Following the installation process there should work for most OLM 
installations.

Look at the [OLM Install Guide][install_olm] for an example using this approach with minikube. 

### Manual Installation

The operator can be installed manually if desired.

> [!IMPORTANT]
> The manual installation method requires cluster credentials that provide the `cluster-admin` ClusterRole or equivalent.

The [Manual Installation Guide][install_manual] provides the steps needed to manually install the operator on any 
Kubernetes cluster.

## Usage

### Basic

The following example shows the most minimal valid manifest to create a new Dapr instance:

```yaml
apiVersion: operator.dapr.io/v1beta1
kind: DaprInstance
metadata:
  name: "dapr-instance"
spec:
  values: {}
```

The `DaprInstance` resource is a Kubernetes Custom Resource (CRD) that describes the desired state for a given Dapr instance and allows for the configuration of the components that make up the instance.

When the `Dapr Kubernetes Operator` sees a new `DaprInstance` resource, the Dapr components are provisioned using Kubernetes resources generated from the official [Dapr Helm Charts](https://github.com/dapr/helm-charts) and managed by the operator.
This means that the same configuration option that are available when [installing Dapr using Helm](https://docs.dapr.io/operations/hosting/kubernetes/kubernetes-deploy/#install-with-helm-advanced) can also be used to configure the `Dapr Kubernetes Operator`
When something changes on an existing `DaprInstance` resource or any resource generated by the operator, the operator works to reconfigure the cluster to ensure the actual state of the cluster matches the desired state.

> [!IMPORTANT]
> The operator expect that a single cluster wide `DaprInstance` named `dapr-instance`.

The `DaprInstance` Custom Resource consists of the following properties

| Name   | Default | Description                                                      |
|--------|---------|------------------------------------------------------------------|
| values | [Empty] | The [values][helm_configuration] passed into the Dapr Helm chart |

[install_manual]:./docs/install/manual.md
[install_olm]:./docs/install/olm.md
[install_openshift]:./docs/install/openshift.md
[olm_home]:https://github.com/operator-framework/operator-lifecycle-manager
[openshift_home]:https://try.openshift.com
[operatorhub_link]:https://operatorhub.io/operator/dapr-kubernetes-operator
[helm_configuration]:https://github.com/dapr/dapr/blob/master/charts/dapr/README.md#configuration

### Create

Create a new Dapr Control Plane operator in the `openshift-operators` namespace using the provided basic example

```bash
kubectl apply -n openshift-operators -f config/samples/basic/dapr-basic.yaml
```

There will be several Dapr controllers and resources created that should be familiar to anyone who has deployed Dapr before using [helm](https://docs.dapr.io/operations/hosting/kubernetes/kubernetes-deploy/#install-with-helm-advanced):

```bash
➜ kubecto tree daprinstance.operator.dapr.io dapr-instance
NAMESPACE            NAME                                                     READY  REASON  AGE
openshift-operators  DaprInstance/dapr-instance                               True   Ready   72s
openshift-operators  ├─ConfigMap/dapr-trust-bundle                            -              68s
openshift-operators  ├─Configuration/daprsystem                               -              68s
openshift-operators  ├─Deployment/dapr-operator                               -              68s
openshift-operators  │ └─ReplicaSet/dapr-operator-5b98cf8c8                   -              68s
openshift-operators  │   └─Pod/dapr-operator-5b98cf8c8-cnt29                  True           67s
openshift-operators  ├─Deployment/dapr-sentry                                 -              68s
openshift-operators  │ └─ReplicaSet/dapr-sentry-74cbc77cb                     -              68s
openshift-operators  │   └─Pod/dapr-sentry-74cbc77cb-n7kb9                    True           67s
openshift-operators  ├─Deployment/dapr-sidecar-injector                       -              68s
openshift-operators  │ └─ReplicaSet/dapr-sidecar-injector-6745d677c7          -              68s
openshift-operators  │   └─Pod/dapr-sidecar-injector-6745d677c7-zbxv4         True           67s
openshift-operators  ├─Role/dapr-injector                                     -              67s
openshift-operators  ├─Role/dapr-operator                                     -              67s
openshift-operators  ├─Role/dapr-sentry                                       -              67s
openshift-operators  ├─Role/secret-reader                                     -              67s
openshift-operators  ├─RoleBinding/dapr-injector                              -              67s
openshift-operators  ├─RoleBinding/dapr-operator                              -              67s
openshift-operators  ├─RoleBinding/dapr-secret-reader                         -              67s
openshift-operators  ├─RoleBinding/dapr-sentry                                -              67s
openshift-operators  ├─Secret/dapr-trust-bundle                               -              67s
openshift-operators  ├─Service/dapr-api                                       -              67s
openshift-operators  │ └─EndpointSlice/dapr-api-mxw2p                         -              67s
openshift-operators  ├─Service/dapr-placement-server                          -              67s
openshift-operators  │ └─EndpointSlice/dapr-placement-server-74rkd            -              67s
openshift-operators  ├─Service/dapr-sentry                                    -              67s
openshift-operators  │ └─EndpointSlice/dapr-sentry-6c46f                      -              67s
openshift-operators  ├─Service/dapr-sidecar-injector                          -              67s
openshift-operators  │ └─EndpointSlice/dapr-sidecar-injector-8lt7j            -              67s
openshift-operators  ├─Service/dapr-webhook                                   -              67s
openshift-operators  │ └─EndpointSlice/dapr-webhook-sq5vm                     -              67s
openshift-operators  ├─ServiceAccount/dapr-injector                           -              67s
openshift-operators  ├─ServiceAccount/dapr-operator                           -              67s
openshift-operators  ├─ServiceAccount/dapr-placement                          -              67s
openshift-operators  ├─ServiceAccount/dapr-sentry                             -              67s
openshift-operators  └─StatefulSet/dapr-placement-server                      -              67s
openshift-operators    └─ControllerRevision/dapr-placement-server-6cb96b4b85  -              67s
```
