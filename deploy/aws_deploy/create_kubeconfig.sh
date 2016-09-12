#!/bin/bash

# Script for generating kubeconfig from credentials
KUBETURBO_CONFIG_PATH=${KUBETURBO_CONFIG_PATH:-"/etc/kubeturbo"}
CLUSTER_NAME=${CLUSTER_NAME:-"kubernetes-aws"}
USER_NAME=${USER_NAME:-"kube-aws-user"}
CONTEXT_NAME=${CONTEXT_NAME:-"kube-aws-context"}

KUBERNETES_SERVER_ADDRESS=""
CA_FILE_PATH=""
CERTIFICATE_PATH=""
KEY_PATH="" 

if [[ $# != 4 ]]; then
  echo "Not enough argument. Run the script with --server=<KUBERNETES_SERVER_ADDRESS> --ca=<CA_FILE_PATH> --cert=<CERTIFICATE_PATH> --key=<KEY_PATH>"
  exit 1
fi

if [[ ! -d $KUBETURBO_CONFIG_PATH ]]; then
  mkdir -p $KUBETURBO_CONFIG_PATH
fi

KUBECONFIG_PATH=$KUBETURBO_CONFIG_PATH/kubeconfig-aws

args=$(getopt -o '' -l server:,ca:,cert:,key: -- "$@")
if [[ $? -ne 0 ]]; then
  >&2 echo "Error in getopt"
  exit 1
fi

eval set -- "${args}"

while true; do
    case $1 in 
      --server) 
  shift
  if [[ -z "$1" ]; then
          >&2 echo "empty argument to --server flag"
          exit 1
        fi
  KUBERNETES_SERVER_ADDRESS=$1
  shift
  ;;
      --ca) 
  shift
  if [ -z "$1" ]; then
          >&2 echo "empty argument to --ca flag"
          exit 1
        fi
  CA_FILE_PATH=$1
  shift
  ;;
      --cert) 
  shift
  if [ -z "$1" ]; then
          >&2 echo "empty argument to --cert flag"
          exit 1
        fi
  CERTIFICATE_PATH=$1
  shift
  ;;
      --key) 
  shift
  if [ -z "$1" ]; then
          >&2 echo "empty argument to --key flag"
          exit 1
        fi
  KEY_PATH=$1
  shift
  ;;
      --)
  shift 
  break;
  ;;
      *) echo "Error"; exit 1 ;;
    esac
done

if [[ $KUBERNETES_SERVER_ADDRESS == "" ]]; then 
    >&2 echo "!Error: Kubernetes server address not provided. Run the script with --server=<SERVER_ADDRESS> "
    exit 1
fi

if [[ $CA_FILE_PATH == "" ]]; then 
    >&2 echo "!Error: Path to certificate authority file is not provided. Run the script with --ca=<PATH_TO_CA_FIL
E>"
    exit 1
elif [[ ! -f $CA_FILE_PATH ]]; then
    >&2 echo "The ca path provided is invalid."
    exit 1
fi

if [[ $CERTIFICATE_PATH == "" ]]; then 
    >&2 echo "!Error: Path to client certificate file is not provided. Run the script with --cert=<PATH_TO_CLIENT_
CERTIFICATE>"
    exit 1
elif [[ ! -f $CERTIFICATE_PATH ]]; then
    >&2 echo "The certificate path provided is invalid."
    exit 1
fi

if [[ $KEY_PATH == "" ]]; then 
    >&2 echo "!Error: Path to client key file is not provided. Run the script with --key=<PATH_TO_CLIENT_KEY>"
    exit 1
elif [[ ! -f $KEY_PATH ]]; then
    >&2 echo "The key path provided is invalid."
    exit 1
fi

kubectl config set-cluster $CLUSTER_NAME \
  --server=$KUBERNETES_SERVER_ADDRESS \
  --certificate-authority=$CA_FILE_PATH \
  --embed-certs=true \
  --kubeconfig=$KUBECONFIG_PATH

kubectl config set-credentials $USER_NAME \
  --client-certificate=$CERTIFICATE_PATH \
  --client-key=$KEY_PATH \
  --embed-certs=true \
  --kubeconfig=$KUBECONFIG_PATH

kubectl config set-context $CONTEXT_NAME \
  --cluster=$CLUSTER_NAME \
  --user=$USER_NAME \
  --kubeconfig=$KUBECONFIG_PATH

echo "Generate Kubeconfig to $KUBECONFIG_PATH"
