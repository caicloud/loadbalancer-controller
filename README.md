# Loadbalancer controller

## Description

LoadbalancerController is a controller which provisions Loadbalancer dynamically.
A loadbalancer is either on-prem solution (e.g. nginx, F5) or cloud Loadbalancer (e.g.
Google Cloud Load Balancing). Loadbalancer is backed by [ingress controller](https://github.com/kubernetes/ingress)
in kubernetes. For more information, see design doc [here](https://github.com/kubernetes/community/pull/275).
The feature is currently implemented via third party resource.

## Usage

* First, create third party resources `loadbalancer` and `loadbalancerclaim`:

```yaml
apiVersion: extensions/v1beta1
kind: ThirdPartyResourceList
items:
  - apiVersion: extensions/v1beta1
    kind: ThirdPartyResource
    metadata:
      name: loadbalancerclaim.k8s.io
    description: "Allow user to claim a loadbalancer instance"
    versions:
    - name: v1
  - apiVersion: extensions/v1beta1
    kind: ThirdPartyResource
    metadata:
      name: loadbalancer.k8s.io
    description: "Allow user to CURD a loadbalancer instance"
    versions:
    - name: v1
```

* Second, deploy loadbalancer-controller

```yaml
apiVersion: v1
kind: ReplicationController
metadata:
  namespace: "kube-system"
  name: "loadbalancer-controller"
  labels:
    run: "loadbalancer-controller"
spec:
  replicas: 1
  selector:
    run: "loadbalancer-controller"
  template:
   metadata:
    labels:
      run: "loadbalancer-controller"
   spec:
    containers:
      - image: "index.caicloud.io/caicloud/loadbalancer-controller:v0.0.1"
        imagePullPolicy: "Always"
        name: "loadbalancer-claim-controller"
        resources:
          limits:
            cpu: 100m
            memory: 100Mi
          requests:
            cpu: 100m
            memory: 100Mi
```

* Then, you can create a LoadbalancerClaim

```
apiVersion: k8s.io/v1
kind: Loadbalancerclaim
metadata:
  namespace: "kube-system"
  name: "loadbalancer-claim-test"
  annotations:
    ingress.alpha.k8s.io/provisioning-required: "required"
    ingress.alpha.k8s.io/ingress-class: "ingress.alpha.k8s.io/ingress-nginx"
    ingress.alpha.k8s.io/ingress-cpu: "50m"
    ingress.alpha.k8s.io/ingress-mem: "50Mi"
    ingress.alpha.k8s.io/ingress-vip: "127.192.0.1"
```

loadbalancer-controller will dynamically provision a loadbalancer for you, by deploying a
`nginx-ingress-controller` replication controller in the kube-system namespace.

#### Current stage

* loadbalancer-provision-controller is completed, which can dynamically provision a Loadbalancer for a LoadbalancerClaim

#### Future plan

* loadbalancer-recycler-controller which recycles a Loadbalancer when its LoadbalancerClaim is deleted
* loadbalancer-binding-controller which binds an Ingress resource to Loadbalancer
