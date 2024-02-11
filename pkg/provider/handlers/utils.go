package handlers

import (
	"context"
	"net/http"
	"path"

	"github.com/forge4flow/forge4flow-manager/pkg"
	manager "github.com/forge4flow/forge4flow-manager/pkg"
	provider "github.com/forge4flow/forge4flow-manager/pkg/provider"
)

func getRequestNamespace(namespace string) string {

	if len(namespace) > 0 {
		return namespace
	}
	return manager.DefaultFunctionNamespace
}

func readNamespaceFromQuery(r *http.Request) string {
	q := r.URL.Query()
	return q.Get("namespace")
}

func getNamespaceSecretMountPath(userSecretPath string, namespace string) string {
	return path.Join(userSecretPath, namespace)
}

// validNamespace indicates whether the namespace is eligable to be
// used for OpenFaaS functions.
func validNamespace(store provider.Labeller, namespace string) (bool, error) {
	if namespace == manager.DefaultFunctionNamespace {
		return true, nil
	}

	labels, err := store.Labels(context.Background(), namespace)
	if err != nil {
		return false, err
	}

	// check for true to keep it backward compatible
	if value, found := labels[pkg.NamespaceLabel]; found && (value == "true" || value == "1") {
		return true, nil
	}

	return false, nil
}
