package controller

import (
	"fmt"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

func resolvePodPort(pod *corev1.Pod, portNameOrNum string) (int32, error) {
	if num, err := strconv.Atoi(portNameOrNum); err == nil {
		return int32(num), nil
	}

	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			if port.Name == portNameOrNum {
				return port.ContainerPort, nil
			}
		}
	}
	return 0, fmt.Errorf("port name %q not found in pod containers", portNameOrNum)
}

func parseKeyValueString(str string) map[string]string {
	result := make(map[string]string)
	if str == "" {
		return result
	}
	pairs := strings.Split(str, ",")
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) == 2 {
			result[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return result
}
