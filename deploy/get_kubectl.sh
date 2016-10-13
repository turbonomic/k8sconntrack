#!/bin/bash
KC_VERS="${1:-1.2.4}"
/usr/bin/curl \
    --create-dirs \
    -o "$HOME/bin/kubectl-$KC_VERS" \
    "https://storage.googleapis.com/kubernetes-release/release/v$KC_VERS/bin/linux/amd64/kubectl"
chmod +x "$HOME/bin/kubectl-$KC_VERS"
pushd "$HOME/bin" &>/dev/null
ln -sf "kubectl-$KC_VERS" kubectl 2>/dev/null
ln -sf kubectl kc 2>/dev/null
popd &>/dev/null

if [ ! -f "$HOME/.kube/config" ]; then
    MASTER_HOST=$(hostname)
    CA_CERT=/etc/kubernetes/ssl/ca.pem
    ADMIN_KEY=/etc/kubernetes/ssl/admin-key.pem
    ADMIN_CERT=/etc/kubernetes/ssl/admin.pem
    kubectl config set-cluster default-cluster \
        --server=https://$MASTER_HOST \
        --certificate-authority=$CA_CERT
    kubectl config set-credentials default-admin \
        --certificate-authority=$CA_CERT \
        --client-key=$ADMIN_KEY \
        --client-certificate=$ADMIN_CERT
    kubectl config set-context default-system \
        --cluster=default-cluster \
        --user=default-admin
    kubectl config use-context default-system
fi