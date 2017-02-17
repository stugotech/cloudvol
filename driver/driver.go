package driver

// Driver abstracts persistent volume functions on cloud
type Driver interface {
	// Attach attaches the given volume to the current instance and returns the device ID
	Attach(volumeID string) (string, error)

	// Detach detaches the given volume from the current instance
	Detach(volumeID string) error
}
