package driver

import (
	"fmt"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/stugotech/cloudvol/fs"
)

const (
	mountFormat = "docker-volume-%s"
)

// Driver performs the actual work of managing volumes for the plugin
type Driver interface {
	// Create creates a volume
	Create(name string, options map[string]string) (*Volume, error)
	// Remove removes a volume
	Remove(name string) error
	// Get gets the details of a volume
	Get(name string) (*Volume, error)
	// List lists all volumes
	List() ([]*Volume, error)
	// Mount mounts a volume
	Mount(volume *Volume) error
	// Unmount unmounts a volume
	Unmount(volume *Volume) error
}

// driverInfo
type driverInfo struct {
	storage   StorageDriver
	mountPath string
	volumes   map[string]*Volume
}

// NewDriver creates a new driver
func NewDriver(storage StorageDriver, mountPath string) Driver {
	log.WithFields(log.Fields{"mountPath": mountPath}).Info("creating new driver")
	return &driverInfo{
		storage:   storage,
		mountPath: mountPath,
		volumes:   make(map[string]*Volume),
	}
}

// Create creates a volume
func (d *driverInfo) Create(name string, options map[string]string) (*Volume, error) {
	log.WithFields(log.Fields{"name": name, "options": options}).Info("create volume")

	if _, exists := d.volumes[name]; exists {
		log.WithFields(log.Fields{"name": name}).Error("volume name already exists")
		return nil, fmt.Errorf("volume name '%s' already exists", name)
	}

	vol := &Volume{Name: name}
	d.volumes[name] = vol

	return vol, nil
}

// Remove removes a volume
func (d *driverInfo) Remove(name string) error {
	log.WithFields(log.Fields{"name": name}).Info("remove volume")

	vol, exists := d.volumes[name]
	if !exists {
		log.WithFields(log.Fields{"name": name}).Error("volume not found")
		return fmt.Errorf("volume '%s' not found", name)
	}

	// unmount before removing if necessary
	if vol.MountPath != "" {
		if err := d.storage.Unmount(vol); err != nil {
			return err
		}
	}

	delete(d.volumes, name)
	return nil
}

// Get gets the details of a volume
func (d *driverInfo) Get(name string) (*Volume, error) {
	log.WithFields(log.Fields{"name": name}).Info("get volume")

	vol, exists := d.volumes[name]
	if !exists {
		log.WithFields(log.Fields{"name": name}).Error("volume not found")
		return nil, fmt.Errorf("volume '%s' not found", name)
	}

	return vol, nil
}

// List gets all volumes
func (d *driverInfo) List() ([]*Volume, error) {
	log.Info("list volumes")

	var volumes []*Volume

	for _, vol := range d.volumes {
		volumes = append(volumes, vol)
	}

	return volumes, nil
}

// Mount mounts a volume
func (d *driverInfo) Mount(volume *Volume) error {
	log.WithFields(log.Fields{"name": volume.Name}).Info("mount volume")

	// don't mount twice
	if err := volume.CheckUnmounted(); err != nil {
		return err
	}

	volume.MountPath = d.getMountPath(volume.Name)
	exists, err := fs.DirExists(volume.MountPath)

	if err != nil {
		log.WithFields(log.Fields{
			"name":      volume.Name,
			"mountPath": volume.MountPath,
			"error":     err,
		}).Error("error accessing mount path")

		return fmt.Errorf("error accessing mount path '%s': %v", volume.MountPath, err)
	}

	if !exists {
		log.WithFields(log.Fields{
			"name":      volume.Name,
			"mountPath": volume.MountPath,
		}).Info("creating mount path")

		if err := fs.CreateDir(volume.MountPath, true, 0700); err != nil {
			log.WithFields(log.Fields{
				"name":      volume.Name,
				"mountPath": volume.MountPath,
				"error":     err,
			}).Error("error creating mount path")

			return fmt.Errorf("error creating mount path '%s': %v", volume.MountPath, err)
		}
	}

	if err := d.storage.Mount(volume); err != nil {
		volume.MountPath = ""
		return err
	}

	return nil
}

// Mount mounts a volume
func (d *driverInfo) Unmount(volume *Volume) error {
	log.WithFields(log.Fields{
		"name":      volume.Name,
		"mountPath": volume.MountPath,
	}).Info("unmount volume")

	if err := volume.CheckMounted(); err != nil {
		return err
	}

	if err := d.storage.Unmount(volume); err != nil {
		return err
	}

	if err := fs.RemoveDir(volume.MountPath, true); err != nil {
		log.WithFields(log.Fields{
			"name":      volume.Name,
			"mountPath": volume.MountPath,
			"error":     err,
		}).Warning("error removing mount path")
	}

	volume.MountPath = ""
	return nil
}

func (d *driverInfo) getMountPath(volumeName string) string {
	return path.Join(d.mountPath, fmt.Sprintf(mountFormat, volumeName))
}
