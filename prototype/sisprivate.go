package sis

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sis/internal/constants"
	"sis/internal/data"
	"sis/internal/pk"
)

func (s SIS) pkExists(key pk.PK) (bool, error) {

	exists, err := s.dataHeaderExists(key)
	if err != nil {
		return exists, fmt.Errorf("error on s.dataHeaderExists: %w", err)
	}

	return exists, err

}

func (s SIS) dataHeaderExists(key pk.PK) (bool, error) {
	return s.crud.Exists(key.Prefix(constants.UserDataSpace).Suffix(constants.DataHeaderSuffix))
}

func (s SIS) readDataHeader(key pk.PK) (data.Header, error) {

	headerBlob, err := s.crud.Read(key.Prefix(constants.UserDataSpace).Suffix(constants.DataHeaderSuffix))
	if err != nil {
		return data.Header{}, fmt.Errorf("error on s.crud.Read: %w", err)
	}

	var header data.Header
	err = json.Unmarshal(headerBlob, &header)
	if err != nil {
		return data.Header{}, fmt.Errorf("error on header unmarshal: %w", err)
	}

	return header, nil
}

func (s SIS) deleteDataHeader(key pk.PK) error {
	dataHeaderPk := key.Prefix(constants.UserDataSpace).Suffix(constants.DataHeaderSuffix)

	err := s.crud.Delete(dataHeaderPk)
	if err != nil {
		return fmt.Errorf("error on s.crud.Delete: %w", err)
	}

	return nil
}

func (s SIS) persistDataHeader(header data.Header) error {

	headerBytes, err := json.Marshal(header)
	if err != nil {
		return fmt.Errorf("error on header marshal: %w", err)
	}

	dataHeaderPk := header.PK.Prefix(constants.UserDataSpace).Suffix(constants.DataHeaderSuffix)

	err = s.crud.Create(dataHeaderPk, headerBytes)
	if err != nil {
		return fmt.Errorf("error on s.crud.Create: %w", err)
	}

	return nil

}

func (s SIS) digestExists(digest string) (bool, error) {
	blobPk := constants.SystemDataSpace.Suffix(pk.New(filepath.Join(digest, "blob")))
	return s.crud.Exists(blobPk)
}

func (s SIS) readBlob(digest string) ([]byte, error) {
	blobPk := constants.SystemDataSpace.Suffix(pk.New(filepath.Join(digest, "blob")))

	blob, err := s.crud.Read(blobPk)
	if err != nil {
		return nil, fmt.Errorf("error on blob s.crud.Read: %w", err)
	}

	return blob, nil
}

func (s SIS) deleteBlob(digest string) error {
	blobPk := constants.SystemDataSpace.Suffix(pk.New(filepath.Join(digest, "blob")))

	err := s.crud.Delete(blobPk)
	if err != nil {
		return fmt.Errorf("error on blob s.crud.Delete: %w", err)
	}

	return nil
}

func (s SIS) persistBlob(digest string, blob []byte) error {

	blobPk := constants.SystemDataSpace.Suffix(pk.New(filepath.Join(digest, "blob")))
	metadataPk := constants.SystemDataSpace.Suffix(pk.New(filepath.Join(digest, "metadata")))

	metadata := data.BlobMetadata{
		PkList: make([]pk.PK, 0),
	}

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("error on metadata marshal: %w", err)
	}

	err = s.crud.Create(blobPk, blob)
	if err != nil {
		return fmt.Errorf("error on blob s.crud.Create: %w", err)
	}

	err = s.crud.Create(metadataPk, metadataBytes)
	if err != nil {
		return fmt.Errorf("error on metadata s.crudCreate: %w", err)
	}

	return nil
}

func (s SIS) addKeyToDigestMetadata(digest string, key pk.PK) error {

	metadata, err := s.readBlobMetadata(digest)
	if err != nil {
		return fmt.Errorf("error on blob metadata read: %w", err)
	}
	metadata.PkList = append(metadata.PkList, key)

	err = s.updateBlobMetadata(digest, metadata)
	if err != nil {
		return fmt.Errorf("error on blob metadata update: %w", err)
	}

	return nil
}

func (s SIS) removeKeyFromDigestMetadata(digest string, key pk.PK) (shouldDelete bool, err error) {

	metadata, err := s.readBlobMetadata(digest)
	if err != nil {
		return false, fmt.Errorf("error on blob metadata read: %w", err)
	}

	if len(metadata.PkList) == 0 {
		err = s.deleteBlobMetadata(digest)
		if err != nil {
			return false, fmt.Errorf("error deleting blob metadata: %w", err)
		}
		return true, nil
	}

	foundIndex := -1
	for i, currKey := range metadata.PkList {
		if currKey.Path() == key.Path() {
			foundIndex = i
			break
		}
	}

	if foundIndex == -1 {
		return false, fmt.Errorf("could not find index of pk '%s' on digest metadata", key.Path())
	}
	metadata.PkList[foundIndex] = metadata.PkList[len(metadata.PkList)-1]
	metadata.PkList = metadata.PkList[:len(metadata.PkList)-1]

	err = s.updateBlobMetadata(digest, metadata)
	if err != nil {
		return false, fmt.Errorf("error on blob metadata update: %w", err)
	}

	if len(metadata.PkList) == 0 {
		err = s.deleteBlobMetadata(digest)
		if err != nil {
			return false, fmt.Errorf("error deleting blob metadata: %w", err)
		}
		return true, nil
	}

	return false, nil
}

func (s SIS) updateBlobMetadata(digest string, new data.BlobMetadata) error {

	metadataPk := constants.SystemDataSpace.Suffix(pk.New(filepath.Join(digest, "metadata")))

	newMetadataBytes, err := json.Marshal(new)
	if err != nil {
		return fmt.Errorf("error on new metadata marshal: %w", err)
	}

	return s.crud.Update(metadataPk, newMetadataBytes)
}

func (s SIS) readBlobMetadata(digest string) (data.BlobMetadata, error) {

	metadataPk := constants.SystemDataSpace.Suffix(pk.New(filepath.Join(digest, "metadata")))

	metadataBytes, err := s.crud.Read(metadataPk)
	if err != nil {
		return data.BlobMetadata{}, fmt.Errorf("error on metadata s.crud.Read: %w", err)
	}

	var metadata data.BlobMetadata
	err = json.Unmarshal(metadataBytes, &metadata)
	if err != nil {
		return data.BlobMetadata{}, fmt.Errorf("error on metadata unmarshal: %w", err)
	}

	return metadata, nil

}

func (s SIS) deleteBlobMetadata(digest string) error {
	metadataPk := constants.SystemDataSpace.Suffix(pk.New(filepath.Join(digest, "metadata")))

	err := s.crud.Delete(metadataPk)
	if err != nil {
		return fmt.Errorf("error deleting metadata file: %w", err)
	}

	return nil

}
