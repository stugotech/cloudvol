package main

import (
	"os"

	"flag"

	"fmt"

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

	port := flag.Int("port", 0, "the port to listen on")
	sock := flag.String("sock", "/var/run/cloudvol.sock", "a unix socket to listen on (ignored if -port is specified)")
	flag.Parse()

	driver, err := driver.NewGceDriver()

	if err != nil {
		log.WithError(err).Fatal("stopping due to last error")
	}

	plugin := plugin.NewCloudvolPlugin("/tmp", driver)
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
