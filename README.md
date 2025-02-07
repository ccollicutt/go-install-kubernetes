# Go Install Kubernetes

This is a Go version of the [install-kubernetes](https://github.com/ccollicutt/install-kubernetes) script. It does the same thing, slighltly differently, but is written in Go, so it is a single binary.

## Stack 

* kubeadm
* containerd - *NOTE: is installed from binary download, not apt package*
* runc
* Kubernetes from Ubuntu
* Calico CNI

## Installation

To download the latest release:

```bash
# Get latest version and download the binary
VERSION=$(basename $(curl -Ls -o /dev/null -w %{url_effective} https://github.com/ccollicutt/go-install-kubernetes/releases/latest))
curl -Ls -o go-install-kubernetes "https://github.com/ccollicutt/go-install-kubernetes/releases/download/${VERSION}/go-install-kubernetes"

# Make it executable
chmod +x go-install-kubernetes

# Move to a directory in your PATH
sudo mv go-install-kubernetes /usr/local/bin/
```

## Usage

### Setup Some Virtual Machines

IF you want a single node that is a control plane and worker, create one virtual machine.

If you want a control plane and worker, create two or more virtual machines.

> NOTE: The program does not create the virtual machines (nodes). It only installs Kubernetes onto them. This means you can create the nodes in any way you want, but they must exist before running this program.

#### Suggested Node Sizes

Control Node
* 4G memory
* 40G disk
* 2 CPUs

Worker Nodes
* 8G memory
* 40G disk
* 4 CPUs

Combined Control Plane / Worker / Single Node "Cluster"
* 8G memory
* 40G disk
* 4 CPUs

## Usage

```
$ go-install-kubernetes -h
USAGE:
  go-install-kubernetes [options]

OPTIONS:
  -c  Configure as a control plane node
  -w  Configure as a worker node
  -s  Configure as a single node (control plane + worker)
  -v  Enable verbose output
  -h  Show this help message
  --version  Show version information
  --export-manifests  Export embedded Calico manifests to disk

At least one of -c, -w, or -s must be specified

## Install Kubernetes Onto the Nodes

Order of Operations

1. Build at least two virtual machines
1. Deploy the control plane node with this program
2. Configure the worker node with this program
3. Get the join command from the control plane node
4. Run that join command on the worker node(s) to join them to the Kubernetes cluster

### Control Plane Node

Use the `-control-plane` flag if the node is a control plane node.

Normally you would have one control plane node and `x` worker nodes.

On a control plane node run:

```
go-install-kubernetes -c
```

### Worker Nodes

On a worker node run:

```
go-install-kubernetes -w
```

Then finally connect the worker node to the control plane node with the kubeadm command based on the output of the below which is run on the CP node.

```
kubeadm token create --print-join-command --ttl 0
```

Run the output of that command on the worker nodes.

### Single Node / Control Plane That Can Have Pods Scheduled On It

If you'd like a single node "cluster", ie. be able to schedule pods on the control plane, then run with the `-single-node` flag.

```
go-install-kubernetes -s
```

This will untaint the control plane node so that pods can be scheduled on it.

## Thanks

This was originally based on the Bash script version called [install-kubernetes](https://github.com/ccollicutt/install-kubernetes) which as itself originally based on the [Killer.sh CKS install script](https://github.com/killer-sh/cks-course-environment).
