package testcase_test

import (
	"crypto/sha256"
	"sis"
	"sis/benchmark/testcase"
	"sis/internal/crud/crudos"
	"sis/internal/metrics"
	"testing"
)

const OpenImagesDataDir = "/home/miguel/tcc/tcc/data"

func TestSetEntryweights(t *testing.T) {

	h := sha256.New()
	crudOs, err := crudos.New("./data/test1/root")
	if err != nil {
		t.Fatalf("error creating crudos instance: %s", err.Error())
	}
	sisInstance := sis.New(h, crudOs)
	testCase, err := testcase.NewTestCase(sisInstance, "data/test1", OpenImagesDataDir, "./log", metrics.MB(50), 0.3)
	if err != nil {
		t.Fatalf("error creating test case: %s", err.Error())
	}

	err = testCase.SetEntryWeights()
	if err != nil {
		t.Fatalf("error setting entry weights: %s", err.Error())
	}

}

// func TestGenerateTestData(t *testing.T) {
// 	h := sha256.New()
// 	crudOs, err := crudos.New("./test2/root")
// 	if err != nil {
// 		t.Fatalf("error creating crudos instance: %s", err.Error())
// 	}
// 	sisInstance := sis.New(h, crudOs)
// 	testCase, err := testcase.NewTestCase(sisInstance, "test2", OpenImagesDataDir, "./log", metrics.MB(100), 0.3)
// 	if err != nil {
// 		t.Fatalf("error creating test case: %s", err.Error())
// 	}

// 	err = testCase.GenerateTestData()
// 	if err != nil {
// 		t.Fatalf("error generating control space: %s", err.Error())
// 	}
// }

func TestSISCrawlAll(t *testing.T) {
	h := sha256.New()
	crudOs, err := crudos.New("./data/test4/sis")
	if err != nil {
		t.Fatalf("error creating crudos instance: %s", err.Error())
	}
	sisInstance := sis.New(h, crudOs)
	testCase, err := testcase.NewTestCase(sisInstance, "data/test4", OpenImagesDataDir, "./log", metrics.GB(1), 0.5)
	if err != nil {
		t.Fatalf("error creating test case: %s", err.Error())
	}

	err = testCase.GenerateTestData()
	if err != nil {
		t.Fatalf("error generating control space: %s", err.Error())
	}

	err = testCase.PopulateSIS()
	if err != nil {
		t.Fatalf("error populating SIS: %s", err.Error())
	}

	err = testCase.CompareOriginalWithSIS()
	if err != nil {
		t.Fatalf("error on original comparison with SIS: %s", err.Error())
	}
}
