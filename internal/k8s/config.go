package k8s

import (
	"flag"
	"os"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/kyma-project/eventing-tools/internal/logger"
)

func ConfigOrDie() *rest.Config {
	c, err := rest.InClusterConfig()
	if err != nil {
		path := getKubeconfigPath()
		c = getConfigOrDie(path)
	}
	return c
}

func getKubeconfigPath() string {
	path := os.Getenv("KUBECONFIG")
	if len(path) == 0 {
		defaultPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
		path = *flag.String("kubeconfig", defaultPath, "Kubeconfig absolute path")
		flag.Parse()
	}
	return path
}

func getConfigOrDie(kubeconfigPath string) *rest.Config {
	c, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	logger.FatalIfError(err)
	return c
}
