package cmd

import (
	"github.com/int128/kubelogin/pkg/adaptors/kubeconfig"
	"github.com/int128/kubelogin/pkg/adaptors/logger"
	"github.com/int128/kubelogin/pkg/usecases/standalone"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/xerrors"
)

const rootDescription = `Log in to the OpenID Connect provider.

You need to set up the OIDC provider, role binding, Kubernetes API server and kubeconfig.
To show the setup instruction:

	kubectl oidc-login setup

See https://github.com/int128/kubelogin for more.
`

// rootOptions represents the options for the root command.
type rootOptions struct {
	Kubeconfig            string
	Context               string
	User                  string
	tlsOptions            tlsOptions
	authenticationOptions authenticationOptions
}

func (o *rootOptions) addFlags(f *pflag.FlagSet) {
	f.StringVar(&o.Kubeconfig, "kubeconfig", "", "Path to the kubeconfig file")
	f.StringVar(&o.Context, "context", "", "Name of the kubeconfig context to use")
	f.StringVar(&o.User, "user", "", "Name of the kubeconfig user to use. Prior to --context")
	o.tlsOptions.addFlags(f)
	o.authenticationOptions.addFlags(f)
}

type Root struct {
	Standalone standalone.Interface
	Logger     logger.Interface
}

func (cmd *Root) New() *cobra.Command {
	var o rootOptions
	c := &cobra.Command{
		Use:   "kubelogin",
		Short: "Log in to the OpenID Connect provider",
		Long:  rootDescription,
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			grantOptionSet, err := o.authenticationOptions.grantOptionSet()
			if err != nil {
				return xerrors.Errorf("invalid option: %w", err)
			}
			in := standalone.Input{
				KubeconfigFilename: o.Kubeconfig,
				KubeconfigContext:  kubeconfig.ContextName(o.Context),
				KubeconfigUser:     kubeconfig.UserName(o.User),
				CACertFilename:     o.tlsOptions.CACertFilename,
				CACertData:         o.tlsOptions.CACertData,
				SkipTLSVerify:      o.tlsOptions.SkipTLSVerify,
				GrantOptionSet:     grantOptionSet,
			}
			if err := cmd.Standalone.Do(c.Context(), in); err != nil {
				return xerrors.Errorf("login: %w", err)
			}
			return nil
		},
	}
	c.Flags().SortFlags = false
	o.addFlags(c.Flags())
	cmd.Logger.AddFlags(c.PersistentFlags())
	return c
}
