package user

import (
	pb "StealthIMMSAP/StealthIM.User"
	"StealthIMMSAP/config"
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var conns []*grpc.ClientConn
var mainlock sync.RWMutex

func createConn(connID int) {
	log.Printf("[User]Connect %d", connID+1)
	conn, err := grpc.NewClient(fmt.Sprintf("%s:%d", config.LatestConfig.User.Host, config.LatestConfig.User.Port),
		grpc.WithTransportCredentials(
			insecure.NewCredentials()))
	if conn == nil {
		log.Printf("[User]Connect %d Error %v\n", connID+1, err)
		conns[connID] = nil
		return
	}
	if err != nil {
		log.Printf("[User]Connect %d Error %v\n", connID+1, err)
		conns[connID] = nil
		return
	}
	conns[connID] = conn
}

func checkAlive(connID int) {
	if len(conns) <= connID {
		return
	}
	for {
		if len(conns) <= connID {
			return
		}
		mainlock.RLock()
		if conns[connID] != nil {
			cli := pb.NewStealthIMUserClient(conns[connID])
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			_, err := cli.Ping(ctx, &pb.PingRequest{})
			cancel()
			if err == nil {
				mainlock.RUnlock()
				continue
			}
		}
		createConn(connID)
		mainlock.RUnlock()
		time.Sleep(5 * time.Second)
	}
}

// InitConns 扩缩容连接
func InitConns() {
	defer func() {
		mainlock.Lock()
		for _, conn := range conns {
			conn.Close()
		}
		mainlock.Unlock()
	}()
	log.Printf("[User]Init Conns\n")
	for {
		time.Sleep(time.Second * 1)
		var lenTmp = len(conns)
		if lenTmp < config.LatestConfig.User.ConnNum {
			log.Printf("[User]Create Conn %d\n", lenTmp+1)
			mainlock.Lock()
			conns = append(conns, nil)
			mainlock.Unlock()
			go checkAlive(lenTmp)
		} else if lenTmp > config.LatestConfig.User.ConnNum {
			log.Printf("[User]Delete Conn %d\n", lenTmp)
			mainlock.Lock()
			conns[lenTmp-1].Close()
			conns = conns[:lenTmp-1]
			mainlock.Unlock()
		} else {
			time.Sleep(time.Second * 5)
		}
	}
}
