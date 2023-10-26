package agent

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	command "github.com/rancher/fleet/internal/cmd"
	"github.com/rancher/fleet/internal/cmd/agent/register"
	"github.com/rancher/wrangler/v2/pkg/kubeconfig"
)

func NewRegister() *cobra.Command {
	cmd := command.Command(&Register{}, cobra.Command{
		Use:   "register [flags]",
		Short: "Register agent with an upstream cluster",
	})
	return cmd
}

type Register struct {
	UpstreamOptions
}

func (r *Register) Run(cmd *cobra.Command, args []string) error {
	// provide a logger in the context to be compatible with controller-runtime
	zopts := zap.Options{
		Development: true,
	}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zopts)))
	ctx := log.IntoContext(cmd.Context(), ctrl.Log)

	clientConfig := kubeconfig.GetNonInteractiveClientConfig(r.Kubeconfig)
	kc, err := clientConfig.ClientConfig()
	if err != nil {
		return err
	}

	setupLog.Info("starting registration on upstream cluster", "namespace", r.Namespace)

	ctx, cancel := context.WithCancel(ctx)
	// try to register with upstream fleet controller by obtaining
	// a kubeconfig for the upstream cluster
	agentInfo, err := register.Register(ctx, r.Namespace, kc)
	if err != nil {
		logrus.Fatal(err)
	}

	ns, _, err := agentInfo.ClientConfig.Namespace()
	if err != nil {
		logrus.Fatal(err)
	}

	_, err = agentInfo.ClientConfig.ClientConfig()
	if err != nil {
		logrus.Fatal(err)
	}

	setupLog.Info("successfully registered with upstream cluster", "namespace", ns)
	cancel()

	return nil
}