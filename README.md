# Go Install Kubernetes

This is a Go version of the [install-kubernetes](https://github.com/ccollicutt/install-kubernetes) script. It does the same thing, slighltly differently, but is written in Go, so it is a single binary.

## Example Installation

This is a gif recording of the install process. It takes a minute or two to completely install Kubernetes, so if you want to see the process, you can watch the entire gif. But it is a couple minutes long. If you have an Ubuntu 22.04 virtual machine, you can run it yourself, and it will be just as fast.

![Install gif](img/install.gif)

## Why Use This?

* It is a single binary that is easy to use.
* It can create a single node cluster that can schedule pods, good for local development, an alternative to Minkube and such.
* It uses very standard Kubernetes components from Ubuntu, nothing special or cutting edge. Just install Kubernetes from packages and use Kubeadm to setup the control plane and worker nodes.

## Caveats

* This program does not create the virtual machines. It only installs Kubernetes onto them. This means you can create the nodes in any way you want, but they must exist before running this program.
* It does not co-ordinate the install across multiple nodes at once. What it does is install the control plane on the first node, and then the worker nodes one at a time, joining them to the control plane with the kubeadm join command that is produced by the control plane node.
* Currently only Ubuntu 22.04 is supported.

## Stack 

* kubeadm
* containerd
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

If you want a single node that is a control plane and worker, create one virtual machine.

If you want a control plane and worker, create two or more virtual machines.

> NOTE: The program does not create the virtual machines (nodes). It only installs Kubernetes onto them. This means you can create the nodes in any way you want, but they must exist before running this program.

### Suggested Node Sizes

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

```bash
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
```

## Install Kubernetes Onto the Nodes

Order of Operations

1. Build either 1) one virtual machine that is a control plane and worker node or 2) two or more virtual machines that are control plane and worker nodes
1. Deploy the control plane node with this program
2. Configure the worker node with this program
3. Get the join command from the control plane node
4. Run that join command on the worker node(s) to join them to the Kubernetes cluster

### Control Plane Node

Use the `-c` flag if the node is a control plane node.

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

This will untaint the control plane node so that pods can be scheduled on it, giving you a single node cluster that you can use for development.

## Thanks

This was originally based on the Bash script version I created called [install-kubernetes](https://github.com/ccollicutt/install-kubernetes) which as itself originally based on the [Killer.sh CKS install script](https://github.com/killer-sh/cks-course-environment).
