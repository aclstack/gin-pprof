package config

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"gopkg.in/yaml.v3"
	"github.com/aclstack/gin-pprof/pkg/core"
)

// NacosConfig implements ConfigProvider interface using Nacos
type NacosConfig struct {
	serverAddr string
	namespace  string
	group      string
	dataID     string
	username   string
	password   string
	client     config_client.IConfigClient
	logger     core.Logger
}

// NacosOptions contains options for Nacos configuration
type NacosOptions struct {
	ServerAddr string
	Namespace  string
	Group      string
	DataID     string
	Username   string
	Password   string
}

// NewNacosConfig creates a new NacosConfig
func NewNacosConfig(opts NacosOptions, logger core.Logger) (core.ConfigProvider, error) {
	nc := &NacosConfig{
		serverAddr: opts.ServerAddr,
		namespace:  opts.Namespace,
		group:      opts.Group,
		dataID:     opts.DataID,
		username:   opts.Username,
		password:   opts.Password,
		logger:     logger,
	}

	// Set defaults
	if nc.namespace == "" {
		nc.namespace = "public"
	}
	if nc.group == "" {
		nc.group = "DEFAULT_GROUP"
	}
	if nc.dataID == "" {
		nc.dataID = "gin-pprof.yaml"
	}

	// Parse server address
	parts := strings.Split(nc.serverAddr, ":")
	if len(parts) != 2 {
		logger.Error("Invalid Nacos server address", map[string]interface{}{
			"addr": nc.serverAddr,
		})
		return nil, &ConfigError{Message: "Invalid server address format"}
	}

	host := parts[0]
	port, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		logger.Error("Invalid Nacos server port", map[string]interface{}{
			"addr":  nc.serverAddr,
			"error": err.Error(),
		})
		return nil, err
	}

	// Create Nacos client configuration
	serverConfigs := []constant.ServerConfig{
		*constant.NewServerConfig(host, port),
	}

	clientConfig := &constant.ClientConfig{
		NamespaceId:         nc.namespace,
		TimeoutMs:           10000,
		NotLoadCacheAtStart: true,
		LogLevel:            "warn",
		LogDir:              "./logs/nacos",
		CacheDir:            "./cache/nacos",
		Username:            nc.username,
		Password:            nc.password,
	}

	// Create config client
	client, err := clients.NewConfigClient(vo.NacosClientParam{
		ClientConfig:  clientConfig,
		ServerConfigs: serverConfigs,
	})
	if err != nil {
		logger.Error("Failed to create Nacos client", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	nc.client = client
	logger.Info("Nacos config provider initialized", map[string]interface{}{
		"host":      host,
		"port":      port,
		"namespace": nc.namespace,
		"group":     nc.group,
		"data_id":   nc.dataID,
	})

	return nc, nil
}

// GetTasks returns current profiling tasks from Nacos
func (n *NacosConfig) GetTasks(ctx context.Context) ([]core.ProfilingTask, error) {
	content, err := n.client.GetConfig(vo.ConfigParam{
		DataId: n.dataID,
		Group:  n.group,
	})
	if err != nil {
		n.logger.Error("Failed to get config from Nacos", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	if content == "" {
		n.logger.Info("Empty config from Nacos", nil)
		return []core.ProfilingTask{}, nil
	}

	return n.parseConfig(content)
}

// Subscribe subscribes to configuration changes from Nacos
func (n *NacosConfig) Subscribe(ctx context.Context, callback func([]core.ProfilingTask)) error {
	onChangeCallback := func(namespace, group, dataId, data string) {
		n.logger.Info("Nacos config changed", map[string]interface{}{
			"namespace": namespace,
			"group":     group,
			"data_id":   dataId,
		})

		tasks, err := n.parseConfig(data)
		if err != nil {
			n.logger.Error("Failed to parse changed config", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		callback(tasks)
	}

	err := n.client.ListenConfig(vo.ConfigParam{
		DataId:   n.dataID,
		Group:    n.group,
		OnChange: onChangeCallback,
	})
	if err != nil {
		n.logger.Error("Failed to subscribe to Nacos config", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	n.logger.Info("Subscribed to Nacos config changes", map[string]interface{}{
		"data_id": n.dataID,
		"group":   n.group,
	})

	return nil
}

// Close closes the Nacos config provider
func (n *NacosConfig) Close() error {
	n.logger.Info("Nacos config provider closed", nil)
	return nil
}

// parseConfig parses the configuration data
func (n *NacosConfig) parseConfig(data string) ([]core.ProfilingTask, error) {
	if data == "" {
		return []core.ProfilingTask{}, nil
	}

	// Try enhanced format first
	var enhancedConfig struct {
		Profiles []core.ProfilingTask `yaml:"profiles"`
	}

	if err := yaml.Unmarshal([]byte(data), &enhancedConfig); err == nil && len(enhancedConfig.Profiles) > 0 {
		return n.processEnhancedConfig(enhancedConfig.Profiles), nil
	}

	// Try simple format (backward compatibility)
	var simpleConfig map[string]string
	if err := yaml.Unmarshal([]byte(data), &simpleConfig); err != nil {
		n.logger.Error("Failed to parse config", map[string]interface{}{
			"error": err.Error(),
			"data":  data,
		})
		return nil, err
	}

	return n.processSimpleConfig(simpleConfig), nil
}

// processEnhancedConfig processes enhanced format configuration
func (n *NacosConfig) processEnhancedConfig(profiles []core.ProfilingTask) []core.ProfilingTask {
	var validTasks []core.ProfilingTask
	now := time.Now()

	for _, profile := range profiles {
		// Set default values
		if profile.Duration == 0 {
			profile.Duration = 30
		}
		if profile.SampleRate == 0 {
			profile.SampleRate = 1
		}
		if profile.ProfileType == "" {
			profile.ProfileType = "cpu"
		}
		// Set default method - only set if both method and methods are empty
		if profile.Method == "" && len(profile.Methods) == 0 {
			profile.Method = "GET"
		}

		// Check if expired
		if now.After(profile.ExpiresAt) {
			n.logger.Warn("Task expired", map[string]interface{}{
				"path":       profile.Path,
				"expires_at": profile.ExpiresAt.Format(time.RFC3339),
			})
			continue
		}

		validTasks = append(validTasks, profile)
	}

	n.logger.Info("Enhanced config processed", map[string]interface{}{
		"total_tasks": len(profiles),
		"valid_tasks": len(validTasks),
	})

	return validTasks
}

// processSimpleConfig processes simple format configuration
func (n *NacosConfig) processSimpleConfig(rawTasks map[string]string) []core.ProfilingTask {
	var validTasks []core.ProfilingTask
	now := time.Now()

	for path, expiresAtStr := range rawTasks {
		expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
		if err != nil {
			n.logger.Warn("Invalid time format", map[string]interface{}{
				"path":       path,
				"expires_at": expiresAtStr,
				"error":      err.Error(),
			})
			continue
		}

		// Check if expired
		if now.After(expiresAt) {
			n.logger.Warn("Task expired", map[string]interface{}{
				"path":       path,
				"expires_at": expiresAtStr,
			})
			continue
		}

		validTasks = append(validTasks, core.ProfilingTask{
			Path:        path,
			Method:      "GET",  // Default GET method
			Methods:     nil,    // Empty methods array
			ExpiresAt:   expiresAt,
			Duration:    30,     // Default 30 seconds
			SampleRate:  1,      // Default no sampling
			ProfileType: "cpu",  // Default CPU profiling
		})
	}

	n.logger.Info("Simple config processed", map[string]interface{}{
		"total_tasks": len(rawTasks),
		"valid_tasks": len(validTasks),
	})

	return validTasks
}

// ConfigError represents a configuration error
type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}