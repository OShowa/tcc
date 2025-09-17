package constants

import (
	"path"
	"sis/internal/pk"
)

// constants
var UserDataSpace pk.PK = pk.New(path.Join("user", "data"))
var SystemDataSpace pk.PK = pk.New(path.Join("sys", "data"))
var DataHeaderSuffix pk.PK = pk.New("data-header")
