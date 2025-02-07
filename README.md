# Go Install Kubernetes

This is a Go version of the [install-kubernetes](https://github.com/ccollicutt/install-kubernetes) script. It does the same thing, slighltly differently, but is written in Go, so it is a single binary.

## Stack 

* kubeadm
* containerd - *NOTE: is installed from binary download, not apt package*
* runc
* Kubernetes from Ubuntu
* Calico CNI

## Usage

### Setup Some Virtual Machines

Build at least two virtual machines, one for the control plane and one worker. Add more workers if you would like but this program will only be able to setup a single control plane node.

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
# Install Go if not already installed
# Download and install the program
git clone https://github.com/ccollicutt/go-install-kubernetes
cd go-install-kubernetes
go build
# Note the "-control-plane" flag here
./go-install-kubernetes -control-plane
```

### Worker Nodes

On a worker node run:

```
git clone https://github.com/ccollicutt/go-install-kubernetes
cd go-install-kubernetes
go build
./go-install-kubernetes
```

Then finally connect the worker node to the control plane node with the kubeadm command based on the output of the below which is run on the CP node.

```
kubeadm token create --print-join-command --ttl 0
```

Run the output of that command on the worker nodes.

### Single Node / Control Plane That Can Have Pods Scheduled On It

If you'd like a single node "cluster", ie. be able to schedule pods on the control plane, then run with the `-single-node` flag.

```
./go-install-kubernetes -single-node
```

This will untaint the control plane node so that pods can be scheduled on it.

## Thanks

This was originally based on the Bash script version called [install-kubernetes](https://github.com/ccollicutt/install-kubernetes) which as itself originally based on the [Killer.sh CKS install script](https://github.com/killer-sh/cks-course-environment).
