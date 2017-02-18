package driver

import (
	"github.com/stugotech/cloudvol/fs"
)

// BlockDriver abstracts persistent volume functions on cloud
type BlockDriver interface {
	// Attach attaches the given volume to the current instance and returns the device ID
	Attach(volumeID string) (string, error)
	// Detach detaches the given volume from the current instance
	Detach(volumeID string) error
}

// blockDriverStorage converts a BlockDriver to a StorageDriver
type blockDriverStorage struct {
	BlockDriver
}

// NewBlockDriverStorage creates a StorageDriver from a BlockDriver
func NewBlockDriverStorage(bd BlockDriver) StorageDriver {
	return &blockDriverStorage{BlockDriver: bd}
}

// Mount mounts a volume from a block device
func (d *blockDriverStorage) Mount(volume *Volume) error {
	// first attach block device, then mount in fs
	dev, err := d.Attach(volume.Name)

	if err != nil {
		return err
	}
	if err = fs.Mount(dev, volume.MountPath); err != nil {
		return err
	}
	return nil
}

// Unmount unmounts a block device volume
func (d *blockDriverStorage) Unmount(volume *Volume) error {
	if err := fs.Unmount(volume.MountPath); err != nil {
		return err
	}

	err := d.Detach(volume.Name)

	if err != nil {
		return err
	}
	return nil
}
