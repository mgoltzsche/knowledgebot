package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	rootCmd = &cobra.Command{
		Use:               "knowledgebot",
		Short:             "An AI app to answer questions about an indexed body of knowledge",
		Long:              `A RAG AI app to answer questions about a body of knowledge that is indexed within a Qdrant vector database.`,
		SilenceErrors:     true,
		SilenceUsage:      true,
		PersistentPreRunE: preRunEnvVars,
	}
)

func init() {
	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		_ = cmd.Help()
		return err
	})
}

func Execute() error {
	return rootCmd.Execute()
}

func preRunEnvVars(cmd *cobra.Command, args []string) error {
	return parseFlagsWithEnvVars(cmd.Flags(), "KLB_")
}

func parseFlagsWithEnvVars(fs *pflag.FlagSet, envVarPrefix string) error {
	var err error

	fs.VisitAll(func(f *pflag.Flag) {
		envVarName := envVarPrefix + strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
		if envVarValue := os.Getenv(envVarName); envVarValue != "" && !f.Changed {
			e := f.Value.Set(envVarValue)
			if e != nil && err == nil {
				err = fmt.Errorf("invalid environment variable %s value provided: %s", envVarValue, err)
			}
		}
	})

	return err
}
