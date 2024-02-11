package pkg

const (
	// DefaultFunctionNamespace is the default containerd namespace functions are created
	DefaultFunctionNamespace = "functions4flow"

	// NamespaceLabel indicates that a namespace is managed by f4f-manager
	NamespaceLabel = "functions4flow"

	// ForgedNamespace is the containerd namespace services are created
	ForgedNamespace = "f4f-services"

	f4fServicesPullAlways = false

	defaultSnapshotter = "overlayfs"
)
