package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// NacosConfig holds the bootstrap parameters for a Nacos client.
type NacosConfig struct {
	Addr        string
	Port        uint64
	NamespaceID string
	Group       string
	DataIDs     []string
	TimeoutMs   uint64
	LogLevel    string
	Username    string
	Password    string
}

// NacosSource reads configuration from a Nacos config centre.
// On startup it pulls every configured DataID and writes it to a local
// directory so that FileSource can serve the cached copy.  It also
// subscribes to Nacos change notifications and updates the local files.
type NacosSource struct {
	cfg    NacosConfig
	client config_client.IConfigClient
	local  *FileSource
	dir    string
	once   sync.Once
}

// NewNacosSource creates a NacosSource.  localDir is where pulled configs
// are written (e.g. "./configs").
func NewNacosSource(cfg NacosConfig, localDir string) (*NacosSource, error) {
	cc := constant.ClientConfig{
		TimeoutMs:   cfg.TimeoutMs,
		NamespaceId: cfg.NamespaceID,
		LogLevel:    cfg.LogLevel,
		Username:    cfg.Username,
		Password:    cfg.Password,
	}
	sc := []constant.ServerConfig{{
		IpAddr: cfg.Addr,
		Port:   cfg.Port,
	}}
	client, err := clients.NewConfigClient(vo.NacosClientParam{
		ClientConfig:  &cc,
		ServerConfigs: sc,
	})
	if err != nil {
		return nil, fmt.Errorf("nacos client: %w", err)
	}
	return &NacosSource{
		cfg:    cfg,
		client: client,
		local:  NewFileSource(localDir),
		dir:    localDir,
	}, nil
}

// PullAll fetches every configured DataID and writes it locally.
func (ns *NacosSource) PullAll() error {
	for _, dataID := range ns.cfg.DataIDs {
		content, err := ns.client.GetConfig(vo.ConfigParam{
			DataId: dataID,
			Group:  ns.cfg.Group,
		})
		if err != nil {
			return fmt.Errorf("nacos get %s: %w", dataID, err)
		}
		if err := ns.writeLocal(dataID, content); err != nil {
			return err
		}
	}
	return nil
}

// Read delegates to the local FileSource.
func (ns *NacosSource) Read(name string) ([]byte, error) { return ns.local.Read(name) }

// Format delegates.
func (ns *NacosSource) Format() string { return ns.local.Format() }

// Watch subscribes to Nacos change events and writes updates to local
// files.  The returned channel comes from the underlying FileSource so
// callers see changes from both Nacos and local edits.
func (ns *NacosSource) Watch() <-chan string {
	ns.once.Do(func() {
		for _, dataID := range ns.cfg.DataIDs {
			go func(id string) {
				_ = ns.client.ListenConfig(vo.ConfigParam{
					DataId: id,
					Group:  ns.cfg.Group,
					OnChange: func(namespace, group, dataId, data string) {
						_ = ns.writeLocal(dataId, data)
					},
				})
			}(dataID)
		}
	})
	return ns.local.Watch()
}

func (ns *NacosSource) writeLocal(dataID, content string) error {
	path := ns.dir + "/" + dataID + ".yaml"
	return os.WriteFile(path, []byte(content), 0o644)
}
