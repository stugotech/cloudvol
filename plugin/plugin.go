package plugin

import (
	"fmt"

	"github.com/stugotech/cloudvol/driver"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/volume"
)

type gceVolume struct {
	name  string
	mount string
}

type cloudvolPlugin struct {
	driver driver.Driver
}

// NewCloudvolPlugin creates a new instance of the volume plugin
func NewCloudvolPlugin(driver driver.Driver) volume.Driver {
	return &cloudvolPlugin{
		driver: driver,
	}
}

// Cabailities returns the capabilities of the driver
func (p *cloudvolPlugin) Capabilities(r volume.Request) volume.Response {
	log.Info("plugin get capabilities")
	return volume.Response{}
}

// Create creates a new volume.
func (p *cloudvolPlugin) Create(r volume.Request) volume.Response {
	_, err := p.driver.Create(r.Name, r.Options)

	if err != nil {
		return volume.Response{Err: fmt.Sprintf("error creating volume '%s': %v", r.Name, err)}
	}

	return volume.Response{}
}

// List lists all volumes the driver knows of.
func (p *cloudvolPlugin) List(r volume.Request) volume.Response {
	log.Info("plugin list volumes")

	driverVols, err := p.driver.List()

	if err != nil {
		return volume.Response{Err: fmt.Sprintf("error listing volumes: %v", err)}
	}

	var vols []*volume.Volume

	for _, vol := range driverVols {
		vols = append(vols, &volume.Volume{Name: vol.Name, Mountpoint: vol.MountPath})
	}

	return volume.Response{Volumes: vols}
}

// Get gets a specific volume.
func (p *cloudvolPlugin) Get(r volume.Request) volume.Response {
	log.WithFields(log.Fields{"name": r.Name}).Info("plugin get volume")

	vol, err := p.driver.Get(r.Name)

	if err != nil {
		return volume.Response{Err: fmt.Sprintf("error getting volume '%s': %v", r.Name, err)}
	}

	return volume.Response{Volume: &volume.Volume{Name: vol.Name, Mountpoint: vol.MountPath}}
}

// Remove deletes a specific volume.
func (p *cloudvolPlugin) Remove(r volume.Request) volume.Response {
	log.WithFields(log.Fields{"Name": r.Name}).Info("plugin remove volume")

	err := p.driver.Remove(r.Name)

	if err != nil {
		return volume.Response{Err: fmt.Sprintf("error removing volume '%s': %v", r.Name, err)}
	}

	return volume.Response{}
}

// Path gets the path of a given volume.
func (p *cloudvolPlugin) Path(r volume.Request) volume.Response {
	log.WithFields(log.Fields{"Name": r.Name}).Info("plugin get path")
	vol, err := p.driver.Get(r.Name)

	if err != nil {
		return volume.Response{Err: fmt.Sprintf("error getting volume '%s': %v", r.Name, err)}
	}
	return volume.Response{Mountpoint: vol.MountPath}
}

// Mount mounts a volume onto the local file system.
func (p *cloudvolPlugin) Mount(r volume.MountRequest) volume.Response {
	log.WithFields(log.Fields{"Name": r.Name, "ID": r.ID}).Info("plugin mount volume")
	vol, err := p.driver.Get(r.Name)

	if err != nil {
		return volume.Response{Err: fmt.Sprintf("error getting volume '%s': %v", r.Name, err)}
	}
	if err = p.driver.Mount(vol); err != nil {
		return volume.Response{Err: fmt.Sprintf("error mounting volume '%s': %v", r.Name, err)}
	}
	return volume.Response{Mountpoint: vol.MountPath}
}

// Unmount removes a volume from the local file system.
func (p *cloudvolPlugin) Unmount(r volume.UnmountRequest) volume.Response {
	log.WithFields(log.Fields{"Name": r.Name, "ID": r.ID}).Info("plugin unmount volume")
	vol, err := p.driver.Get(r.Name)

	if err != nil {
		return volume.Response{Err: fmt.Sprintf("error getting volume '%s': %v", r.Name, err)}
	}
	if err = p.driver.Unmount(vol); err != nil {
		return volume.Response{Err: fmt.Sprintf("error unmounting volume '%s': %v", r.Name, err)}
	}
	return volume.Response{Mountpoint: vol.MountPath}
}
