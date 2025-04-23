package main

import (
	"bank_system/server"

	"github.com/spf13/viper"
)

var port, certFile, keyFile string

func init() {
	viper.AddConfigPath("./configs")
	viper.SetConfigName("config")
	viper.SetConfigType("json")

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	port = viper.GetString("server.port")
	certFile = viper.GetString("certs.path.cert")
	keyFile = viper.GetString("certs.path.key")
}

func main() {
	server, err := server.NewServer()
	if err != nil {
		panic(err)
	}
	if err := server.Start(port, certFile, keyFile); err != nil {
		panic(err)
	}

	defer server.Stop()
}
