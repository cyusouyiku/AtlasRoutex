package valueobject

import "fmt"

// Continent / Country / City 表示目的地的基础地理枚举值。
type Continent string
type Country string
type City string

const (
	ContinentNorthAmerica Continent = "NA"
	ContinentSouthAmerica Continent = "SA"
	ContinentEurope       Continent = "EU"
	ContinentAsia         Continent = "AS"
	ContinentAfrica       Continent = "AF"
	ContinentOceania      Continent = "OC"
)

const (
	CountryUSA       Country = "USA"
	CountryCanada    Country = "Canada"
	CountryUK        Country = "UK"
	CountryFrance    Country = "France"
	CountryGermany   Country = "Germany"
	CountryChina     Country = "China"
	CountryJapan     Country = "Japan"
	CountrySingapore Country = "Singapore"
	CountryAustralia Country = "Australia"
)

const (
	CityBeijing    City = "Beijing"
	CityShanghai   City = "Shanghai"
	CityTokyo      City = "Tokyo"
	CitySingaporeC City = "Singapore"
	CityLondon     City = "London"
	CityParis      City = "Paris"
	CityBerlin     City = "Berlin"
	CityNewYork    City = "New York"
	CityToronto    City = "Toronto"
	CitySydney     City = "Sydney"
)

// Region 是一个简单的地理位置值对象。
type Region struct {
	Continent Continent `json:"continent"`
	Country   Country   `json:"country"`
	City      City      `json:"city"`
}

// Validate 校验地理层级是否合法。
func (r Region) Validate() error {
	if r.Continent == "" || r.Country == "" || r.City == "" {
		return fmt.Errorf("continent, country and city are required")
	}
	return nil
}

func (r Region) String() string {
	return fmt.Sprintf("%s/%s/%s", r.Continent, r.Country, r.City)
}
