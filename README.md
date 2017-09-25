<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [LoadBalancer Controller](#loadbalancer-controller)
  - [About the project](#about-the-project)
    - [Status](#status)
    - [Design](#design)
    - [See also](#see-also)
  - [Getting started](#getting-started)
    - [Layout](#layout)
  - [TODO](#todo)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# LoadBalancer Controller 

## About the project

A LoadBalancer, containing a `proxy` and multiple `providers`,  provides external traffic load balancing for kubernetes applications.

A `proxy` is an ingress controller watching ingress resources to provide accesses that allow inbound connections to reach the cluster services.

A `provider` is the entrance of the cluster providing high availability for connections to proxy (ingress controller).

### Status

**Working in process** 

This project is still in alpha version.

### Design

Learn more about loadbalancer on [design doc](./docs/design.md)

### See also

-   [loadbalancer provider](https://github.com/caicloud/loadbalancer-provider)

## Getting started

### Layout

```
├── cmd
│   └── controller
├── config
├── controller
├── docs
│   └── images
├── hack
│   └── license
├── pkg
│   ├── apis
│   │   └── networking
│   │       └── v1alpha1
│   ├── informers
│   │   ├── internalinterfaces
│   │   └── networking
│   │       └── v1alpha1
│   ├── listers
│   │   └── networking
│   │       └── v1alpha1
│   ├── toleration
│   ├── tprclient
│   │   └── networking
│   │       └── v1alpha1
│   └── util
│       ├── controller
│       ├── lb
│       ├── strings
│       ├── taints
│       └── validation
├── provider
│   └── providers
│       └── ipvsdr
├── proxy
│   └── proxies
│       └── nginx
└── version
```

A brief description:

-   `cmd` contains main packages, each subdirecoty of `cmd` is a main package.
-   `docs` for project documentations.
-   `hack` contains scripts used to manage this repository.
-   `pkg` contains apis, informers, listers, clients, util for LoadBalancer TPR.
-   `provider` contains provider plugins, each subdirectory is one kind of a provider
-   `proxy` contains proxy plugins, each subdirectory is one kind of a proxy
-   `version` is a placeholder which will be filled in at compile time

## TODO

-   readjust the directory structure
-   update api to v1alpha2
-   separate api from the project to clientset
-   auto generate clients and informers