## Deploy K8sconntrack on AWS Kubernetes Cluster

This example shows how to deploy k8sconntrack monitoring agent in Kubernetes cluster running on AWS. K8sconntrack is deployed as a DaemonSet in the cluster. So that it makes sure on every node there is a k8sconntrack pod running.

### Prerequisites

This example requires a running Kubernetes cluster. First check the current cluster status with kubectl.

```console
$ kubectl cluster-info
```

Also, make sure you have all the cluster authentication files ready, including certificate authority file, client certificate and client key.

### Step One: Create Kubeconfig

#### Define Environment Variables

```console
$ export KUBETURBO_CONFIG_PATH=path/to/kubeturbo/config
```

If not specified, the default path is /etc/kubeturbo/

```console
$ export CLUSTER_NAME=<your Kubernetes cluster name>
```

If not specified, will use "kubernetes-aws"

```console
$ export USER_NAME=<username>
```

If not specified, will use "kube-aws-user"

```console
$ export CONTEXT_NAME=<your context name>
```

If not specified, will use "kube-aws-context"

#### Get Create_kubeconfig.sh

You can find the script under aws_deploy or get it from [here](create_kubeconfig.sh?raw=true)

### Step Two: Create Secret
Kubeconfig is mounted as a secret into each k8sconntrack pod. To create vmt-config secret, you can run:

```console
$ kubectl create secret generic vmt-config --from-file=path/to/kubeturbo/config
```

Then list your secrets:

```console
$ kubectl get secrets
NAME		TYPE		DATA		AGE
vmt-config	Opaque		3		5s
```

### Step Three: Create K8sconntrack DaemonSet

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
        ports:
          - name: http
            containerPort: 2222
            hostPort: 2222
        command:
          - /bin/conntracker
        args:
          - --v=3
          - --kubeconfig=/etc/kubeturbo/kubeconfig
        volumeMounts:
        - name: vmt-config
          mountPath: /etc/kubeturbo
          readOnly: true
      restartPolicy: Always
      volumes:
      - name: vmt-config
        secret:
          secretName: vmt-config
```

[Download example](k8sconntrack-use-secret-ds.yaml?raw=true)

### Create a DaemontSet

```console
$ kubectl create -f k8sconntrack-use-secret-ds.yaml
```

Then check the list of daemonsets, which include k8snet:

```console
$ kubectl get ds
NAME      DESIRED   CURRENT   NODE-SELECTOR   AGE
k8snet    4         4         <none>          19h
```

And the pods created by the DaemonSet:

```console
$ kubectl get po
NAME                                  READY     STATUS    RESTARTS   AGE
k8snet-22imb                          1/1       Running   0          19h
k8snet-mkk8x                          1/1       Running   0          19h
k8snet-thr05                          1/1       Running   0          19h
k8snet-v670d                          1/1       Running   0          19h
```

Now K8sconntrack monitoring agent is running on each node in your Kubernetes cluster.


