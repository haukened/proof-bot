package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

var baseurl = "http://ip-api.com/json/"

type geoApiResponse struct {
	Status       string  `json:"status"`
	Message      string  `json:"message"`
	Country      string  `json:"country"`
	CountryCode  string  `json:"countryCode"`
	Region       string  `json:"region"`
	RegionName   string  `json:"regionName"`
	City         string  `json:"city"`
	Zip          string  `json:"zip"`
	Latitude     float32 `json:"lat"`
	Longitude    float32 `json:"lon"`
	Timezone     string  `json:"timezone"`
	ISP          string  `json:"isp"`
	Organization string  `json:"org"`
	AS_Name      string  `json:"as"`
	Query        string  `json:"query"`
}

func getGeoIP(ipAddress string) (*geoApiResponse, error) {
	// create a client
	client := &http.Client{}

	// create a query
	req, err := http.NewRequest("GET", baseurl+ipAddress, nil)
	if err != nil {
		return nil, err
	}
	// set a user agent to be polite
	req.Header.Set("User-Agent", "Golang_IP_Discovery/0.99")
	// make the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// read the body of the request
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// unpack the json
	geoResponse := new(geoApiResponse)
	if err := json.Unmarshal(body, geoResponse); err != nil {
		return nil, err
	}
	return geoResponse, nil
}
