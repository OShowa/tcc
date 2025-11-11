package testcase

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sis"
	"sis/benchmark"
	"sis/internal/metrics"
	"sis/internal/pk"
)

type TestCase struct {
	// sisInstance is the sis.SIS instance being tested
	sisInstance sis.SIS
	// srcDir is the path to a local directory which contains all available source data to use while testing
	srcDir string
	// entryWeights is used to compute the probability of a random entry on the dataset being picked
	entryWeights []float64
	// testNamespace is the root PK for every persistence inside a test case
	testNamespace pk.PK
	// testData is a benchmark.TestData instance with a duplicated dataset generated from srcDir
	testData benchmark.TestData
	// maxSize is the maximum size of the control space in bytes
	maxSize metrics.Byte
	// expectedDuplicationRate is the approximate rate of duplicated bytes on the test data - if 0.5, roughly 50% of the space is consumed by duplicated files
	expectedDuplicationRate float64
}

func NewTestCase(sisInstance sis.SIS, name, sourceDir, logFilePath string, maxSize metrics.Byte, expectedDuplicationRate float64) (*TestCase, error) {
	testNamespace := pk.New(name)
	if expectedDuplicationRate > 0.5 {
		fmt.Printf("duplication rate %.2f cannot be larger than 50%%\nfixing at 0.50\n", expectedDuplicationRate)
		expectedDuplicationRate = 0.5
	}

	srcInfo, err := os.Stat(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("error getting source dir info: %w", err)
	}
	if !srcInfo.IsDir() {
		return nil, fmt.Errorf("source dir '%s' is not a dir", sourceDir)
	}

	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("error reading source dir: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			return nil, fmt.Errorf("source dir should not have subdirectories")
		}
	}

	testCase := &TestCase{
		sisInstance:             sisInstance,
		srcDir:                  sourceDir,
		testNamespace:           testNamespace,
		maxSize:                 maxSize,
		expectedDuplicationRate: expectedDuplicationRate,
	}

	err = testCase.setUpDirectories()
	if err != nil {
		return nil, fmt.Errorf("error setting up testcase directories: %w", err)
	}

	return testCase, nil
}

func (t *TestCase) setUpDirectories() error {

	root := t.testNamespace.Path()
	testCaseDataPath := filepath.Join(root, "testdata", "data")
	testCaseInfoPath := filepath.Join(root, "testdata", "info")

	err := os.MkdirAll(testCaseDataPath, 0777)
	if err != nil {
		return fmt.Errorf("error creating test case data path: %w", err)
	}

	err = os.MkdirAll(testCaseInfoPath, 0777)
	if err != nil {
		return fmt.Errorf("error creating test case info path: %w", err)
	}

	return nil
}

func (t *TestCase) testDataPk() pk.PK {
	return t.testNamespace.Suffix(pk.New("testdata"))
}

// Reads the original data and the SIS data to find if there are entries lacking, or content was altered
func (t *TestCase) CompareOriginalWithSIS() error {

	entries, err := t.testData.GetDataEntryPaths()
	if err != nil {
		return fmt.Errorf("error getting test data entry paths: %w", err)
	}

	for i, entry := range entries {
		fmt.Printf("comparing entry '%s' (%d/%d)\n", entry, i+1, len(entries))
		originalContent, err := os.ReadFile(entry)
		if err != nil {
			return fmt.Errorf("error reading original file '%s': %w", entry, err)
		}
		pk := pk.New(entry)
		sisContent, err := t.sisInstance.Read(pk)
		if err != nil {
			return fmt.Errorf("error reading SIS file '%s': %w", pk, err)
		}
		err = benchmark.ByteOnByte(originalContent, sisContent)
		if err != nil {
			return fmt.Errorf("found inconsistency: %w", err)
		}
		fmt.Printf("SUCCESS: SIS and original files are identical\n")
	}

	return nil
}

func (t *TestCase) PopulateSIS() error {

	testDataDir := t.testData.DataDir()
	crawler, err := benchmark.NewSISCrawler(t.sisInstance, testDataDir)
	if err != nil {
		return fmt.Errorf("error creating sis crawler: %w", err)
	}

	err = crawler.CrawlAll()
	if err != nil {
		return fmt.Errorf("error crawling through source directory: %w", err)
	}

	return nil
}

func (t *TestCase) GenerateTestData() error {

	testDataPk := t.testDataPk()
	testData, err := benchmark.NewTestData(testDataPk, t.maxSize, t.expectedDuplicationRate)
	t.testData = *testData
	if err != nil {
		return fmt.Errorf("error creating test data instance: %w", err)
	}

	fmt.Println("starting test data generation")

	err = t.SetEntryWeights()
	if err != nil {
		return fmt.Errorf("error setting entry weights: %w", err)
	}

	for {
		picked, err := t.pickRandomPathFromSource()
		if err != nil {
			return fmt.Errorf("error picking random file from source: %w", err)
		}
		err = t.testData.AddFile(picked)
		if err != nil {
			if err == benchmark.ErrSpaceFull {
				break
			}
			return fmt.Errorf("error adding local file to control space: %w", err)
		}
		err = t.testData.SaveLog()
		if err != nil {
			return fmt.Errorf("error saving log: %w", err)
		}
	}

	err = t.testData.SaveTestInfo()
	if err != nil {
		return fmt.Errorf("error saving test info: %w", err)
	}

	info := t.testData.Info()
	fmt.Printf(
		"finished generating test data\n size: %s/%s\n duplication rate: %.2f/%.2f\n",
		info.Size, t.maxSize,
		info.DuplicationRate, t.expectedDuplicationRate,
	)

	return nil

}

func (t *TestCase) SetEntryWeights() error {

	entries, err := os.ReadDir(t.srcDir)
	if err != nil {
		return fmt.Errorf("error reading source dir: %w", err)
	}

	srcDirInfo, err := metrics.MeasureDir(t.srcDir)
	if err != nil {
		return fmt.Errorf("error measuring source dir: %w", err)
	}

	maxSize := srcDirInfo.Max()

	var entryWeights []float64
	var weightsSum float64
	for _, entry := range entries {
		entryInfo, err := entry.Info()
		if err != nil {
			return fmt.Errorf("error retrieving entry info: %w", err)
		}
		entryWeight := float64(maxSize) / float64(entryInfo.Size())
		entryWeights = append(entryWeights, entryWeight)
		weightsSum += entryWeight
	}

	// normaliizing entry weights
	for i, weight := range entryWeights {
		normalized := weight / weightsSum
		entryWeights[i] = normalized
	}

	t.entryWeights = entryWeights

	return nil

}

func (t *TestCase) pickRandomPathFromSource() (string, error) {
	entries, err := os.ReadDir(t.srcDir)
	if err != nil {
		return "", fmt.Errorf("error reading source dir: %w", err)
	}
	picked := entries[t.getWeightedRandomIndex()]
	return filepath.Join(t.srcDir, picked.Name()), nil
}

func (t *TestCase) getWeightedRandomIndex() int {
	randomPercentage := rand.Float64()
	var sum float64
	var currIndex int
	for sum < randomPercentage {
		sum += t.entryWeights[currIndex]
		currIndex++
		if currIndex > len(t.entryWeights) {
			break
		}
	}
	currIndex--
	return currIndex
}
