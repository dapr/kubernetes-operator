# Dapr Kubernetes Operator

This project was created by [@lburgazzoli in this repository](https://github.com/lburgazzoli/dapr-operator) and donated to the Dapr Sandbox organization. 

## Setup
 
```shell
# install the catalog
make openshift/deploy/catalog
```

## installation via cli

```shell
# waith for the catalog to be installed,
# then install the subsription
make openshift/deploy/subscritpion

# wait thil the subscription is ready,
# then deploy a dapr instance
make openshift/deploy/dapr
```

- Cleanup:
```shell
# cleanup
make openshift/undeploy:
```

## installation via UI

![image](https://github.com/lburgazzoli/dapr-operator/assets/1868933/9fc376a0-aec1-4bae-861f-361ccd9952aa)
![image](https://github.com/lburgazzoli/dapr-operator/assets/1868933/97fc8672-1f0c-4c1b-bd39-59f3c72287f2)
![image](https://github.com/lburgazzoli/dapr-operator/assets/1868933/faab9ee5-23b5-469d-8fd5-7d1f8aee34d7)
![image](https://github.com/lburgazzoli/dapr-operator/assets/1868933/19168795-817f-420f-95e5-b3523e2c4b2b)
![image](https://github.com/lburgazzoli/dapr-operator/assets/1868933/d76d9e55-86a1-4d22-857c-28550660d3fd)
![image](https://github.com/lburgazzoli/dapr-operator/assets/1868933/0379f506-1a52-4cad-ace7-c14c241af76f)
![image](https://github.com/lburgazzoli/dapr-operator/assets/1868933/c14a3022-cdc3-4469-b668-5afeb8cbfb8f)
![image](https://github.com/lburgazzoli/dapr-operator/assets/1868933/c26fec46-182e-4eee-8f23-208379ac9afe)
![image](https://github.com/lburgazzoli/dapr-operator/assets/1868933/ada9f1bb-6055-44f4-bac8-a5a83dc50689)

