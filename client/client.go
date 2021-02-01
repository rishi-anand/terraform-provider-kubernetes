package k8s

import (
	"crypto/tls"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetClient() (*kubernetes.Clientset, error) {
	config, err := GetConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func GetClientFromKubeconfig(kubeconfig, masterURL string) (*kubernetes.Clientset, error) {
	config, err := GetConfigFromKubeconfig(kubeconfig, masterURL)
	if err != nil {
		return nil, err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func GetDynamicClient() (dynamic.Interface, error) {
	config, err := GetConfig()
	if err != nil {
		return nil, err
	}

	return GetDynamicClientForConfig(config)

}

func GetDynamicClientForConfig(config *rest.Config) (dynamic.Interface, error) {

	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return dynClient, nil

}

func GetConfig() (*rest.Config, error) {
	// If an env variable is specified with the config locaiton, use that
	if len(os.Getenv("KUBECONFIG")) > 0 {
		return clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	}
	// If no explicit location, try the in-cluster config
	if c, err := rest.InClusterConfig(); err == nil {
		return c, nil
	}
	// If no in-cluster config, try the default location in the user's home directory
	if usr, err := user.Current(); err == nil {
		if c, err := clientcmd.BuildConfigFromFlags(
			"", filepath.Join(usr.HomeDir, ".kube", "config")); err == nil {
			return c, nil
		}
	}

	return nil, fmt.Errorf("could not locate a kubeconfig")
}

func GetConfigFromKubeconfig(kubeconfig, masterURL string) (*rest.Config, error) {
	// If a flag is specified with the config location, use that
	if len(kubeconfig) > 0 {
		return clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	}
	// If an env variable is specified with the config locaiton, use that
	if len(os.Getenv("KUBECONFIG")) > 0 {
		return clientcmd.BuildConfigFromFlags(masterURL, os.Getenv("KUBECONFIG"))
	}
	// If no explicit location, try the in-cluster config
	if c, err := rest.InClusterConfig(); err == nil {
		return c, nil
	}
	// If no in-cluster config, try the default location in the user's home directory
	if usr, err := user.Current(); err == nil {
		if c, err := clientcmd.BuildConfigFromFlags(
			"", filepath.Join(usr.HomeDir, ".kube", "config")); err == nil {
			return c, nil
		}
	}

	return nil, fmt.Errorf("could not locate a kubeconfig")
}

func GetInClusterConfigCert() (caPEMBlock, certPEMBlock, keyPEMBlock []byte, err error) {
	cfg, err := GetConfig()
	if err != nil {
		return
	}

	if len(cfg.TLSClientConfig.CAData) > 0 {
		caPEMBlock = cfg.TLSClientConfig.CAData
	}

	if len(cfg.TLSClientConfig.CertData) > 0 {
		certPEMBlock = cfg.TLSClientConfig.CertData
	}

	if len(cfg.TLSClientConfig.KeyData) > 0 {
		keyPEMBlock = cfg.TLSClientConfig.KeyData
	}
	return
}

func GetTlsConfig() (*tls.Config, error) {
	_, certPEMBlock, keyPEMBlock, err := GetInClusterConfigCert()
	if err != nil {
		return nil, err
	}

	cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		return nil, err
	}
	return &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}, nil
}
