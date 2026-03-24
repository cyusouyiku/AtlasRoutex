package valueobject

import {
	"fmt"
}

type Continent string
type Country string
type City string

const (
	ContinentNorthAmerica Continent = "NA" // 北美洲
	ContinentSouthAmerica Continent = "SA" // 南美洲
	ContinentEurope       Continent = "EU" // 欧洲
	ContinentAsia         Continent = "AS" // 亚洲
	ContinentAfrica       Continent = "AF" // 非洲
	ContinentOceania      Continent = "OC" // 大洋洲
)

type ContinentNorthAmerica Country{
	USA  Country = "USA"
	Canada Country = "Canada"
}

type ContinentEurope Country {
	UK Country = "UK"
	France Country = "France"
	Germany Country = "Germany"
}

type ContinentAsia Country {
	China Country = "China"
	Japan Country = "Japan"
	Singapore Country = "Singapore"
}

type ContinentOceania Country {
	Australia Country = "Australia"
}

type USA City {
	NewYork City = "New York"
	LosAngeles City = "Los Angeles"
	Chicago City = "Chicago"
	Houston City = "Houston"
	Phoenix City = "Phoenix"
	Philadelphia City = "Philadelphia"
	SanAntonio City = "San Antonio"
	SanDiego City = "San Diego"
	Dallas City = "Dallas"
}

type Canada City {
	Toronto City = "Toronto"
	Vancouver City = "Vancouver"
}

type UK City {
	London City = "London"
	Manchester City = "Manchester"
	Glasgow City = "Glasgow"
	Liverpool City = "Liverpool"
	Birmingham City = "Birmingham"
	Edinburgh City = "Edinburgh"
}

type France City {
	Paris City = "Paris"
	Lyon City = "Lyon"
	Marseille City = "Marseille"
	Toulouse City = "Toulouse"
}

type Germany City {
	Berlin City = "Berlin"
	Munich City = "Munich"
	Hamburg City = "Hamburg"
	Cologne City = "Cologne"
}

type China City {
	Beijing City = "Beijing"
	Shanghai City = "Shanghai"
	Guangzhou City = "Guangzhou"
	Shenzhen City = "Shenzhen"
	Chengdu City = "Chengdu"
	XiAn City = "Xi'an"
	Dalian City = "Dalian"
	ShenYang City = "Shenyang"
	HongKong City = "Hong Kong"
	Suzhou City = "Suzhou"
	Hangzhou City = "Hangzhou"
	Nanjing City = "Nanjing"
	Wuhan City = "Wuhan"
	Chongqing City = "Chongqing"
	Tianjin City = "Tianjin"
	Ningbo City = "Ningbo"
	Qingdao City = "Qingdao"
	XiAn City = "Xi'an"
	Fuzhou City = "Fuzhou"
	Xiamen City = "Xiamen"
	Kunming City = "Kunming"
	Urumqi City = "Urumqi"
	Harbin City = "Harbin"
	Hefei City = "Hefei"
	Kunshan City = "Kunshan"
	Jinan City = "Jinan"
	Changsha City = "Changsha"
	Shijiazhuang City = "Shijiazhuang"
	Nantong City = "Nantong"
	Jiangsu City = "Jiangsu"

}