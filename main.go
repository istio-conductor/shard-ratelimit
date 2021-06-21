package main

import (
	"context"
	"errors"
	"github.com/istio-conductor/shard-ratelimit/misc/signals"
	"github.com/istio-conductor/shard-ratelimit/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	GrpcPort  int
	HTTPPort  int
	WatchDir  string
	LogLevel  string
	Replicas  int
	Namespace string
	Service   string
	ConfigMap string
)

var rootCmd = &cobra.Command{
	Use:   "ratelimit",
	Short: "A high performance ratelimit server.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if HTTPPort != 0 {
			go func() {

			}()
		}
		cmd.Flags().Visit(func(flag *pflag.Flag) {
			log.Info().Msgf("[%s]=%s", flag.Name, flag.Value.String())
		})
		ctx := signals.Context()
		s := server.New(GrpcPort, HTTPPort, WatchDir, Namespace, Service, ConfigMap, Replicas)
		err := s.Run(ctx)
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	},
}

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().IntVarP(&GrpcPort, "grpc", "p", 8081, "grpc listen port")
	rootCmd.PersistentFlags().IntVarP(&HTTPPort, "http", "d", 8080, "http listen port")
	rootCmd.PersistentFlags().IntVarP(&Replicas, "replicas", "r", 0, "replicas")

	rootCmd.PersistentFlags().StringVarP(&WatchDir, "watch", "w", "./configs", "watching directory")
	rootCmd.PersistentFlags().StringVarP(&LogLevel, "log_level", "l", "INFO", "watching directory")
	rootCmd.PersistentFlags().StringVarP(&Namespace, "namespace", "n", "istio-system", "namespace")
	rootCmd.PersistentFlags().StringVarP(&Service, "service", "s", "ratelimit", "service name")
	rootCmd.PersistentFlags().StringVarP(&ConfigMap, "configmap", "c", "", "configmap name")

}

func initConfig() {
	level, _ := zerolog.ParseLevel(LogLevel)
	if level == zerolog.NoLevel {
		level = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(level)
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		log.Err(err).Msg("process exit")
		return
	}
}
