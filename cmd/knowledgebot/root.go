package main

import (
	"fmt"
	"log/slog"
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
	rootCmd.PersistentFlags().Var(logLevelFlag("INFO"), "log-level", "set the log level")
	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		_ = cmd.Help()
		return err
	})
}

func Execute() error {
	return rootCmd.Execute()
}

func preRunEnvVars(cmd *cobra.Command, args []string) error {
	return applyEnvVarsToFlags(cmd.Flags(), "KLB_")
}

func applyEnvVarsToFlags(fs *pflag.FlagSet, envVarPrefix string) error {
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

type logLevelFlag string

func (logLevelFlag) Set(s string) error {
	var level slog.Level

	switch s {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		return fmt.Errorf("unsupported log level %q provided. supported log levels are DEBUG, INFO, WARN, ERROR", s)
	}

	slog.SetLogLoggerLevel(level)

	return nil
}

func (f logLevelFlag) String() string {
	return string(f)
}

func (f logLevelFlag) Type() string {
	return "LEVEL"
}
