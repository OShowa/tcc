package data

import (
	"sis/internal/pk"
)

type Header struct {
	PK       pk.PK          `json:"pk"`
	Digest   string         `json:"digest"`
	Metadata map[string]any `json:"metadata,omitempty"`
}
