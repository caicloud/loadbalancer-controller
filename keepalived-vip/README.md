# keepalived-vip
A side car deplyed with nginx-ingress-controller for HA

# How to use

* First, create a Ingress ReplicationController (two replicas, uisng host network) with the keep-alived-vip side car:

```
apiVersion: v1
kind: ReplicationController
metadata:
  namespace: "kube-system"
  name: "ingress-test1"
  labels:
    run: "ingress"
spec:
  replicas: 2
  selector:
    run: "ingress"
  template:
   metadata:
    labels:
      run: "ingress"
   spec:
    hostNetwork: true
    containers:
      - image: "ingress-keepalived-vip:v0.0.1"
        imagePullPolicy: "Always"
        name: "keepalived-vip"
        resources:
          limits:
            cpu: 50m
            memory: 100Mi
          requests:
            cpu: 10m
            memory: 10Mi
        securityContext:
          privileged: true
        env:
          - name: POD_NAME
            valueFrom:
              fieldRef:
                fieldPath: metadata.name
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          - name: SERVICE_NAME
            value: "ingress"
      - image: "ingress-controller:v0.0.1"
        imagePullPolicy: "Always"
        name: "ingress-controller"
        resources:
          limits:
            cpu: 200m
            memory: 200Mi
          requests:
            cpu: 200m
            memory: 200Mi
```

* Then create a Service named "ingress" pointing to the Ingress ReplicaSet, and assign a vip to the service using annotation:

```
apiVersion: v1
kind: Service
metadata:
  name: ingress
  namespace: kube-system
  annotations:
    "ingress.alpha.k8s.io/ingress-vip": "192.168.10.1"
spec:
  type: ClusterIP
  selector:
    - run: "ingress"
  ports:
  - name: http
    port: 80
    protocol: TCP
    targetPort: 80
```

* keepalived-vip side car container will watch the "ingress" Service and update keepalived's config.


* User could access in-cluster service by the Ingress VIP.
