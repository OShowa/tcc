package crud

type Crud interface {
	Create(pk []string, blob []byte) error
	Read(pk []string) ([]byte, error)
	Update(pk []string, blob []byte) error
	Delete(pk []string) error
	Exists(pk []string) (bool, error)
}
