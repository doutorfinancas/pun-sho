package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGeoIPService_NilReader(t *testing.T) {
	log, _ := zap.NewDevelopment()

	svc := NewGeoIPService(log, "", "")

	country, city := svc.Lookup("8.8.8.8")
	assert.Empty(t, country)
	assert.Empty(t, city)
}

func TestGeoIPService_InvalidIP(t *testing.T) {
	log, _ := zap.NewDevelopment()

	svc := NewGeoIPService(log, "", "")

	country, city := svc.Lookup("not-an-ip")
	assert.Empty(t, country)
	assert.Empty(t, city)
}

func TestGeoIPService_MissingDBFile_NoLicenseKey(t *testing.T) {
	log, _ := zap.NewDevelopment()

	svc := NewGeoIPService(log, "/tmp/nonexistent.mmdb", "")
	assert.Nil(t, svc.reader)

	country, city := svc.Lookup("8.8.8.8")
	assert.Empty(t, country)
	assert.Empty(t, city)
}

func TestGeoIPService_Close(t *testing.T) {
	log, _ := zap.NewDevelopment()

	svc := NewGeoIPService(log, "", "")
	svc.Close()
}
