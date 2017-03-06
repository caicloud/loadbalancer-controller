# loadbalancer-controller

#### Description
LoadbalabcerController is a controller which allow to dynamically provision a Loadbalancer.
For more information, see design doc [here](https://github.com/kubernetes/community/pull/275),
currently implemented by third party resource.


#### How to use
* First, create third party resources
```
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
```
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

* loadbalancer-controller will dynamically provision a loadbalancer for you, by deploy a nginx-ingress-controller
replication controller in the kube-system namespace


#### Current stage
* loadbalancer-provision-controller is completed, which can dynamically provision a Loadbalancer for a LoadbalancerClaim

#### Future plan
* loadbalancer-recycler-controller whcih recycle a Loadbalancer is it's LoadbalancerClaim is deleted
* loadbalancer-binding-controller which bind an Ingress resource to Loadbalancer