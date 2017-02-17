package driver

import (
	"fmt"
	"time"

	"cloud.google.com/go/compute/metadata"
	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
)

const (
	deviceBaseName        = "docker-volume-%s"
	devicePathFormat      = "/dev/disk/by-id/google-%s"
	operationWaitTimeout  = 5 * time.Second
	operationPollInterval = 100 * time.Millisecond
)

type gceDriver struct {
	client   *compute.Service
	project  string
	zone     string
	instance string
}

// NewGceDriver creates a new instance of the GCE volume driver
func NewGceDriver() (Driver, error) {
	if !metadata.OnGCE() {
		log.Warn("not running on GCE")
		return nil, fmt.Errorf("not running on GCE")
	}

	ctx := context.Background()

	client, err := google.DefaultClient(ctx, compute.ComputeScope)
	if err != nil {
		log.WithError(err).Warn("error creating GCE client")
		return nil, fmt.Errorf("error creating GCE client: %s", err)
	}

	computeService, err := compute.New(client)
	if err != nil {
		log.WithError(err).Warn("error creating GCE client")
		return nil, fmt.Errorf("error creating GCE client: %s", err)
	}

	instance, err := metadata.InstanceName()
	if err != nil {
		log.WithError(err).Warn("error retrieving GCE instance name")
		return nil, fmt.Errorf("error retrieving instance name: %s", err)
	}

	zone, err := metadata.Zone()
	if err != nil {
		log.WithError(err).Warn("error retrieving GCE zone")
		return nil, fmt.Errorf("error retrieving zone: %s", err)
	}

	project, err := metadata.ProjectID()
	if err != nil {
		log.WithError(err).Warn("error retrieving GCE project ID")
		return nil, fmt.Errorf("error retrieving project ID: %s", err)
	}

	log.WithFields(log.Fields{
		"instance": instance,
		"zone":     zone,
		"project":  project,
	}).Info("detected instance parameters")

	provider := &gceDriver{
		client:   computeService,
		instance: instance,
		zone:     zone,
		project:  project,
	}
	return provider, nil
}

// Attach attaches a disk to the current instance
func (d *gceDriver) Attach(name string) (string, error) {
	log.WithFields(log.Fields{"name": name}).Info("attach disk")
	device := fmt.Sprintf(deviceBaseName, name)

	disk := compute.AttachedDisk{
		DeviceName: device,
	}

	op, err := d.client.Instances.AttachDisk(d.project, d.zone, d.instance, &disk).Do()

	if err != nil {
		log.WithError(err).Warn("error attaching disk")
		return "", fmt.Errorf("error attaching disk: %s", err)
	}

	err = d.waitForOp(op)

	if err != nil {
		return "", err
	}
	// return the path to the device
	devicePath := fmt.Sprintf(devicePathFormat, device)
	log.WithFields(log.Fields{"devicePath": devicePath}).Info("attached device")
	return devicePath, nil
}

// Detach detaches a disk from the current instance
func (d *gceDriver) Detach(name string) error {
	log.WithFields(log.Fields{"name": name}).Info("detach disk")
	device := fmt.Sprintf(deviceBaseName, name)

	op, err := d.client.Instances.DetachDisk(d.project, d.zone, d.instance, device).Do()
	if err != nil {
		log.WithError(err).Warn("error detaching disk")
		return err
	}

	return d.waitForOp(op)
}

// waitForOp waits for an operation to complete
func (d *gceDriver) waitForOp(op *compute.Operation) error {
	// poll for operation completion
	for start := time.Now(); time.Since(start) < operationWaitTimeout; time.Sleep(operationPollInterval) {
		log.WithFields(log.Fields{
			"project":   d.project,
			"zone":      d.zone,
			"operation": op.Name,
		}).Info("waiting for GCE operation")

		if op, err := d.client.ZoneOperations.Get(d.project, d.zone, op.Name).Do(); err == nil {
			log.WithFields(log.Fields{
				"project":   d.project,
				"zone":      d.zone,
				"operation": op.Name,
				"status":    op.Status,
			}).Info("GCE operation status")

			if op.Status == "DONE" {
				return nil
			}
		} else {
			// output warning
			log.WithFields(log.Fields{
				"operation":  op.Name,
				"targetLink": op.TargetLink,
				"error":      err,
			}).Warn("error while getting GCE operation")
		}
	}

	log.WithFields(log.Fields{
		"operation":  op.Name,
		"targetLink": op.TargetLink,
		"timeout":    operationWaitTimeout,
	}).Warn("timeout while waiting for GCE operation to complete")

	return fmt.Errorf("timeout while waiting for operation %s on %s to complete", op.Name, op.TargetLink)
}
