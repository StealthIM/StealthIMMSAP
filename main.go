package main

import (
	"StealthIMMSAP/config"
	"StealthIMMSAP/gateway"
	"StealthIMMSAP/nats"
	"StealthIMMSAP/reader"
	"StealthIMMSAP/sender"
	"StealthIMMSAP/spliter"
	"StealthIMMSAP/subscriber"
	"StealthIMMSAP/user"
	"log"
	"os"
)

func main() {
	cfg := config.ReadConf()
	log.Printf("Start server [%v]\n", config.Version)
	log.Printf("+ GRPC\n")
	log.Printf("    Host: %s\n", cfg.SendGRPCProxy.Host)
	log.Printf("    Port: %d\n", cfg.SendGRPCProxy.Port)
	log.Printf("+ DBGateway\n")
	log.Printf("    Host: %s\n", cfg.DBGateway.Host)
	log.Printf("    Port: %d\n", cfg.DBGateway.Port)
	log.Printf("    ConnNum: %d\n", cfg.DBGateway.ConnNum)
	log.Printf("+ User\n")
	log.Printf("    Host: %s\n", cfg.User.Host)
	log.Printf("    Port: %d\n", cfg.User.Port)
	log.Printf("    ConnNum: %d\n", cfg.User.ConnNum)
	log.Printf("+ NATS\n")
	log.Printf("    Host: %s\n", cfg.Nats.Host)
	log.Printf("    Port: %d\n", cfg.Nats.Port)
	log.Printf("    Username: %s\n", cfg.Nats.Username)
	log.Printf("    Password: **********\n")
	log.Printf("    ConnNum: %d\n", cfg.Nats.ConnNum)

	// 启动 DBGateway
	go gateway.InitConns()

	mode := os.Getenv("StealthIMMSAP_MODE")
	if mode == "" || mode == "grpc" {
		// 启动 GRPC 服务
		go reader.Start(cfg)
	}
	if mode == "" || mode == "subscriber" {
		// 启动消息服务
		subscriber.Init(cfg)
	}
	if mode == "" || mode == "spliter" {
		// 启动消息服务
		go spliter.Init(cfg)
	}
	if mode == "" || mode == "sender" {
		// 启动消息服务
		sender.Init(cfg)
		go user.InitConns()
		go sender.Start(cfg)
	}

	// 启动 NATS
	go nats.InitConns()

	select {}
}
