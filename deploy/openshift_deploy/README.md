## Deploy K8sConntrack on OpenShift Cluster

This guide is about how to deploy Kubeturbo service in an existing OpenShift cluster.

### Prerequisites
This example requires a running OpenShift cluster. First check the current cluster status with kubectl.

```console
$ kubectl cluster-info
```

### Step One: Create ServiceAccount
A ServiceAccount is needed for Kubeturbo service to access the OpenShift cluster.

#### Define Turbo-user Service Account

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: turbo-user
  namespace: default
```

[Download example](turbo-user-service-account.yaml?raw=true)

Then you would see turbo-user when you list service accounts in OpenShift.

```console
$kubectl get sa --namespace=default
NAMESPACE          NAME                        SECRETS   AGE
default            builder                     2         62d
default            default                     2         62d
default            deployer                    2         62d
default            registry                    2         62d
default            router                      2         62d
default            turbo-user                  2         25s
```
### Step Two: Edit Security Context Constraint
In OpenShift, security context constraints allow administrator to control permissions for pods. As K8sConntrack pod needs privileged permissions.
you need to add turbo-user service account to proper security context constraints. Here turbo-user is added to *privileged* security context constraint.

```console
$oc edit scc privileged
```

Then add "system:serviceaccount:default:turbo-user" under users, as shown

```console
users:
- system:serviceaccount:openshift-infra:build-controller
- system:serviceaccount:management-infra:management-admin
- system:serviceaccount:management-infra:inspector-admin
- system:serviceaccount:default:router
- system:serviceaccount:default:registry
- system:serviceaccount:osproj1:turbo-user
- admin
- system
- root
```

### Step Three: Create vmt-config Secret

A vmt-config secret is required for K8sConntrack pods to interact with Kube-apiserver. So you need to put admin.kubeconfig into vmt-config secret.


```console
$ kubectl create secret generic vmt-config --from-file=path/to/admin.kubeconfig
```

Then list your secrets:

```console
$ kubectl get secrets
NAME		TYPE		DATA		AGE
vmt-config	Opaque		3		5s
```

### Step Four: Create K8sConntrack DaemonSet

```yaml
apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  name: k8snet
  namespace: default
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
        ports:
          - name: http
            containerPort: 2222
            hostPort: 2222
        command:
          - /bin/conntracker
        args:
          - --v=3
          - --kubeconfig=/etc/kubeturbo/admin.kubeconfig
        volumeMounts:
        - name: vmt-config
          mountPath: /etc/kubeturbo
          readOnly: true
      restartPolicy: Always
      serviceAccount: turbo-user
      volumes:
      - name: vmt-config
        secret:
          secretName: vmt-config
```

[Download example](K8sConntrack-openshift-secret-ds.yaml?raw=true)

Then you would find k8snet daemonset is created and pods are deployed onto every node in the cluster.

```console
$kubectl get ds
NAME       DESIRED   CURRENT   NODE-SELECTOR   AGE
k8snet     3         3         <none>          24m
$kubectl get po
NAME                         READY     STATUS    RESTARTS   AGE
k8snets-09r05                1/1       Running   0          26m
k8snets-8vf9t                1/1       Running   0          26m
k8snets-voja3                1/1       Running   0          26m
```
