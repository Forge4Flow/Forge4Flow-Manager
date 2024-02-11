package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/opencontainers/runtime-spec/specs-go"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/forge4flow/forge4flow-manager/pkg"
	manager "github.com/forge4flow/forge4flow-manager/pkg"
	"github.com/forge4flow/forge4flow-manager/pkg/cninetwork"
)

type Function struct {
	name        string
	namespace   string
	image       string
	pid         uint32
	replicas    int
	IP          string
	labels      map[string]string
	annotations map[string]string
	secrets     []string
	envVars     map[string]string
	envProcess  string
	memoryLimit int64
	createdAt   time.Time
}

// ListFunctions returns a map of all functions with running tasks on namespace
func ListFunctions(client *containerd.Client, namespace string) (map[string]*Function, error) {

	// Check if namespace exists, and it has the openfaas label
	valid, err := validNamespace(client.NamespaceService(), namespace)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, errors.New("namespace not valid")
	}

	ctx := namespaces.WithNamespace(context.Background(), namespace)
	functions := make(map[string]*Function)

	containers, err := client.Containers(ctx)
	if err != nil {
		return functions, err
	}

	for _, c := range containers {
		name := c.ID()
		f, err := GetFunction(client, name, namespace)
		if err != nil {
			log.Printf("skipping %s, error: %s", name, err)
		} else {
			functions[name] = &f
		}
	}

	return functions, nil
}

// GetFunction returns a function that matches name
func GetFunction(client *containerd.Client, name string, namespace string) (Function, error) {
	ctx := namespaces.WithNamespace(context.Background(), namespace)
	fn := Function{}

	c, err := client.LoadContainer(ctx, name)
	if err != nil {
		return Function{}, fmt.Errorf("unable to find function: %s, error %w", name, err)
	}

	image, err := c.Image(ctx)
	if err != nil {
		return fn, err
	}

	containerName := c.ID()
	allLabels, labelErr := c.Labels(ctx)

	if labelErr != nil {
		log.Printf("cannot list container %s labels: %s", containerName, labelErr)
	}

	labels, annotations := buildLabelsAndAnnotations(allLabels)

	spec, err := c.Spec(ctx)
	if err != nil {
		return Function{}, fmt.Errorf("unable to load function %s error: %w", name, err)
	}

	info, err := c.Info(ctx)
	if err != nil {
		return Function{}, fmt.Errorf("can't load info for: %s, error %w", name, err)
	}

	envVars, envProcess := readEnvFromProcessEnv(spec.Process.Env)
	secrets := readSecretsFromMounts(spec.Mounts)

	fn.name = containerName
	fn.namespace = namespace
	fn.image = image.Name()
	fn.labels = labels
	fn.annotations = annotations
	fn.secrets = secrets
	fn.envVars = envVars
	fn.envProcess = envProcess
	fn.createdAt = info.CreatedAt
	fn.memoryLimit = readMemoryLimitFromSpec(spec)

	replicas := 0
	task, err := c.Task(ctx, nil)
	if err == nil {
		// Task for container exists
		svc, err := task.Status(ctx)
		if err != nil {
			return Function{}, fmt.Errorf("unable to get task status for container: %s %w", name, err)
		}

		if svc.Status == "running" {
			replicas = 1
			fn.pid = task.Pid()

			// Get container IP address
			ip, err := cninetwork.GetIPAddress(name, task.Pid())
			if err != nil {
				return Function{}, err
			}
			fn.IP = ip
		}
	} else {
		replicas = 0
	}

	fn.replicas = replicas
	return fn, nil
}

func readEnvFromProcessEnv(env []string) (map[string]string, string) {
	foundEnv := make(map[string]string)
	fprocess := ""
	for _, e := range env {
		kv := strings.Split(e, "=")
		if len(kv) == 1 {
			continue
		}

		if kv[0] == "PATH" {
			continue
		}

		if kv[0] == "fprocess" {
			fprocess = kv[1]
			continue
		}

		foundEnv[kv[0]] = kv[1]
	}

	return foundEnv, fprocess
}

func readSecretsFromMounts(mounts []specs.Mount) []string {
	secrets := []string{}
	for _, mnt := range mounts {
		x := strings.Split(mnt.Destination, "/var/openfaas/secrets/")
		if len(x) > 1 {
			secrets = append(secrets, x[1])
		}

	}
	return secrets
}

// buildLabelsAndAnnotations returns a separated list with labels first,
// followed by annotations by checking each key of ctrLabels for a prefix.
func buildLabelsAndAnnotations(ctrLabels map[string]string) (map[string]string, map[string]string) {
	labels := make(map[string]string)
	annotations := make(map[string]string)

	for k, v := range ctrLabels {
		if strings.HasPrefix(k, annotationLabelPrefix) {
			annotations[strings.TrimPrefix(k, annotationLabelPrefix)] = v
		} else {
			labels[k] = v
		}
	}

	return labels, annotations
}

func ListNamespaces(client *containerd.Client) []string {
	set := []string{}
	store := client.NamespaceService()
	namespaces, err := store.List(context.Background())
	if err != nil {
		log.Printf("Error listing namespaces: %s", err.Error())
		set = append(set, manager.DefaultFunctionNamespace)
		return set
	}

	for _, namespace := range namespaces {
		labels, err := store.Labels(context.Background(), namespace)
		if err != nil {
			log.Printf("Error listing label for namespace %s: %s", namespace, err.Error())
			continue
		}

		if _, found := labels[pkg.NamespaceLabel]; found {
			set = append(set, namespace)
		}

		if !findNamespace(manager.DefaultFunctionNamespace, set) {
			set = append(set, manager.DefaultFunctionNamespace)
		}
	}

	return set
}

func findNamespace(target string, items []string) bool {
	for _, n := range items {
		if n == target {
			return true
		}
	}
	return false
}

func readMemoryLimitFromSpec(spec *specs.Spec) int64 {
	if spec.Linux == nil || spec.Linux.Resources == nil || spec.Linux.Resources.Memory == nil || spec.Linux.Resources.Memory.Limit == nil {
		return 0
	}
	return *spec.Linux.Resources.Memory.Limit
}
