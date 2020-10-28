# 说明

**由于当前依赖的 Kubernetes 版本太低，无法使用上游定义的版本。**

apis 包都是拷贝自 https://github.com/Kong/kubernetes-ingress-controller 0.9.x 分支，虽然上游已将 internal 移到 pkg 目录，
但其依赖的 k8s 版本太高，与当前 admin 不兼容，故 client 包是本地通过 code-generator 生成后拷贝过来的，与当前项目 k8s 版本兼容且不再依赖高版本 k8s。

为保证编译通过，调整了 import 路径 kong/kubernetes-ingress-controller/internal/ ->  caicloud/loadbalancer-admin/pkg/kong
