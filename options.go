package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// Options defines the configuration which the application will run with. Options can be set as environment variables, flag or cfg.
// Is not mandatory all of them at once.
type Options struct {
	FailoverIps    []string `flag:"failover-ips" cfg:"failover_ips" env:"FAILOVER_IPS"`
	AssetsPath     string   `flag:"assets" cfg:"assets_path" env:"ASSETS_PATH"`
	KubernetesPath string   `flag:"kubernetes" cfg:"kubernetes_path" env:"KUBERNETES_PATH"`
	KubeConfig     string   `flag:"kubeconfig" cfg:"kubeconfig" env:"KUBECONFIG"`
}

// EnvOptions defines options through environment variable.
type EnvOptions map[string]interface{}

// LoadEnvForStruct set options via environment variable
func (cfg EnvOptions) LoadEnvForStruct(options interface{}) {
	val := reflect.ValueOf(options).Elem()
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		// pull out the struct tags:
		//    flag - the name of the command line flag
		//    deprecated - (optional) the name of the deprecated command line flag
		//    cfg - (optional, defaults to underscored flag) the name of the config file option
		field := typ.Field(i)
		flagName := field.Tag.Get("flag")
		envName := field.Tag.Get("env")
		cfgName := field.Tag.Get("cfg")
		if cfgName == "" && flagName != "" {
			cfgName = strings.Replace(flagName, "-", "_", -1)
		}
		if envName == "" || cfgName == "" {
			// resolvable fields must have the `env` and `cfg` struct tag
			continue
		}
		v := os.Getenv(envName)
		if v != "" {
			cfg[cfgName] = v
		}
	}
}

// Validate function applies validations towards the options (flags/environment variables/config).
func (o *Options) Validate() error {
	if h := os.Getenv("HOME"); h != "" && opts.KubeConfig == "" {
		opts.KubeConfig = filepath.Join(h, ".kube", "config")
	}
	fmt.Printf("%v", opts.FailoverIps)
	return nil
}

// NewOptions returns default *Options.
func NewOptions() *Options {
	return &Options{}
}
