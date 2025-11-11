package crud

import "sis/internal/metrics"

type Crud interface {
	Create(pk []string, blob []byte) error
	Read(pk []string) ([]byte, error)
	Update(pk []string, blob []byte) error
	Delete(pk []string) error
	Exists(pk []string) (bool, error)
	SizeOf(pk []string) (metrics.Byte, error)
}
