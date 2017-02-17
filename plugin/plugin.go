package plugin

import (
	"fmt"

	uuid "github.com/satori/go.uuid"
	"github.com/stugotech/cloudvol/driver"
	"github.com/stugotech/cloudvol/fs"

	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/volume"
)

type gceVolume struct {
	name  string
	mount string
}

type cloudvolPlugin struct {
	mountPath string
	volumes   map[string]*gceVolume
	driver    driver.Driver
	fs        fs.FileSystem
}

// NewCloudvolPlugin creates a new instance of the volume plugin
func NewCloudvolPlugin(mountPath string, driver driver.Driver) volume.Driver {
	return &cloudvolPlugin{
		mountPath: mountPath,
		driver:    driver,
	}
}

// Cabailities returns the capabilities of the driver
func (p *cloudvolPlugin) Capabilities(r volume.Request) volume.Response {
	log.Info("plugin get capabilities")
	return volume.Response{}
}

// Create creates a new volume.
func (p *cloudvolPlugin) Create(r volume.Request) volume.Response {
	log.WithFields(log.Fields{"name": r.Name, "options": r.Options}).Info("plugin create volume")

	vol, exists := p.volumes[r.Name]

	if exists {
		// Docker won't always cleanly remove entries.  It's okay so long
		// as the target isn't already mounted by someone else.
		if vol.mount != "" {
			log.WithFields(log.Fields{"name": r.Name}).Error("name already in use")
			return volume.Response{Err: "name already in use"}
		}
	} else {
		// start tracking the volume
		vol = &gceVolume{name: r.Name}
		p.volumes[r.Name] = vol
	}
	return volume.Response{}
}

// List lists all volumes the driver knows of.
func (p *cloudvolPlugin) List(r volume.Request) volume.Response {
	log.Info("plugin list volumes")

	var vols []*volume.Volume

	for _, vol := range p.volumes {
		vols = append(vols, &volume.Volume{Name: vol.name, Mountpoint: vol.mount})
	}

	return volume.Response{Volumes: vols}
}

// Get gets a specific volume.
func (p *cloudvolPlugin) Get(r volume.Request) volume.Response {
	log.WithFields(log.Fields{"name": r.Name}).Info("plugin get volume")
	vol, exist := p.volumes[r.Name]

	if !exist {
		log.WithFields(log.Fields{"name": r.Name}).Error("volume does not exist")
		return volume.Response{Err: fmt.Sprintf("requested volume %s does not exist", r.Name)}
	}
	log.WithFields(log.Fields{"name": vol.name, "mountpoint": vol.mount}).Info("found volume")
	return volume.Response{Volume: &volume.Volume{Name: vol.name, Mountpoint: vol.mount}}
}

// Remove deletes a specific volume.
func (p *cloudvolPlugin) Remove(r volume.Request) volume.Response {
	log.WithFields(log.Fields{"Name": r.Name}).Info("plugin remove volume")
	vol, exist := p.volumes[r.Name]

	if !exist {
		log.WithFields(log.Fields{"name": r.Name}).Error("volume does not exist")
		return volume.Response{Err: fmt.Sprintf("requested volume %s does not exist", r.Name)}
	}

	if vol.mount != "" {
		if err := p.unmount(vol); err != nil {
			log.WithFields(log.Fields{"name": r.Name, "error": err}).Error("error while unmounting")
			return volume.Response{Err: fmt.Sprintf("error while unmounting %s: %v", r.Name, err)}
		}
	}

	delete(p.volumes, r.Name)
	return volume.Response{}
}

// Path gets the path of a given volume.
func (p *cloudvolPlugin) Path(r volume.Request) volume.Response {
	log.WithFields(log.Fields{"Name": r.Name}).Info("plugin get path")
	vol, exist := p.volumes[r.Name]

	if !exist {
		log.WithFields(log.Fields{"name": r.Name}).Error("volume does not exist")
		return volume.Response{Err: fmt.Sprintf("requested volume %s does not exist", r.Name)}
	}
	return volume.Response{Mountpoint: vol.mount}
}

// Mount mounts a volume onto the local file system.
func (p *cloudvolPlugin) Mount(r volume.MountRequest) volume.Response {
	log.WithFields(log.Fields{"Name": r.Name, "ID": r.ID}).Info("plugin mount volume")
	vol, exist := p.volumes[r.Name]

	if !exist {
		log.WithFields(log.Fields{"name": r.Name}).Error("volume does not exist")
		return volume.Response{Err: fmt.Sprintf("requested volume %s does not exist", r.Name)}
	}

	if vol.mount != "" {
		log.WithFields(log.Fields{"name": r.Name}).Error("volume does not exist")
		return volume.Response{Err: fmt.Sprintf("requested volume %s does not exist", r.Name)}
	}

	device, err := p.driver.Attach(r.Name)
	if err != nil {
		log.WithFields(log.Fields{"name": r.Name, "error": err}).Error("error attaching volume")
		return volume.Response{Err: fmt.Sprintf("error attaching volume %s: %v", r.Name, err)}
	}

	mount := path.Join(p.mountPath, uuid.NewV4().String())

	if err := p.fs.CreateDir(mount, true, 0700); err != nil {
		log.WithFields(log.Fields{
			"name":  r.Name,
			"error": err,
			"path":  mount,
		}).Error("error creating directory")
		return volume.Response{Err: fmt.Sprintf("error creating mount path %s: %v", mount, err)}
	}
	if err := p.fs.Mount(device, mount); err != nil {
		log.WithFields(log.Fields{
			"name":   r.Name,
			"error":  err,
			"path":   mount,
			"device": device,
		}).Error("error mounting device")
		return volume.Response{Err: fmt.Sprintf("error mounting device %s: %v", device, err)}
	}

	// save mount point in vol entry
	vol.mount = mount
	return volume.Response{Mountpoint: vol.mount}
}

// Unmount removes a volume from the local file system.
func (p *cloudvolPlugin) Unmount(r volume.UnmountRequest) volume.Response {
	log.WithFields(log.Fields{"Name": r.Name, "ID": r.ID}).Info("plugin unmount volume")
	vol, exist := p.volumes[r.Name]

	if !exist {
		log.WithFields(log.Fields{"name": r.Name}).Error("volume does not exist")
		return volume.Response{Err: fmt.Sprintf("requested volume %s does not exist", r.Name)}
	}

	if vol.mount == "" {
		log.WithFields(log.Fields{"name": r.Name}).Error("volume not mounted")
		return volume.Response{Err: fmt.Sprintf("requested volume %s does not exist", r.Name)}
	}

	if err := p.unmount(vol); err != nil {
		log.WithFields(log.Fields{
			"name":  r.Name,
			"error": err,
			"mount": vol.mount,
		}).Error("error while unmounting")

		return volume.Response{Err: fmt.Sprintf("error while unmounting %s: %v", r.Name, err)}
	}

	return volume.Response{}
}

// unmount unmounts a volume
func (p *cloudvolPlugin) unmount(vol *gceVolume) error {
	if vol.mount == "" {
		return nil
	}
	if err := p.fs.Unmount(vol.mount); err != nil {
		return err
	}
	if err := p.fs.RemoveDir(vol.mount, false); err != nil {
		return err
	}
	vol.mount = ""
	return nil
}
