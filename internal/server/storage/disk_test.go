package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDiskStorage(t *testing.T) {
	fName := `metrics.json`

	metricsSt := NewMetricsStorage()
	metricsSt.SetGauge("nameG", 10)
	metricsSt.SetCounter("nameC", 2)

	diskW, _ := NewDiskWriter(metricsSt, fName)
	err := diskW.Save()
	if err != nil {
		t.Error(err)
	}
	saveResult := metricsSt.GetMetricsJSON()

	metricsSt = NewMetricsStorage()
	diskR, _ := NewDiskReader(metricsSt, fName)
	err = diskR.Load()
	if err != nil {
		t.Error(err)
	}
	loadResult := metricsSt.GetMetricsJSON()
	diskR.Close()

	assert.Equal(t, saveResult, loadResult)
}
