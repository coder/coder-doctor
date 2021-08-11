package kubernetes

import (
	"github.com/Masterminds/semver/v3"
	"github.com/cdr/coder-doctor/internal/api"
	"github.com/cdr/coder-doctor/internal/checks"
	"github.com/spf13/cobra"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/exec"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	_ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
	"k8s.io/client-go/tools/clientcmd"
)

func NewCommand() *cobra.Command {
	kubernetesCmd := &cobra.Command{
		Use:   "kubernetes",
		Short: "scan the Kubernetes cluster for compatibility",
		RunE:  run,
	}

	return kubernetesCmd
}

func run(cmd *cobra.Command, args []string) error {
	config, err := clientcmd.BuildConfigFromFlags("", "/home/coder/.kube/config")
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	coderVersion, err := cmd.InheritedFlags().GetString("coder-version")
	if err != nil {
		panic(err.Error())
	}

	cv, err := semver.NewVersion(coderVersion)
	if err != nil {
		panic(err.Error())
	}

	results := checks.RunKubernetes(cmd.Context(), api.CheckOptions{
		Kubernetes:   clientset,
		CoderVersion: cv,
	})
	PrintResults(cmd.OutOrStdout(), results)

	return nil
}
