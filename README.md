<img src="images/title.png" width="500px;"/>

Dynamically change workload affinity in Kubernetes. Useful for testing, development and reacting to environmental changes.

#### Installation

This will install the operator and custom resources

```
kubectl apply -f https://github.com/cloud-native-skunkworks/placement-operator/releases/download/latest/release.yaml
```

#### Run the demo

This will install `placement-application` with a unique `layout` that you can adjust the strategy of.

```
kubectl apply -f https://github.com/cloud-native-skunkworks/placement-operator/releases/download/latest/demo.yaml
```


### Design 

There are initial two modes of operation **balanced** and **stacked** within the Layout.

<img src="images/03.png" width="500px;"/>

#### Custom resource

The custom resource is called **Layout** and is defined as follows:
```
apiVersion: core.cnskunkworks.io/v1alpha1
kind: Layout
metadata:
  name: layout-sample
spec:
  # balanced | stacked 
  strategy: stacked
```

  To utilise this resource a workload must use `spec.template.metadata.labels` 

```
spec:
  template:
    metadata:
      labels:
        app: placed-application-demo
        cnskunkworks.io/placement-operator-enabled: "true"
        cnskunkworks.io/placement-operator-layout: layout-sample
```



Setting a strategy will rebalance, this can be both **balanced** or **stacked** this can be updated dynamically within the layout custom resource and will recreate pods.

<img src="images/01.png" width="550px;" />


### TODO

- [ ] Add custom DSL to create layout conditions
- [ ] Add additional default layout types
- [ ] Allow for graceful pod termination and restart conditions