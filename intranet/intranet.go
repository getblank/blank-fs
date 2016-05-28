package intranet

import (
	"encoding/json"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/getblank/wango"
)

var (
	srAddress string
	tqAddress string
	srClient  *wango.Wango
	srLocker  sync.RWMutex
	tqClient  *wango.Wango
	tqLocker  sync.RWMutex
)

// Init is the main entry point for the intranet package
func Init(addr string) {
	srAddress = addr
	go connectToTaskQ()
	connectToSr()
}

type service struct {
	Type    string `json:"type"`
	Address string `json:"address"`
	Port    string `json:"port"`
}

func connectedToSR(w *wango.Wango) {
	log.Info("Connected to SR: ", srAddress)
	srLocker.Lock()
	srClient = w
	srLocker.Unlock()

	srClient.Call("register", map[string]interface{}{"type": "fileStore", "port": "8082"})
}

func connectToSr() {
	reconnectChan := make(chan struct{})
	for {
		log.Info("Attempt to connect to SR: ", srAddress)
		client, err := wango.Connect(srAddress, "http://getblank.net")
		if err != nil {
			log.Warn("Can'c connect to service registry: " + err.Error())
			time.Sleep(time.Second)
			continue
		}
		client.Subscribe("registry", registryUpdateHandler)
		client.SetSessionCloseCallback(func(c *wango.Conn) {
			srLocker.Lock()
			srClient = nil
			srLocker.Unlock()
			reconnectChan <- struct{}{}
		})
		connectedToSR(client)
		<-reconnectChan
	}
}

func registryUpdateHandler(_ string, _event interface{}) {
	encoded, err := json.Marshal(_event)
	if err != nil {
		log.WithField("error", err).Error("Can't marshal registry update event")
		return
	}
	var services map[string][]service
	err = json.Unmarshal(encoded, &services)
	if err != nil {
		log.WithField("error", err).Error("Can't unmarshal registry update event to []Services")
		return
	}
	tqServices, ok := services["taskQueue"]
	if !ok {
		log.Warn("No taskq services in registry")
		return
	}
	var _tqAddress string
	for _, service := range tqServices {
		if tqAddress == service.Address+":"+service.Port {
			break
		}
		_tqAddress = service.Address + ":" + service.Port
	}
	if _tqAddress != "" {
		tqLocker.Lock()
		tqAddress = _tqAddress
		tqLocker.Unlock()
		reconnectToTaskQ()
	}
}

func connectedToTaskQ(w *wango.Wango) {
	log.Info("Connected to TaskQ: ", tqAddress)
	tqLocker.Lock()
	tqClient = w
	tqLocker.Unlock()
}

func connectToTaskQ() {
	for {
		tqLocker.RLock()
		addr := tqAddress
		tqLocker.RUnlock()
		if addr != "" {
			break
		}
	}
	reconnectChan := make(chan struct{})
	for {
		log.Info("Attempt to connect to TaskQ: ", tqAddress)
		client, err := wango.Connect(tqAddress, "http://getblank.net")
		if err != nil {
			log.Warn("Can'c connect to taskq: " + err.Error())
			time.Sleep(time.Second)
			continue
		}
		client.SetSessionCloseCallback(func(c *wango.Conn) {
			log.Warn("Disconnected from TaskQ")
			tqLocker.Lock()
			tqClient = nil
			tqLocker.Unlock()
			reconnectChan <- struct{}{}
		})
		connectedToTaskQ(client)
		<-reconnectChan
	}
}

func reconnectToTaskQ() {
	if tqClient != nil {
		tqClient.Disconnect()
	}
}
