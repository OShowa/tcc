package test

import (
	"sis/internal/metrics"
	"testing"
)

func TestMeasureDir(t *testing.T) {

	dirPath := "/home/miguel/tcc/tcc/prototype/benchmark/testcase/_test/root/test1/control"

	info, err := metrics.MeasureDir(dirPath)
	if err != nil {
		t.Fatalf("error measuring directory: %s", err.Error())
	}

	t.Log(info.Size())
}
