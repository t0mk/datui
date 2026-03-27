package config

import (
	"strings"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Config holds the loaded kubeconfig REST config plus context metadata.
type Config struct {
	RestConfig  *rest.Config
	ContextName string
	ProjectName string // extracted from datum-project-<name>-<id> context
	Namespace   string
}

// Load reads ~/.kube/config (or KUBECONFIG), prefers a datum-project-* context,
// and returns a Config.
func Load() (*Config, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()

	// Read raw config to inspect contexts before choosing one.
	rawLoader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, nil)
	rawCfg, err := rawLoader.RawConfig()
	if err != nil {
		return nil, err
	}

	// Prefer a datum-project-* context over whatever is current.
	contextName := rawCfg.CurrentContext
	projectName := ""
	for name := range rawCfg.Contexts {
		if strings.HasPrefix(name, "datum-project-") {
			contextName = name
			// "datum-project-personal-project-bafacb5a" → "personal-project-bafacb5a"
			projectName = strings.TrimPrefix(name, "datum-project-")
			break
		}
	}

	overrides := &clientcmd.ConfigOverrides{CurrentContext: contextName}
	loader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)

	restCfg, err := loader.ClientConfig()
	if err != nil {
		return nil, err
	}

	ns, _, err := loader.Namespace()
	if err != nil {
		ns = "default"
	}

	return &Config{
		RestConfig:  restCfg,
		ContextName: contextName,
		ProjectName: projectName,
		Namespace:   ns,
	}, nil
}
