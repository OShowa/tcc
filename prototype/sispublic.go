package sis

import (
	"fmt"
	"hash"
	"sis/internal/crud"
	"sis/internal/data"
	"sis/internal/pk"
)

// SIS is an instance of a Single Instance Storage system with full CRUD capabilities
type SIS struct {
	// main functionality
	h    hash.Hash
	crud crud.Crud
}

func New(h hash.Hash, crud crud.Crud) SIS {
	return SIS{
		h:    h,
		crud: crud,
	}
}

func (s *SIS) GetCrud() crud.Crud {
	return s.crud
}

func (s *SIS) Create(pk pk.PK, blob []byte) error {

	s.h.Write(blob)
	digestSum := s.h.Sum(nil)
	digest := fmt.Sprintf("%x", digestSum)
	s.h.Reset()

	header := data.Header{
		PK:     pk,
		Digest: digest,
	}

	pkExists, err := s.pkExists(pk)
	if err != nil {
		return fmt.Errorf("error on s.pkExists: %w", err)
	}

	if pkExists {
		return fmt.Errorf("pk already exists")
	}

	err = s.persistDataHeader(header)
	if err != nil {
		return fmt.Errorf("error on s.persistDataHeader: %w", err)
	}

	digestExists, err := s.digestExists(digest)
	if err != nil {
		return fmt.Errorf("error on s.digestExists: %w", err)
	}

	if !digestExists {
		err = s.persistBlob(digest, blob)
		if err != nil {
			return fmt.Errorf("error on s.persistBlob: %w", err)
		}
	}

	err = s.addKeyToDigestMetadata(digest, pk)
	if err != nil {
		return fmt.Errorf("error on s.updateDigestMetadata: %w", err)
	}

	return nil

}

func (s *SIS) Read(pk pk.PK) ([]byte, error) {
	header, err := s.readDataHeader(pk)
	if err != nil {
		return nil, fmt.Errorf("error on data header read: %w", err)
	}

	blob, err := s.readBlob(header.Digest)
	if err != nil {
		return nil, fmt.Errorf("error on blob read: %w", err)
	}

	return blob, nil
}

func (s *SIS) Delete(pk pk.PK) error {

	pkExists, err := s.pkExists(pk)
	if err != nil {
		return fmt.Errorf("error on s.pkExists: %w", err)
	}

	if !pkExists {
		return fmt.Errorf("pk does not exist")
	}

	header, err := s.readDataHeader(pk)
	if err != nil {
		return fmt.Errorf("error on data header read: %w", err)
	}

	digestExists, err := s.digestExists(header.Digest)
	if err != nil {
		return fmt.Errorf("error on s.digestExists: %w", err)
	}

	if !digestExists {
		// if code gets here, probably an incomplete deletion previously occured.
		// we should complete it
		err := s.deleteDataHeader(pk)
		if err != nil {
			return fmt.Errorf("error on data header delete: %w", err)
		}
	}

	err = s.deleteDataHeader(pk)
	if err != nil {
		return fmt.Errorf("error on data header delete: %w", err)
	}

	shouldDelete, err := s.removeKeyFromDigestMetadata(header.Digest, pk)
	if err != nil {
		return fmt.Errorf("error on key removal from metadata: %w", err)
	}

	if shouldDelete {
		err = s.deleteBlob(header.Digest)
		if err != nil {
			return fmt.Errorf("error on blob deletion: %w", err)
		}
	}
	return nil

}
