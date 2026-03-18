package service

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/oschwald/geoip2-golang"
	"go.uber.org/zap"
)

const (
	maxMindDownloadURL = "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=%s&suffix=tar.gz"
)

type GeoIPService struct {
	log    *zap.Logger
	reader *geoip2.Reader
}

func NewGeoIPService(log *zap.Logger, dbPath, licenseKey string) *GeoIPService {
	svc := &GeoIPService{log: log}

	if dbPath == "" {
		log.Info("GeoIP database path not configured, geo lookups disabled")
		return svc
	}

	// Try to open existing database
	reader, err := geoip2.Open(dbPath)
	if err == nil {
		svc.reader = reader
		log.Info("GeoIP database loaded", zap.String("path", dbPath))
		return svc
	}

	// File doesn't exist — try to download if license key is provided
	if licenseKey == "" {
		log.Info("GeoIP database not found and no license key configured, geo lookups disabled")
		return svc
	}

	log.Info("GeoIP database not found, downloading from MaxMind...")
	if downloadErr := downloadGeoIPDB(dbPath, licenseKey); downloadErr != nil {
		log.Warn("Failed to download GeoIP database, geo lookups disabled", zap.Error(downloadErr))
		return svc
	}

	reader, err = geoip2.Open(dbPath)
	if err != nil {
		log.Warn("Failed to open downloaded GeoIP database", zap.Error(err))
		return svc
	}

	svc.reader = reader
	log.Info("GeoIP database downloaded and loaded", zap.String("path", dbPath))

	return svc
}

func downloadGeoIPDB(destPath, licenseKey string) error {
	url := fmt.Sprintf(maxMindDownloadURL, licenseKey)

	resp, err := http.Get(url) //nolint:gosec // URL is constructed from a constant + config value
	if err != nil {
		return fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	// The response is a tar.gz containing a directory with the .mmdb file
	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to decompress: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		if !strings.HasSuffix(header.Name, ".mmdb") {
			continue
		}

		// Ensure destination directory exists
		dir := filepath.Dir(destPath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		out, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}

		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			return fmt.Errorf("failed to write file: %w", err)
		}
		out.Close()

		return nil
	}

	return fmt.Errorf("no .mmdb file found in archive")
}

func (s *GeoIPService) Lookup(ipStr string) (country, city string) {
	if s.reader == nil {
		return "", ""
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "", ""
	}

	record, err := s.reader.City(ip)
	if err != nil {
		return "", ""
	}

	country = record.Country.Names["en"]
	if len(record.City.Names) > 0 {
		city = record.City.Names["en"]
	}

	return country, city
}

func (s *GeoIPService) Close() {
	if s.reader != nil {
		s.reader.Close()
	}
}
