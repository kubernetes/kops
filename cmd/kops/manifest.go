package main

import (
	"github.com/spf13/viper"
	"path/filepath"
	"strings"
	"fmt"
)

// Will return a Viper configuration, or an err based on a potential yml config file.
func getManifest(manifest string) (*viper.Viper, error) {
	mDir := filepath.Dir(manifest)
	mFile := filepath.Base(manifest)
	mExt := filepath.Ext(manifest)
	mName := strings.Replace(mFile, mExt, "", -1)
	clusterConfig := viper.New()
	clusterConfig.SetConfigType("yaml")
	clusterConfig.SetConfigName(mName)
	clusterConfig.AddConfigPath(mDir)
	err := clusterConfig.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("Unable to parse manifest %s", err.Error())
	}
	return clusterConfig, nil
}
