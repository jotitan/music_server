package music

import (
	"github.com/jotitan/music_server/logger"
	"net/http"
)

// Manage

type Devices struct {
	// Set true for an ip if a remote server to control volume exist (port 9098 by default)
	remoteServers map[string]bool
}

func NewDevices() *Devices {
	return &Devices{make(map[string]bool)}
}

func (d *Devices) Reset() {
	d.remoteServers = make(map[string]bool)
}

func (d *Devices) SetVolume(volumeDown bool, host string) {
	volume := "volumeUp"
	if volumeDown {
		volume = "volumeDown"
	}
	// Check if service it's not already check or it's true
	if serverRunning, exist := d.remoteServers[host]; serverRunning || !exist {
		if _, err := http.Get("http://" + host + ":9098/" + volume); err != nil {
			// Close it
			d.remoteServers[host] = false
			logger.GetLogger().Info("Impossible to contact server", host, "on port 9098 :", err)
		} else {
			if !exist {
				d.remoteServers[host] = true
			}
		}
	}
}
