package driver

// StorageDriver abstracts storage availability functions
type StorageDriver interface {
	// Mount mounts a volume
	Mount(volume *Volume) error
	// Unmount unmounts a volume
	Unmount(volume *Volume) error
}
