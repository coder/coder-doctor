package kube

import (
	"context"

	"cdr.dev/slog"
	"cdr.dev/slog/sloggers/sloghuman"
	"github.com/Masterminds/semver/v3"
	"github.com/spf13/cobra"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/exec"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	_ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/cdr/coder-doctor/internal/api"
	"github.com/cdr/coder-doctor/internal/checks/kube"
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

	log := slog.Make(sloghuman.Sink(cmd.OutOrStdout())).Leveled(slog.LevelDebug)
	log.Debug(context.TODO(), "test message")

	checker := kube.NewKubernetesChecker(
		kube.WithClient(clientset),
		kube.WithCoderVersion(cv),
		kube.WithLogger(log),
	)
	results := checker.Run(cmd.Context())
	api.WriteResults(cmd.OutOrStdout(), results)

	return nil
}
