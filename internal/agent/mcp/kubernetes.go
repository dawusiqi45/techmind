package mcp

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeOnce    sync.Once
	kubeClient  kubernetes.Interface
	kubeInitErr error
)

// KubernetesSnapshot 只读取指定服务的 Pod、Warning Event 和 Deployment。
func KubernetesSnapshot(ctx context.Context, namespace, service string) Evidence {
	if strings.TrimSpace(service) == "" {
		return Evidence{"kubernetes_status": "skipped", "kubernetes_error": "service is required"}
	}
	client, err := getKubernetesClient()
	if err != nil {
		return Evidence{"kubernetes_status": "unavailable", "kubernetes_error": err.Error()}
	}
	if namespace == "" {
		namespace = "techmind"
	}
	evidence := Evidence{}
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return Evidence{"kubernetes_error": err.Error()}
	}
	podNames := make([]string, 0, 5)
	podSummary := make([]string, 0, 5)
	for _, pod := range pods.Items {
		if !matchesService(pod.Name, pod.Labels, service) {
			continue
		}
		if len(podNames) >= 5 {
			break
		}
		restarts := int32(0)
		for _, status := range pod.Status.ContainerStatuses {
			restarts += status.RestartCount
		}
		podNames = append(podNames, pod.Name)
		podSummary = append(podSummary, fmt.Sprintf("%s phase=%s ready=%t restarts=%d", pod.Name, pod.Status.Phase, isPodReady(pod.Status.Conditions), restarts))
	}
	evidence["k8s_pods"] = podSummary

	events, err := client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		warnings := make([]string, 0, 10)
		for _, event := range events.Items {
			if event.Type != "Warning" || !contains(podNames, event.InvolvedObject.Name) {
				continue
			}
			warnings = append(warnings, fmt.Sprintf("%s: %s", event.Reason, event.Message))
			if len(warnings) == 10 {
				break
			}
		}
		evidence["k8s_warning_events"] = warnings
	}

	if service != "" {
		deployment, err := client.AppsV1().Deployments(namespace).Get(ctx, service, metav1.GetOptions{})
		if err == nil {
			evidence["k8s_deployment"] = fmt.Sprintf("%s desired=%d available=%d updated=%d", deployment.Name, deployment.Status.Replicas, deployment.Status.AvailableReplicas, deployment.Status.UpdatedReplicas)
		}
	}

	return evidence
}

// KubernetesLogSnapshot 读取单个匹配 Pod 最近十分钟的末尾日志，严格限制行数与字节数。
func KubernetesLogSnapshot(ctx context.Context, namespace, service string) Evidence {
	if strings.TrimSpace(service) == "" {
		return Evidence{"k8s_logs_error": "service is required"}
	}
	client, err := getKubernetesClient()
	if err != nil {
		return Evidence{"k8s_logs_error": err.Error()}
	}
	if namespace == "" {
		namespace = "techmind"
	}
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return Evidence{"k8s_logs_error": err.Error()}
	}
	logs := make([]string, 0, 3)
	for _, pod := range pods.Items {
		if !matchesService(pod.Name, pod.Labels, service) {
			continue
		}
		for _, container := range pod.Spec.Containers {
			if len(logs) >= 3 {
				break
			}
			tail, since := int64(50), int64(600)
			stream, err := client.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{Container: container.Name, TailLines: &tail, SinceSeconds: &since}).Stream(ctx)
			if err != nil {
				logs = append(logs, fmt.Sprintf("%s/%s: log error: %v", pod.Name, container.Name, err))
				continue
			}
			body, readErr := io.ReadAll(io.LimitReader(stream, 8*1024))
			_ = stream.Close()
			if readErr != nil {
				logs = append(logs, fmt.Sprintf("%s/%s: log error: %v", pod.Name, container.Name, readErr))
				continue
			}
			logs = append(logs, fmt.Sprintf("%s/%s:\n%s", pod.Name, container.Name, redactLog(string(body))))
		}
		if len(logs) >= 3 {
			break
		}
	}
	if len(logs) == 0 {
		return Evidence{"k8s_logs": "no matching pod"}
	}
	return Evidence{"k8s_log_samples": logs}
}

func getKubernetesClient() (kubernetes.Interface, error) {
	kubeOnce.Do(func() {
		cfg, err := rest.InClusterConfig()
		if err != nil {
			kubeconfig := os.Getenv("KUBECONFIG")
			cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		}
		if err != nil {
			kubeInitErr = err
			return
		}
		cfg.Timeout = 10 * time.Second
		kubeClient, kubeInitErr = kubernetes.NewForConfig(cfg)
	})
	return kubeClient, kubeInitErr
}

func matchesService(name string, labels map[string]string, service string) bool {
	if service == "" {
		return true
	}
	if strings.HasPrefix(name, service+"-") || name == service {
		return true
	}
	component := strings.TrimPrefix(service, "techmind-")
	return labels["app.kubernetes.io/component"] == component
}

func isPodReady(conditions []corev1.PodCondition) bool {
	for _, condition := range conditions {
		if condition.Type == "Ready" {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func redactLog(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lower := strings.ToLower(line)
		if containsSensitiveLogKey(lower) {
			lines[i] = "[redacted sensitive log line]"
		}
	}
	return strings.Join(lines, "\n")
}

func containsSensitiveLogKey(line string) bool {
	for _, key := range []string{"authorization", "password", "passwd", "token", "api_key", "api-key", "apikey", "secret", "cookie", "set-cookie", "bearer ", "sk-"} {
		if strings.Contains(line, key) {
			return true
		}
	}
	return false
}
