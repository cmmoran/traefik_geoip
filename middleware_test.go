package traefik_geoip_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	mw "github.com/cmmoran/traefik_geoip" //nolint:depguard
)

const (
	ValidIP       = "188.193.88.199"
	ValidIPNoCity = "20.1.184.61"
)

func TestGeoIPConfig(t *testing.T) {
	mwCfg := mw.CreateConfig()
	if mw.DefaultDBPath != mwCfg.DBPath {
		t.Fatalf("Incorrect path")
	}

	mwCfg.DBPath = "./non-existing"
	_, err := mw.New(context.TODO(), nil, mwCfg, "")
	if err == nil {
		t.Fatalf("Must fail on missing DB")
	}

	mwCfg.DBPath = "./README.md"
	_, err = mw.New(context.TODO(), nil, mwCfg, "")
	if err == nil {
		t.Fatalf("Must fail on invalid DB format")
	}
}

func TestGeoIPBasic(t *testing.T) {
	mwCfg := mw.CreateConfig()
	mwCfg.DBPath = "./geolite2-city.mmdb"

	called := false
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) { called = true })

	instance, err := mw.New(context.TODO(), next, mwCfg, "traefik_geoip")
	if err != nil {
		t.Fatalf("Error creating %v", err)
	}

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)

	instance.ServeHTTP(recorder, req)
	if recorder.Result().StatusCode != http.StatusOK {
		t.Fatalf("Invalid return code")
	}
	if called != true {
		t.Fatalf("next handler was not called")
	}
}

func TestMissingGeoIPDB(t *testing.T) {
	mwCfg := mw.CreateConfig()
	mwCfg.DBPath = "./missing"

	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	_, err := mw.New(context.TODO(), next, mwCfg, "traefik_geoip")
	if err == nil {
		t.Fatalf("Did not abort creation after an invalid DB: %v", err)
	}
}

func TestGeoIPFromRemoteAddr(t *testing.T) {
	mwCfg := mw.CreateConfig()
	mwCfg.DBPath = "./geolite2-city.mmdb"
	mwCfg.Debug = true

	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
	instance, _ := mw.New(context.TODO(), next, mwCfg, "traefik_geoip")

	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	req.RemoteAddr = fmt.Sprintf("%s:9999", ValidIP)
	instance.ServeHTTP(httptest.NewRecorder(), req)
	assertHeader(t, req, mw.CountryHeader, "Germany")
	assertHeader(t, req, mw.CountryCodeHeader, "DE")
	assertHeader(t, req, mw.RegionHeader, "BY")
	assertHeader(t, req, mw.CityHeader, "Munich")
	assertHeader(t, req, mw.LatitudeHeader, "48.1872")
	assertHeader(t, req, mw.LongitudeHeader, "11.4802")
	assertHeader(t, req, mw.GeohashHeader, "u284jhpwu11s")

	req = httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	req.RemoteAddr = fmt.Sprintf("%s:9999", ValidIPNoCity)
	instance.ServeHTTP(httptest.NewRecorder(), req)
	assertHeader(t, req, mw.CountryHeader, "United States")
	assertHeader(t, req, mw.CountryCodeHeader, "US")
	assertHeader(t, req, mw.RegionHeader, "VA")
	assertHeader(t, req, mw.CityHeader, "Boydton")
	assertHeader(t, req, mw.LatitudeHeader, "36.6676")
	assertHeader(t, req, mw.LongitudeHeader, "-78.3875")
	assertHeader(t, req, mw.GeohashHeader, "dq8285puqb59")

	req = httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	req.RemoteAddr = "qwerty:9999"
	instance.ServeHTTP(httptest.NewRecorder(), req)
	assertHeader(t, req, mw.CountryHeader, "")
	assertHeader(t, req, mw.CountryCodeHeader, "")
	assertHeader(t, req, mw.RegionHeader, "")
	assertHeader(t, req, mw.CityHeader, "")
	assertHeader(t, req, mw.LatitudeHeader, "")
	assertHeader(t, req, mw.LongitudeHeader, "")
	assertHeader(t, req, mw.GeohashHeader, "")
}

func TestGeoIPCountryDBFromRemoteAddr(t *testing.T) {
	mwCfg := mw.CreateConfig()
	mwCfg.DBPath = "./geolite2-country.mmdb"

	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
	instance, _ := mw.New(context.TODO(), next, mwCfg, "traefik_geoip")

	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	req.RemoteAddr = fmt.Sprintf("%s:9999", ValidIP)
	instance.ServeHTTP(httptest.NewRecorder(), req)

	assertHeader(t, req, mw.CountryHeader, "Germany")
	assertHeader(t, req, mw.CountryCodeHeader, "DE")
	assertHeader(t, req, mw.RegionHeader, "")
	assertHeader(t, req, mw.CityHeader, "")
	assertHeader(t, req, mw.LatitudeHeader, "")
	assertHeader(t, req, mw.LongitudeHeader, "")
	assertHeader(t, req, mw.GeohashHeader, "")
}

func TestIgnoresExcludedIPs(t *testing.T) {
	mwCfg := mw.CreateConfig()
	mwCfg.DBPath = "./geolite2-city.mmdb"
	mwCfg.Debug = true
	mwCfg.ExcludeIPs = []string{ValidIP}

	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
	instance, _ := mw.New(context.TODO(), next, mwCfg, "traefik_geoip")

	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	req.RemoteAddr = fmt.Sprintf("%s:9999", ValidIP)
	instance.ServeHTTP(httptest.NewRecorder(), req)
	assertHeader(t, req, mw.CountryHeader, "")
	assertHeader(t, req, mw.CountryCodeHeader, "")
	assertHeader(t, req, mw.RegionHeader, "")
	assertHeader(t, req, mw.CityHeader, "")
	assertHeader(t, req, mw.LatitudeHeader, "")
	assertHeader(t, req, mw.LongitudeHeader, "")
	assertHeader(t, req, mw.GeohashHeader, "")
}

func TestHandleInvalidExcludeIP(t *testing.T) {
	mwCfg := mw.CreateConfig()
	mwCfg.DBPath = "./geolite2-city.mmdb"
	mwCfg.Debug = true
	mwCfg.ExcludeIPs = []string{"invalid"}

	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
	instance, _ := mw.New(context.TODO(), next, mwCfg, "traefik_geoip")

	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	req.RemoteAddr = fmt.Sprintf("%s:9999", ValidIP)
	instance.ServeHTTP(httptest.NewRecorder(), req)
	assertHeader(t, req, mw.CountryHeader, "Germany")
	assertHeader(t, req, mw.CountryCodeHeader, "DE")
	assertHeader(t, req, mw.RegionHeader, "BY")
	assertHeader(t, req, mw.CityHeader, "Munich")
	assertHeader(t, req, mw.LatitudeHeader, "48.1872")
	assertHeader(t, req, mw.LongitudeHeader, "11.4802")
	assertHeader(t, req, mw.GeohashHeader, "u284jhpwu11s")
}

func TestGeoIPFromXForwardedFor(t *testing.T) {
	mwCfg := mw.CreateConfig()
	mwCfg.DBPath = "./geolite2-city.mmdb"
	mwCfg.Debug = true

	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
	instance, _ := mw.New(context.TODO(), next, mwCfg, "traefik_geoip")

	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	req.Header.Set("X-Forwarded-For", ValidIP+", 192.168.1.1")
	instance.ServeHTTP(httptest.NewRecorder(), req)
	assertHeader(t, req, mw.CountryHeader, "Germany")
	assertHeader(t, req, mw.CountryCodeHeader, "DE")
	assertHeader(t, req, mw.RegionHeader, "BY")
	assertHeader(t, req, mw.CityHeader, "Munich")
	assertHeader(t, req, mw.LatitudeHeader, "48.1872")
	assertHeader(t, req, mw.LongitudeHeader, "11.4802")
	assertHeader(t, req, mw.GeohashHeader, "u284jhpwu11s")
}

func assertHeader(t *testing.T, req *http.Request, key, expected string) {
	t.Helper()
	if req.Header.Get(key) != expected {
		t.Fatalf("invalid value of header [%s] is '%s', not '%s'", key, req.Header.Get(key), expected)
	}
}
