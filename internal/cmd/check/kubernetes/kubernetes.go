package kubernetes

import (
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/cobra"

	"cdr.dev/slog"
	"cdr.dev/slog/sloggers/sloghuman"

	kclient "k8s.io/client-go/kubernetes"
	// Kubernetes authentication plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/exec"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	_ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/cdr/coder-doctor/internal/checks/kube"
	"github.com/cdr/coder-doctor/internal/humanwriter"
)

func NewCommand() *cobra.Command {
	kubernetesCmd := &cobra.Command{
		Use:   "kubernetes",
		Short: "scan the Kubernetes cluster for compatibility",
		RunE:  run,
	}

	kubernetesCmd.PersistentFlags().String(clientcmd.FlagClusterName, "", "the name of the Kubernetes cluster to use")
	kubernetesCmd.PersistentFlags().String(clientcmd.FlagContext, "", "the name of the Kubernetes context to use")
	kubernetesCmd.PersistentFlags().String(clientcmd.RecommendedConfigPathFlag, "", "path to the Kubernetes configuration file")
	kubernetesCmd.PersistentFlags().StringP(clientcmd.FlagNamespace, "n", "", "the name of the Kubernetes namespace to deploy into")

	return kubernetesCmd
}

func getConfigOverridesFromFlags(cmd *cobra.Command) (*clientcmd.ConfigOverrides, error) {
	var err error

	overrides := &clientcmd.ConfigOverrides{}

	overrides.CurrentContext, err = cmd.Flags().GetString(clientcmd.FlagContext)
	if err != nil {
		return nil, err
	}

	overrides.Context.Namespace, err = cmd.Flags().GetString(clientcmd.FlagNamespace)
	if err != nil {
		return nil, err
	}

	overrides.Context.Cluster, err = cmd.Flags().GetString(clientcmd.FlagClusterName)
	if err != nil {
		return nil, err
	}

	return overrides, nil
}

func run(cmd *cobra.Command, _ []string) error {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	var err error
	loadingRules.ExplicitPath, err = cmd.Flags().GetString(clientcmd.RecommendedConfigPathFlag)
	if err != nil {
		return err
	}

	overrides, err := getConfigOverridesFromFlags(cmd)
	if err != nil {
		return err
	}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides).ClientConfig()
	if err != nil {
		return err
	}

	clientset, err := kclient.NewForConfig(config)
	if err != nil {
		return err
	}

	coderVersion, err := cmd.Flags().GetString("coder-version")
	if err != nil {
		return err
	}

	cv, err := semver.NewVersion(coderVersion)
	if err != nil {
		return err
	}

	log := slog.Make(sloghuman.Sink(cmd.OutOrStdout()))
	verbosity, err := cmd.Flags().GetInt("verbosity")
	if err != nil {
		return err
	}

	// TODO: this is pretty arbitrary, use a defined verbosity similar to
	// kubectl
	if verbosity > 5 {
		log = log.Leveled(slog.LevelDebug)
	}

	checker := kube.NewKubernetesChecker(
		clientset,
		kube.WithLogger(log),
		kube.WithCoderVersion(cv),
		kube.WithWriter(humanwriter.New(os.Stdout)),
	)

	err = checker.Run(cmd.Context())
	if err != nil {
		return err
	}

	return nil
}
