package config

type Config struct {
	IsControlNode bool
	IsWorkerNode  bool
	IsSingleNode  bool
	IsVerbose     bool
	LogFile       string
}

const (
	KubeVersion       = "1.31.5"
	ContainerdVersion = "1.7.20"
	CalicoVersion     = "3.27.5"
	UbuntuVersion     = "22.04"
	KubectlTimeout    = "300s"
	CLIVersion        = "0.3.2"
)
