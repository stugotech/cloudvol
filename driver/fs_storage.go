package driver

type fsStorage struct {
}

// NewFsStorage creates a StorageDriver for the filesystem
func NewFsStorage() (StorageDriver, error) {
	return &fsStorage{}, nil
}

// Mount mounts a volume from the filesystem
func (d *fsStorage) Mount(volume *Volume) error {
	return nil
}

// Unmount unmounts a volume on the filesystem
func (d *fsStorage) Unmount(volume *Volume) error {
	return nil
}
