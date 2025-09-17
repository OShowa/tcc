package data

import "sis/internal/pk"

type BlobMetadata struct {
	PkList []pk.PK `json:"pkList"`
}
