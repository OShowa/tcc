package benchmark

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sis/internal/crud/crudos"
	"sis/internal/metrics"
	"sis/internal/pk"
	"strings"
)

type TestData struct {
	rootDir                                string
	size, maxSize                          metrics.Byte
	crud                                   crudos.CrudOs
	duplicationRate, duplicationRateTarget float64
	entriesMap                             map[string]EntryInfo
	entriesLog                             entriesLog
}

type EntryInfo struct {
	SingleSize metrics.Byte
	Id         string
	Copies     int
}

type entriesLog struct {
	entries []EntryInfo
}

func (e *entriesLog) Add(entryId string, size metrics.Byte) {
	e.entries = append(e.entries, EntryInfo{Id: entryId, SingleSize: size, Copies: 1})
}

type TestInfo struct {
	Size            metrics.Byte `json:"size"`
	DuplicationRate float64      `json:"duplicationRate"`
}

func NewTestData(root pk.PK, maxSize metrics.Byte, duplicationRateTarget float64) (*TestData, error) {
	crud, err := crudos.New(root.Path())
	if err != nil {
		return nil, fmt.Errorf("error initializing crudos instance: %w", err)
	}
	return &TestData{
		rootDir:               root.Path(),
		size:                  0,
		maxSize:               maxSize,
		crud:                  crud,
		duplicationRate:       0,
		duplicationRateTarget: duplicationRateTarget,
		entriesMap:            make(map[string]EntryInfo),
		entriesLog:            entriesLog{},
	}, nil
}

func (t *TestData) DataDir() string {
	return filepath.Join(t.rootDir, "data")
}

func (t *TestData) dataDir() string {
	return "data"
}

func (t *TestData) infoDir() string {
	return "info"
}

func (t *TestData) Info() TestInfo {
	return TestInfo{
		Size:            t.size,
		DuplicationRate: t.duplicationRate,
	}
}

func (t *TestData) Log() []EntryInfo {
	return t.entriesLog.entries
}

func (t *TestData) Size() metrics.Byte {
	return t.size
}

func (t *TestData) DuplicationRate() float64 {
	return t.duplicationRate
}

func (t *TestData) GetDataEntryPaths() ([]string, error) {
	entries, err := os.ReadDir(t.DataDir())
	if err != nil {
		return nil, fmt.Errorf("error reading data dir: %w", err)
	}

	paths := make([]string, len(entries))
	for i, entry := range entries {
		path := filepath.Join(t.DataDir(), entry.Name())
		paths[i] = path
	}

	return paths, nil
}

func (t *TestData) SaveTestInfo() error {
	infoPk := pk.New(t.infoDir())

	infoBytes, err := json.Marshal(t.Info())
	if err != nil {
		return fmt.Errorf("error marshalling info bytes: %w", err)
	}

	infoFilePk := infoPk.Suffix(pk.New("info.json"))

	exists, err := t.crud.Exists(infoFilePk)
	if err != nil {
		return fmt.Errorf("error checking info file existence: %w", err)
	}

	if exists {
		err := t.crud.Delete(infoFilePk)
		if err != nil {
			return fmt.Errorf("error deleting old info file: %w", err)
		}
	}

	err = t.crud.Create(infoFilePk, infoBytes)
	if err != nil {
		return fmt.Errorf("error creating info file: %w", err)
	}

	return nil

}

func (t *TestData) SaveLog() error {
	infoPk := pk.New(t.infoDir())

	logBytes, err := json.Marshal(t.Log())
	if err != nil {
		return fmt.Errorf("error marshalling log bytes: %w", err)
	}

	logPk := infoPk.Suffix(pk.New("log.json"))

	exists, err := t.crud.Exists(logPk)
	if err != nil {
		return fmt.Errorf("error checking log existence: %w", err)
	}

	if exists {
		err := t.crud.Delete(logPk)
		if err != nil {
			return fmt.Errorf("error deleting old log: %w", err)
		}
	}

	err = t.crud.Create(logPk, logBytes)
	if err != nil {
		return fmt.Errorf("error creating log file: %w", err)
	}

	return nil
}

func (t *TestData) AddFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer f.Close()

	fileInfo, err := f.Stat()
	if err != nil {
		return fmt.Errorf("error retrieving file info: %w", err)
	}

	singleSize := metrics.Byte(fileInfo.Size())
	sanitizedName := strings.ReplaceAll(fileInfo.Name(), ".", "-")

	entryInfo := EntryInfo{
		Id:         sanitizedName,
		SingleSize: singleSize,
	}

	shouldDupe := t.shouldDupe()
	finalSize := t.size + singleSize
	if shouldDupe {
		finalSize += metrics.Byte(singleSize)
	}

	if finalSize > t.maxSize {
		return ErrSpaceFull
	}

	if shouldDupe {
		entryInfo.Copies = 2
	} else {
		entryInfo.Copies = 1
	}

	err = t.addEntry(f, entryInfo)
	if err != nil {
		return fmt.Errorf("error adding control entry: %w", err)
	}

	// log.Printf("added file '%s'\n", fileInfo.Name())
	// log.Printf(" duped: %t\n", shouldDupe)
	// log.Printf(" size: %s\n", singleSize)
	log.Printf(" total size: %s/%s\n", t.size, t.maxSize)
	// log.Printf(" duplicationRate: %.2f\n", t.duplicationRate)

	return nil
}

func (t *TestData) shouldDupe() bool {
	return t.duplicationRate < t.duplicationRateTarget
}

func (t *TestData) addEntry(file *os.File, info EntryInfo) error {
	fileData, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("error reading file data: %w", err)
	}

	dataDirPk := pk.New(t.dataDir())

	existingCopies := t.entriesMap[info.Id].Copies
	for i := range info.Copies {
		copyIndex := i + existingCopies + 1
		entryName := fmt.Sprintf("%s-%d", info.Id, copyIndex)
		entryPk := dataDirPk.Suffix(pk.New(entryName))
		err := t.crud.Create(entryPk, fileData)
		if err != nil {
			return fmt.Errorf("error creating '%s' pk on control: %w", entryPk, err)
		}
		t.addNewEntryInfo(info.Id, info.SingleSize)
	}

	return nil
}

func (t *TestData) addNewEntryInfo(entryId string, entrySize metrics.Byte) {
	_, isDuped := t.entriesMap[entryId]
	dupedTotalBytes := t.duplicationRate * float64(t.size)
	if isDuped {
		dupedTotalBytes += float64(entrySize)
	}
	// update entries map
	currEntryInfo := t.entriesMap[entryId]
	currEntryInfo.Copies++
	currEntryInfo.Id = entryId
	currEntryInfo.SingleSize = entrySize
	t.entriesMap[entryId] = currEntryInfo
	// update control space size
	t.size += entrySize
	// calculate new duplication rate
	t.duplicationRate = dupedTotalBytes / float64(t.size)
	// add non-copied entry to log
	t.entriesLog.Add(entryId, entrySize)
}
