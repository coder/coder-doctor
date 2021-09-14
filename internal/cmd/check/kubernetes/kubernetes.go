package kubernetes

import (
	"fmt"
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/cobra"
	"golang.org/x/xerrors"

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

	"github.com/cdr/coder-doctor/internal/api"
	"github.com/cdr/coder-doctor/internal/checks/kube"
	"github.com/cdr/coder-doctor/internal/checks/local"
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
		return nil, xerrors.Errorf("parse %s: %w", clientcmd.FlagContext, err)
	}

	overrides.Context.Namespace, err = cmd.Flags().GetString(clientcmd.FlagNamespace)
	if err != nil {
		return nil, xerrors.Errorf("parse %s: %w", clientcmd.FlagNamespace, err)
	}

	overrides.Context.Cluster, err = cmd.Flags().GetString(clientcmd.FlagClusterName)
	if err != nil {
		return nil, xerrors.Errorf("parse %s: %w", clientcmd.FlagClusterName, err)
	}

	return overrides, nil
}

func run(cmd *cobra.Command, _ []string) error {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	var err error
	loadingRules.ExplicitPath, err = cmd.Flags().GetString(clientcmd.RecommendedConfigPathFlag)
	if err != nil {
		return xerrors.Errorf("parse %s: %w", clientcmd.RecommendedConfigPathFlag, err)
	}

	overrides, err := getConfigOverridesFromFlags(cmd)
	if err != nil {
		return xerrors.Errorf("parse flags: %w", err)
	}

	configLoader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)
	config, err := configLoader.ClientConfig()
	if err != nil {
		return xerrors.Errorf("creating NonInteractiveDeferredLoadingClientConfig: %w", err)
	}

	rawConfig, err := configLoader.RawConfig()
	if err != nil {
		return xerrors.Errorf("creating RawConfig: %w", err)
	}

	clientset, err := kclient.NewForConfig(config)
	if err != nil {
		return xerrors.Errorf("creating kube client from config: %w", err)
	}

	coderVersion, err := cmd.Flags().GetString("coder-version")
	if err != nil {
		return xerrors.Errorf("parse coder-version string: %w", err)
	}

	cv, err := semver.NewVersion(coderVersion)
	if err != nil {
		return xerrors.Errorf("parse coder-version from string %q: %w", coderVersion, err)
	}

	log := slog.Make(sloghuman.Sink(cmd.OutOrStdout()))
	verbosity, err := cmd.Flags().GetInt("verbosity")
	if err != nil {
		return xerrors.Errorf("parse verbosity: %w", err)
	}

	// TODO: this is pretty arbitrary, use a defined verbosity similar to
	// kubectl
	if verbosity > 5 {
		log = log.Leveled(slog.LevelDebug)
	}

	currentContext := rawConfig.Contexts[rawConfig.CurrentContext]
	if currentContext.Namespace == "" {
		currentContext.Namespace = "default"
	}

	colorFlag, err := cmd.Flags().GetBool("color")
	if err != nil {
		return xerrors.Errorf("parse color: %w", err)
	}

	asciiFlag, err := cmd.Flags().GetBool("ascii")
	if err != nil {
		return xerrors.Errorf("parse ascii: %w", err)
	}

	outputMode := humanwriter.OutputModeEmoji
	if asciiFlag {
		outputMode = humanwriter.OutputModeText
	}

	var writer api.ResultWriter = humanwriter.New(
		os.Stdout,
		humanwriter.WithColors(colorFlag),
		humanwriter.WithMode(outputMode),
	)

	localChecker := local.NewChecker(
		local.WithLogger(log),
		local.WithCoderVersion(cv),
		local.WithWriter(writer),
		local.WithTarget(api.CheckTargetKubernetes),
	)

	kubeChecker := kube.NewKubernetesChecker(
		clientset,
		kube.WithLogger(log),
		kube.WithCoderVersion(cv),
		kube.WithWriter(writer),
		kube.WithNamespace(currentContext.Namespace),
	)

	_ = writer.WriteResult(&api.CheckResult{
		Name:    "kubernetes current-context",
		State:   api.StateInfo,
		Summary: fmt.Sprintf("kube context: %q", rawConfig.CurrentContext),
		Details: map[string]interface{}{
			"current-context": rawConfig.CurrentContext,
			"cluster":         currentContext.Cluster,
			"namespace":       currentContext.Namespace,
			"user":            currentContext.AuthInfo,
		},
	})

	if err := localChecker.Validate(); err != nil {
		return xerrors.Errorf("failed to validate local checks: %w", err)
	}

	if err := localChecker.Run(cmd.Context()); err != nil {
		return xerrors.Errorf("run local checker: %w", err)
	}

	if err := kubeChecker.Validate(); err != nil {
		return xerrors.Errorf("failed to validate kube checker: %w", err)
	}

	if err := kubeChecker.Run(cmd.Context()); err != nil {
		return xerrors.Errorf("run kube checker: %w", err)
	}

	return nil
}
