# Kube Cluster Manager

Cluster Bootstrap

* Move the certificates from assets to /etc/kubernetes/bootstrap-secrets.
* Move the control plane manifests from assets to /etc/kubernetes/manifests.
* Move kubeconfig from assets to /etc/kubernetes/bootstrap-secrets.
* Waits until all control plane components to be ready.
* Install core addons (Calico and Kube-Proxy) using [Tiller Plugin](https://github.com/rimusz/helm-tiller)
* Install Tiller on the master node.

### Possible improvements:

* Install other addons and ensure they are consistent with the desired configuration (helmfile)
* Orchestrate cluster upgrade
    * Drain Pods from Nodes based on business logic. Ex.: set the node as NoSchedule and wait until Jobs to finish before draining.
    * Terminate instance once the node is drained.
    * Keep control-plane static manifests in sync and coordinate rollout of new versions. Ex. Do not continue the upgrade if one fails. (What the Deployment controller does but for static manifests)
