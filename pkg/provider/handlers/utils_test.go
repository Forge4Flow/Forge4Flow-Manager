package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	faasd "github.com/forge4flow/forge4flow-manager/pkg"
)

func Test_getRequestNamespace(t *testing.T) {
	tables := []struct {
		name              string
		requestNamespace  string
		expectedNamespace string
	}{
		{name: "RequestNamespace is not provided", requestNamespace: "", expectedNamespace: faasd.DefaultFunctionNamespace},
		{name: "RequestNamespace is provided", requestNamespace: "user-namespace", expectedNamespace: "user-namespace"},
	}

	for _, tc := range tables {
		t.Run(tc.name, func(t *testing.T) {
			actualNamespace := getRequestNamespace(tc.requestNamespace)
			if actualNamespace != tc.expectedNamespace {
				t.Errorf("Want: %s, got: %s", actualNamespace, tc.expectedNamespace)
			}
		})
	}
}

func Test_getNamespaceSecretMountPath(t *testing.T) {
	userSecretPath := "/var/openfaas/secrets"
	tables := []struct {
		name               string
		requestNamespace   string
		expectedSecretPath string
	}{
		{name: "Default Namespace is provided", requestNamespace: faasd.DefaultFunctionNamespace, expectedSecretPath: "/var/openfaas/secrets/" + faasd.DefaultFunctionNamespace},
		{name: "User Namespace is provided", requestNamespace: "user-namespace", expectedSecretPath: "/var/openfaas/secrets/user-namespace"},
	}

	for _, tc := range tables {
		t.Run(tc.name, func(t *testing.T) {
			actualNamespace := getNamespaceSecretMountPath(userSecretPath, tc.requestNamespace)
			if actualNamespace != tc.expectedSecretPath {
				t.Errorf("Want: %s, got: %s", actualNamespace, tc.expectedSecretPath)
			}
		})
	}
}

func Test_readNamespaceFromQuery(t *testing.T) {
	tables := []struct {
		name              string
		queryNamespace    string
		expectedNamespace string
	}{
		{name: "No Namespace is provided", queryNamespace: "", expectedNamespace: ""},
		{name: "User Namespace is provided", queryNamespace: "user-namespace", expectedNamespace: "user-namespace"},
	}

	for _, tc := range tables {
		t.Run(tc.name, func(t *testing.T) {

			url := fmt.Sprintf("/test?namespace=%s", tc.queryNamespace)
			r := httptest.NewRequest(http.MethodGet, url, nil)

			actualNamespace := readNamespaceFromQuery(r)
			if actualNamespace != tc.expectedNamespace {
				t.Errorf("Want: %s, got: %s", actualNamespace, tc.expectedNamespace)
			}
		})
	}
}
