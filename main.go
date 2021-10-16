package main

import (
    "context"
    "encoding/json"
    "github.com/bytelang/kplayer/module"
    "github.com/spf13/viper"
    "os"

    "github.com/bytelang/kplayer/app"
    "github.com/bytelang/kplayer/cmd"
    kptypes "github.com/bytelang/kplayer/types"
    "github.com/rs/zerolog"
    log "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
)

func init() {
    log.SetOutput(os.Stdout)
    log.SetReportCaller(true)
    log.SetLevel(log.TraceLevel)
    logFormat := &log.TextFormatter{}
    log.SetFormatter(logFormat)
}

func main() {
    rootCmd := cmd.NewRootCmd()

    if err := Execute(rootCmd, app.DefaultConfigFilePath, app.DefaultConfigFileName); err != nil {
        switch e := err.(type) {
        case kptypes.ErrorCode:
            os.Exit(e.Code)
        default:
            os.Exit(1)
        }
    }
}

// Execute execute from flags and commands
func Execute(rootCmd *cobra.Command, defaultHome string, defaultFile string) error {
    rootCmd.PersistentFlags().String("log_level", zerolog.InfoLevel.String(), "The logging level (trace|debug|info|warn|error|fatal|panic)")
    rootCmd.PersistentFlags().String("log_format", "plain", "The logging format (json|plain)")
    rootCmd.PersistentFlags().StringP(kptypes.FlagHome, "", defaultHome, "directory for config and data")
    rootCmd.PersistentFlags().StringP(kptypes.FlagConfigFileName, "", defaultFile, "config file name")
    rootCmd.PersistentFlags().Bool("trace", false, "print out full stack trace on errors")
    rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
        if cmd.Parent().Use != "init" {
            InitGlobalConfig(cmd)
        }
        return nil
    }

    ctx := context.Background()
    ctx = context.WithValue(ctx, kptypes.ClientContextKey, kptypes.DefaultClientContext())
    ctx = context.WithValue(ctx, kptypes.ModuleManagerContextKey, app.ModuleManager)

    return kptypes.SetCommandContextAndExecute(rootCmd, ctx)
}

func InitGlobalConfig(cmd *cobra.Command) {
    mm := cmd.Context().Value(kptypes.ModuleManagerContextKey).(module.ModuleManager)
    clientCtx := cmd.Context().Value(kptypes.ClientContextKey).(*kptypes.ClientContext)

    home, err := app.GetHome(cmd)
    if err != nil {
        log.Fatal(err)
    }
    configFileName, err := app.GetConfigFileName(cmd)
    if err != nil {
        log.Fatal(err)
    }

    v := viper.New()
    v.AddConfigPath(home)
    v.SetConfigType("json")
    v.SetConfigName(configFileName)
    if err := v.ReadInConfig(); err != nil {
        log.Fatal(err)
    }

    clientCtx.Viper = v
    if err := v.Unmarshal(clientCtx.Config); err != nil {
        log.Fatal(err)
    }

    for _, item := range mm {
        d, err := json.Marshal(v.Get(item.GetModuleName()))
        if err != nil {
            log.Fatal(err)
        }
        if err := item.InitConfig(clientCtx, d); err != nil {
            log.Fatal(err)
        }
        if err := item.ValidateConfig(); err != nil {
            log.Fatal(err)
        }
    }
}
