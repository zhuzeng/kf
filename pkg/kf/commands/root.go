package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	serving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/tools/clientcmd"
)

// KfParams stores everything needed to interact with the user and Knative.
type KfParams struct {
	Output         io.Writer
	ServingFactory func() (serving.ServingV1alpha1Interface, error)
	Namespace      string
}

var kubeCfgFile string

func GetConfig() (serving.ServingV1alpha1Interface, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeCfgFile)
	if err != nil {
		return nil, err
	}
	client, err := serving.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// NewKfCommand creates the root kf command.
func NewKfCommand() *cobra.Command {
	p := &KfParams{
		Output:         os.Stdout,
		ServingFactory: GetConfig,
	}

	var rootCmd = &cobra.Command{
		Use:   "kf",
		Short: "kf is like cf for Knative",
		Long: `kf is like cf for Knative

	Supported sub-commands are:

	  kf push
	  kf delete <app>
	  kf apps

	You can get more info by adding the --help flag to any sub-command.
	  `,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}

	// Model new commands after:
	// https://github.com/knative/client/blob/master/pkg/kn/commands/service_list.go
	// to take an idiomatic k8s like approach.
	rootCmd.PersistentFlags().StringVar(&kubeCfgFile, "kubeconfig", "", "kubectl config file (default is $HOME/.kube/config)")
	rootCmd.PersistentFlags().StringVar(&p.Namespace, "namespace", "", "namespace")
	rootCmd.AddCommand(NewPushCommand(p))
	rootCmd.AddCommand(NewAppsCommand(p))

	return rootCmd
}

func InitializeConfig() {
	cobra.OnInitialize(initKubeConfig)
}

func initKubeConfig() {
	if kubeCfgFile == "" {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		kubeCfgFile = filepath.Join(home, ".kube", "config")
	}
}