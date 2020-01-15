package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var db *gorm.DB

type Entry struct {
	ID           uint      `gorm:"primary_key"`
	Timestamp    time.Time `gorm:"not null"`
	IpAddress    string    `gorm:"not null"`
	UserAgent    string
	Country      string
	CountryCode  string
	Region       string
	RegionName   string
	City         string
	Zip          string
	Latitude     float32
	Longitude    float32
	Timezone     string
	ISP          string
	Organization string
	AS_Name      string
	PrivateIP    bool
}

func (e Entry) Save() {
	db.Create(&e)
}

func (e Entry) Print() string {
	if e.PrivateIP {
		return fmt.Sprintf("IP: %s User-Agent: %s", e.IpAddress, e.UserAgent)
	} else {
		return fmt.Sprintf("IP: %s Country: %s Region: %s City: %s Latitude: %.4f Longitude: %.4f ISP: %s User-Agent: %s", e.IpAddress, e.Country, e.Region, e.City, e.Latitude, e.Longitude, e.ISP, e.UserAgent)
	}
}

func CreateEntry(r *http.Request) (*Entry, error) {
	// create an empty instance of Entry
	result := &Entry{}
	// fill the Entry
	result.Timestamp = time.Now()
	result.IpAddress = getClientIPAddress(r)
	result.UserAgent = r.UserAgent()
	// get the geoip results
	geo, err := getGeoIP(result.IpAddress)
	if err != nil {
		return nil, err
	}
	if geo.Status == "success" {
		result.Country = geo.Country
		result.CountryCode = geo.CountryCode
		result.Region = geo.Region
		result.RegionName = geo.RegionName
		result.City = geo.City
		result.Zip = geo.Zip
		result.Latitude = geo.Latitude
		result.Longitude = geo.Longitude
		result.Timezone = geo.Timezone
		result.ISP = geo.ISP
		result.Organization = geo.Organization
		result.AS_Name = geo.AS_Name
		result.PrivateIP = false
		// return the finished entry
		return result, nil
	} else if geo.Status == "fail" && geo.Message == "private range" {
		result.PrivateIP = true
		return result, nil
	}
	return nil, fmt.Errorf("Unexpected result in creating Entry")
}

func InitDB() {
	var err error
	db, err = gorm.Open("sqlite3", "/home/keybase/proof/hits.db")
	if err != nil {
		log.LogPanic("failed to connect to database")
	}
	db.AutoMigrate(&Entry{})
}
