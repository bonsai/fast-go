package util

import (
	"encoding/json"
	"net/http"
	"time"
)

type geoIPResponse struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

var geoClient = &http.Client{Timeout: 5 * time.Second}

func FetchLocation() (lat, lon float64, err error) {
	resp, err := geoClient.Get("http://ip-api.com/json/?fields=lat,lon")
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	var geo geoIPResponse
	if err := json.NewDecoder(resp.Body).Decode(&geo); err != nil {
		return 0, 0, err
	}
	return geo.Lat, geo.Lon, nil
}
