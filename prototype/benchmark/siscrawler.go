package benchmark

import (
	"fmt"
	"os"
	"path/filepath"
	"sis"
	"sis/internal/pk"
)

// a SIS Crawler takes every file from a local srcDir and saves it to a SIS instance
type SISCrawler struct {
	sisInstance    sis.SIS
	srcDir         string
	crawlPath      []fileEntry
	alreadyCrawled bool
}

type fileEntry struct {
	destKey pk.PK
	srcPath string
}

func NewSISCrawler(sisInstance sis.SIS, srcDir string) (*SISCrawler, error) {
	info, err := os.Stat(srcDir)
	if err != nil {
		return nil, fmt.Errorf("error getting srcDir info: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("srcDir '%s' is not a directory", srcDir)
	}
	return &SISCrawler{
		sisInstance: sisInstance,
		srcDir:      srcDir,
	}, nil
}

func (c *SISCrawler) CrawlAll() error {
	var err error
	for err == nil {
		err = c.Crawl()
	}
	if err != ErrEOF {
		return fmt.Errorf("unexpected error during crawling: %w", err)
	}
	return nil
}

func (c *SISCrawler) Crawl() error {

	if !c.alreadyCrawled {
		err := c.generateCrawlPath()
		if err != nil {
			return fmt.Errorf("error generating crawl path: %w", err)
		}
	}

	if len(c.crawlPath) == 0 {
		return ErrEOF
	}

	c.alreadyCrawled = true
	currEntry := c.crawlPath[0]
	fileBytes, err := os.ReadFile(currEntry.srcPath)
	if err != nil {
		return fmt.Errorf("error reading source file '%s': %w", currEntry.srcPath, err)
	}
	err = c.sisInstance.Create(currEntry.destKey, fileBytes)
	if err != nil {
		return fmt.Errorf("error creating destination file '%s': %w", currEntry.destKey.Path(), err)
	}

	if len(c.crawlPath) > 1 {
		c.crawlPath = c.crawlPath[1:]
	} else {
		c.crawlPath = []fileEntry{}
	}

	return nil
}

func (c *SISCrawler) generateCrawlPath() error {

	fileEntries, err := c.getDirFileEntries(c.srcDir)
	if err != nil {
		return fmt.Errorf("error getting file entries: %w", err)
	}
	c.crawlPath = fileEntries
	return nil
}

func (c *SISCrawler) getDirFileEntries(dirPath string) ([]fileEntry, error) {

	dirEntries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error reading dir '%s': %w", dirPath, err)
	}

	var fileEntries []fileEntry
	for _, entry := range dirEntries {
		entryPath := filepath.Join(dirPath, entry.Name())
		if entry.IsDir() {
			currFileEntries, err := c.getDirFileEntries(entryPath)
			if err != nil {
				return nil, fmt.Errorf("error reading dir '%s': %w", dirPath, err)
			}
			fileEntries = append(fileEntries, currFileEntries...)
			continue
		}
		currFileEntry := fileEntry{
			srcPath: entryPath,
			destKey: pk.New(entryPath),
		}
		fileEntries = append(fileEntries, currFileEntry)
	}

	return fileEntries, nil
}
