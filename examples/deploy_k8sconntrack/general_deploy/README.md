## Deploy K8sconntrack on AWS Kubernetes Cluster

This example shows how to deploy k8sconntrack monitoring agent in an existing Kubernetes cluster. K8sconntrack is deployed as a DaemonSet in the cluster. So that it makes sure there is a k8sconntrack pod running on each node.

### Prerequisites

This example requires a running Kubernetes cluster. First check the current cluster status with kubectl.

```console
$ kubectl cluster-info
```

NOTE: This guide assumes that there is no authentication configured on kube-apiserver. If there is, please refer to the guide for deploying K8sconntrack on Kubernets AWS cluster.

### Create K8sconntrack DaemonSet

Here you need to get the IP address and port number for the kube-apiserver.

### Define a DaemonSet

```yaml
apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  name: k8snet
  labels:
    name: k8snet
spec:
  template:
    metadata:
      labels:
        name: k8snet
    spec:
      hostNetwork: true
      containers:
      - name: k8sconntracker
        image: dongyiyang/k8sconntracker:dev
        securityContext:
          privileged: true
        command:
          - /bin/conntracker
        args:
          - --v=2
          - --master=<Kubernetes_API_Server_Address>
      restartPolicy: Always
```

[Download example](k8sconntrack-ds.yaml?raw=true)

### Create a DaemontSet

```console
$ kubectl create -f k8sconntrack-ds.yaml
```

Then check the list of daemonsets, which include k8snet:

```console
$ kubectl get ds
NAME      DESIRED   CURRENT   NODE-SELECTOR   AGE
k8snet    4         4         <none>          19s
```

And the pods created by the DaemonSet:

```console
$ kubectl get po
NAME                                  READY     STATUS    RESTARTS   AGE
k8snet-1sttw                          1/1       Running   0          19s
k8snet-74pgt                          1/1       Running   0          19s
k8snet-hmpbj                          1/1       Running   0          19s
k8snet-pvezd                          1/1       Running   0          19s
```

Now K8sconntrack monitoring agent is running on each node in your Kubernetes cluster.
