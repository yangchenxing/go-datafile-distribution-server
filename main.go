package main

import (
	"flag"
	"github.com/yangchenxing/go-logging"
	"github.com/yangchenxing/go-toml2struct"
	"golang.org/x/net/http2"
	"net/http"
	"os"
)

type serverConfig struct {
	Logging    *logging.Config
	ListenAddr string
	DataPath   string
	Backups    uint
	TLS        bool
	CertFile   string
	KeyFile    string
}

var (
	config = serverConfig{
		ListenAddr: "0.0.0.0:80",
		DataPath:   "data",
		Backups:    5,
		TLS:        false,
		CertFile:   "conf/server.crt",
		KeyFile:    "conf/server.key",
	}
	configPath = flag.String("config", "conf/config.toml", "配置文件路径")
)

func main() {
	if err := toml2struct.Load(*configPath, "include", &config); err != nil {
		logging.Fatal("load config file fail: %s", err.Error())
		os.Exit(1)
	}
	server := &server{
		dataPath:   config.DataPath,
		backupPath: config.BackupPath,
		backups:    config.Backups,
		files:      make(map[string]*file),
	}
	httpServer := &http.Server{
		Addr:    config.ListenAddr,
		Handler: server,
	}
	http2.ConfigureServer(httpServer, &http2.Server{})
	if config.TLS {
		httpServer.ListenAndServeTLS(config.CertFile, config.KeyFile)
	} else {
		httpServer.ListenAndServe()
	}
}
