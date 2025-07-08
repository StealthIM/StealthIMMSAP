package nats

import (
	"StealthIMMSAP/config"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

// Conn 连接对象
type Conn struct {
	Conn      *nats.Conn
	JetStream *nats.JetStreamContext
}

var conns []*Conn
var mainlock sync.RWMutex

var waitSubscribers []*func(*Conn) error

func createConn(connID int) {
	log.Printf("[NATS]Connect %d", connID+1)
	conn, err := nats.Connect(
		fmt.Sprintf("nats://%s:%d", config.LatestConfig.Nats.Host, config.LatestConfig.Nats.Port),
		nats.UserInfo(config.LatestConfig.Nats.Username, config.LatestConfig.Nats.Password))
	if conn == nil {
		log.Printf("[NATS]Connect %d Error %v on conn\n", connID+1, err)
		conns[connID] = nil
		return
	}
	if err != nil {
		log.Printf("[NATS]Connect %d Error %v on conn err\n", connID+1, err)
		conns[connID] = nil
		return
	}
	js, err := conn.JetStream()
	if err != nil {
		log.Printf("[NATS]Connect %d Error %v on create stream\n", connID+1, err)
		conns[connID] = nil
	}
	connObj := &Conn{Conn: conn, JetStream: &js}

	for i, v := range streamTable {
		_, err := js.AddStream(&v)
		if err != nil {
			log.Printf("[NATS]Connect %d Error %v on create stream %d\n", connID+1, err, i)
		}
	}

	go func() {
		for _, subscriber := range waitSubscribers {
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("[NATS]Panic in subscriber callback: %v\n", r)
					}
				}()
				err := (*subscriber)(connObj)
				if err != nil {
					log.Printf("[NATS]Error in subscriber callback: %v\n", err)
				}
			}()
		}
	}()

	conns[connID] = connObj
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
			if conns[connID].Conn.IsConnected() {
				mainlock.RUnlock()
				time.Sleep(5 * time.Second)
				continue
			}
		}
		createConn(connID)
		mainlock.RUnlock()
		time.Sleep(10 * time.Second)
	}
}

// InitConns 扩缩容连接
func InitConns() {
	defer func() {
		mainlock.Lock()
		for _, conn := range conns {
			conn.Conn.Close()
		}
		mainlock.Unlock()
	}()
	log.Printf("[NATS]Init Conns\n")
	for {
		time.Sleep(time.Second * 1)
		var lenTmp = len(conns)
		if lenTmp < config.LatestConfig.Nats.ConnNum {
			log.Printf("[NATS]Create Conn %d\n", lenTmp+1)
			mainlock.Lock()
			conns = append(conns, nil)
			go checkAlive(lenTmp)
			mainlock.Unlock()
			time.Sleep(time.Millisecond * 100)
		} else if lenTmp > config.LatestConfig.Nats.ConnNum {
			log.Printf("[NATS]Delete Conn %d\n", lenTmp)
			mainlock.Lock()
			conns[lenTmp-1].Conn.Close()
			conns = conns[:lenTmp-1]
			mainlock.Unlock()
		} else {
			time.Sleep(time.Second * 5)
		}
	}
}

// chooseConn 随机选择链接
func chooseConn() (*Conn, error) {
	if len(conns) == 0 {
		return nil, errors.New("No available connections")
	}
	for range 10 {
		conntmp := conns[rand.Intn(len(conns))]
		if conntmp != nil {
			return conntmp, nil
		}
	}
	return nil, errors.New("No available connections")
}
