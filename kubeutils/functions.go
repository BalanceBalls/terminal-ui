package kubeutils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"time"

	core "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

func executeRemoteCommand(ctx context.Context, restCfg *rest.Config, clientset *kubernetes.Clientset, pod *v1.Pod, command string) (string, string, error) {
	buf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	request := clientset.CoreV1().RESTClient().
		Post().
		Namespace(pod.Namespace).
		Resource("pods").
		Name(pod.Name).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Command:   []string{"/bin/sh", "-c", command},
			Container: pod.Name,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(restCfg, "POST", request.URL())
	if err != nil {
		return "", "", err
	}

	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: buf,
		Stderr: errBuf,
		Tty:    false,
	})

	if err != nil {
		return "", "", fmt.Errorf("%w failed executing command %s on %v/%v", err, command, pod.Namespace, pod.Name)
	}

	return buf.String(), errBuf.String(), nil
}

func switchLocalK8sContext(ctxName string) error {
	switchCmd := exec.Command(
		"kubectl",
		"config",
		"use-context",
		ctxName,
	)

	switchCmd.Stdout = os.Stdout
	switchCmd.Stderr = os.Stderr

	err := switchCmd.Run()
	if err != nil {
		slog.Error(err.Error())
	}
	return nil
}

func deletePod(ctx context.Context, clientset *kubernetes.Clientset, namespace, podName string) error {
	deletePolicy := metav1.DeletePropagationBackground
	return clientset.CoreV1().Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
}

func checkClusterConnection(ctx context.Context, clientset *kubernetes.Clientset) (bool, error) {
	path := "/healthz"
	content, err := clientset.Discovery().RESTClient().Get().AbsPath(path).DoRaw(ctx)
	if err != nil {
		return false, errors.New("failed to connect to cluster. Reason: " + err.Error())
	}

	contentStr := string(content)
	if contentStr != "ok" {
		return false, errors.New("cluster not healthy")
	}

	return true, nil
}

func createPod(ctx context.Context, clientset *kubernetes.Clientset, namespace, podName string, keepAliveSec int) (*v1.Pod, error) {
	podDefinition := getPodObject(namespace, podName, keepAliveSec)
	pod, err := clientset.CoreV1().Pods(namespace).Create(ctx, podDefinition, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	onlineCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()
	for {
		select {
		case <-onlineCtx.Done():
			return nil, errors.New("cancelled")
		default:
			status, err := clientset.CoreV1().Pods(namespace).Get(ctx, pod.Name, metav1.GetOptions{})
			if err != nil {
				slog.Error(err.Error())
				time.Sleep(time.Second * 3)
				continue
			}

			hasNotReadyContainers := slices.ContainsFunc(
				status.Status.ContainerStatuses,
				func(c v1.ContainerStatus) bool { return c.Ready == false })

			if !hasNotReadyContainers && status.Status.Phase == v1.PodRunning {
				return status, nil
			}

			time.Sleep(time.Second * 3)
		}
	}
}

func getPodObject(namespace, podName string, keepAliveSec int) *core.Pod {
	return &core.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
			Labels: map[string]string{
				"app": "jmeter_pod",
			},
		},
		Spec: core.PodSpec{
			// Restarts are pointless since a pod will be erased on restart
			RestartPolicy: core.RestartPolicyNever,
			Containers: []core.Container{
				{
					Name:            podName,
					Image:           "ubuntu:22.04",
					ImagePullPolicy: core.PullIfNotPresent,
					Command: []string{
						"sleep",
						strconv.Itoa(keepAliveSec),
					},
				},
			},
		},
	}
}
