package driver

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
)

// Volume contains details of a plugin
type Volume struct {
	Name      string
	MountPath string
}

// CheckMounted makes sure volume is mounted
func (v *Volume) CheckMounted() error {
	if v.MountPath == "" {
		log.WithFields(log.Fields{
			"name": v.Name,
		}).Error("not mounted")

		return fmt.Errorf("volume '%s' not mounted", v.Name)
	}
	return nil
}

// CheckUnmounted makes sure volume is mounted
func (v *Volume) CheckUnmounted() error {
	if v.MountPath != "" {
		log.WithFields(log.Fields{
			"name":      v.Name,
			"mountPath": v.MountPath,
		}).Error("already mounted")

		return fmt.Errorf("volume '%s' already mounted on '%s'", v.Name, v.MountPath)
	}
	return nil
}
