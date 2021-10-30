package play

import (
    "encoding/json"
    "github.com/bytelang/kplayer/module"
    playprovider "github.com/bytelang/kplayer/module/play/provider"
    "github.com/bytelang/kplayer/module/resource/provider"
    kptypes "github.com/bytelang/kplayer/types"
    "github.com/bytelang/kplayer/types/config"
    kpproto "github.com/bytelang/kplayer/types/core/proto"
    "github.com/spf13/cobra"
)

type AppModule struct {
    *provider.Provider
}

var _ module.AppModule = &AppModule{}

func NewAppModule(playProvider playprovider.ProviderI) AppModule {
    return AppModule{provider.NewProvider(playProvider)}
}

func (m AppModule) GetModuleName() string {
    return provider.ModuleName
}

func (m AppModule) GetCommand() *cobra.Command {
    return provider.GetCommand()
}

func (m AppModule) InitConfig(ctx *kptypes.ClientContext, data json.RawMessage) error {
    var cfg config.Resource
    if err := json.Unmarshal(data, &cfg); err != nil {
        return err
    }

    m.InitModule(ctx, cfg)

    return nil
}

func (m AppModule) ValidateConfig() error {
    return m.Provider.ValidateConfig()
}

func (m AppModule) TriggerMessage(message *kpproto.KPMessage) {
    m.Trigger(message)
}
