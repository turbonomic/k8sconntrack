## Deploy K8sConntrack on AWS Kubernetes Cluster

This example shows how to deploy K8sConntrack monitoring agent in an existing Kubernetes cluster. K8sConntrack is deployed as a DaemonSet in the cluster. So that it makes sure there is a K8sConntrack pod running on each node.

### Prerequisites

This example requires a running Kubernetes cluster. First check the current cluster status with kubectl.

```console
$ kubectl cluster-info
```

NOTE: This guide assumes that there is no authentication configured on kube-apiserver. If there is, please refer to the guide for deploying K8sConntrack on Kubernets AWS cluster.

### Create K8sConntrack DaemonSet

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
      - name: K8sConntracker
        image: dongyiyang/K8sConntracker:dev
        securityContext:
          privileged: true
        command:
          - /bin/conntracker
        args:
          - --v=2
          - --master=<Kubernetes_API_Server_Address>
      restartPolicy: Always
```

[Download example](K8sConntrack-ds.yaml?raw=true)

### Create a DaemontSet

```console
$ kubectl create -f K8sConntrack-ds.yaml
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

Now K8sConntrack monitoring agent is running on each node in your Kubernetes cluster.
