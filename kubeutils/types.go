package kubeutils

import (
	"log/slog"
	"os/exec"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Cluster struct {
	RestCfg         *rest.Config
	Clientset       *kubernetes.Clientset
	Namespace       string
	KubeCtxName     string
	PodPrefix       string
	PodsCache       *PodsCache
	PodKeepAliveSec int
	Logger          slog.Logger
}

type TestInfo struct {
	PodName          string
	PropFileName     string
	ScenarioFileName string
}

type ActionDone struct {
	PodName  string
	Name     string
	Duration time.Duration
}

type remoteCommand struct {
	displayName string
	command     string
}

type localCommand struct {
	displayName string
	command     *exec.Cmd
}

type PodsCache struct {
	Pods map[string]*v1.Pod

	mu sync.Mutex
}
