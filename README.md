# K8sconnection

## What is K8sconnection
K8sconnection keeps track of networking metrics of communication between pods in Kubernetes cluster.
Right now it helps to answer the following questions:

1. It exposes the number of transactions for applications at the service level. So it helps users to understand how the applications are used and determine whether to scale out/in applications automatically based on QoS.
2. It exposes end-to-end traffic information among pods in Kubernetes cluster. So that users can understand the communication pattern between the pods and make initial placement or replacement based on networking topology.

## Architecture

![K8sconnection Architecture](https://cloud.githubusercontent.com/assets/7660489/18649719/03feeb8a-7e8f-11e6-9995-9de6ec9e05b3.png)

## Run K8sconnection on Kubernetes
As K8sconnection gathers networking metrics from netfilter, it requires to deploy K8sconnection application on every node in the Kubernetes cluster. The best way to deploy K8sconnection in a Kubernetes cluster is to deploy it as DaemonSet.
You can find deploy guide for different scenarios [here](deploy)

## How to get metrics from K8sconnection
Metrics are exposed on port 2222 of each host.

### Transaction Metrics

Transaction metrics expose serviceID and the number of transactions of each endpoint.
To get transaction metrics, go to <HOST_IP>:2222/transactions, an output example is like the following:
```json
[{
  "serviceID":"default/redis-slave",
  "endpointCounter":{
    "172.17.0.4":3,
    "172.17.0.5":2
  }
}]
```

### Flow Metrics
Flow metrics expose the amount of traffic between two endpoints in bytes.

To get flow metrics, go to <HOST_IP>:2222/flows
```json
[{
  "uid":"172.17.0.3:6379->172.17.0.5:38318#1471007430",
  "source":"172.17.0.3",
  "destination":"172.17.0.5",
  "value":52,
  "timestamp":1471010475
}]
```
