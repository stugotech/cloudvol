package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/volume"
	"github.com/stugotech/cloudvol/driver"
	"github.com/stugotech/cloudvol/plugin"
)

const (
	driverName = "cloudvol"
)

func main() {
	log.WithFields(log.Fields{"pid": os.Getpid()}).Info("*** STARTED cloudvol volume driver ***")

	mode := flag.String("mode", "fs", "storage mode (fs, gce, aws)")
	port := flag.Int("port", 0, "port to listen on")
	sock := flag.String("sock", "cloudvol", "a unix socket to listen on (ignored if -port is specified)")
	flag.Parse()

	log.WithFields(log.Fields{"mode": *mode}).Info("creating storage driver")

	storage, err := createStorageDriver(*mode)

	if err != nil {
		log.WithError(err).Fatal("stopping due to last error")
	}

	d := driver.NewDriver(storage, "/mnt")
	plugin := plugin.NewCloudvolPlugin(d)
	handler := volume.NewHandler(plugin)

	if *port > 0 {
		log.WithFields(log.Fields{"port": *port}).Infof("listening on port %d", *port)
		addr := fmt.Sprintf(":%d", *port)
		err = handler.ServeTCP(driverName, addr, nil)
	} else {
		log.WithFields(log.Fields{"socket": *sock}).Infof("listening on socket file")
		err = handler.ServeUnix("root", *sock)
	}

	if err != nil {
		log.Fatal(err)
	} else {
		log.Info("Started.")
	}
}

func createStorageDriver(name string) (driver.StorageDriver, error) {
	if name == "gce" {
		block, err := driver.NewGceDriver()

		if err != nil {
			return nil, err
		}

		return driver.NewBlockDriverStorage(block), nil
	} else if name == "fs" {
		return driver.NewFsStorage()
	} else {
		return nil, fmt.Errorf("unknown driver type '%s'", name)
	}
}
