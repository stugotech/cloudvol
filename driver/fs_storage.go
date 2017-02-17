package driver

import (
	log "github.com/Sirupsen/logrus"
)

type fsStorage struct {
}

// NewFsStorage creates a StorageDriver for the filesystem
func NewFsStorage() (StorageDriver, error) {
	return &fsStorage{}, nil
}

// Mount mounts a volume from the filesystem
func (d *fsStorage) Mount(volume *Volume) error {
	log.WithFields(log.Fields{"name": volume.Name}).Info("mount volume")
	return nil
}

// Unmount unmounts a volume on the filesystem
func (d *fsStorage) Unmount(volume *Volume) error {
	log.WithFields(log.Fields{
		"name":      volume.Name,
		"mountPath": volume.MountPath,
	}).Info("unmount volume")

	return nil
}
