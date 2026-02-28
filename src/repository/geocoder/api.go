package geocoder

import (
	"InstantWellnessKits/src/entity"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	url = "https://geocode.googleapis.com/v4beta/geocode/location"
)

type GeocodingResponse struct {
	Results []Result `json:"results"`
}

type Result struct {
	AddressComponents []AddressComponent `json:"addressComponents"`
}

type AddressComponent struct {
	LongText string   `json:"longText"`
	Types    []string `json:"types"`
}

type Api struct {
	key    string
	client *http.Client
}

func NewApi(key string) *Api {
	return &Api{
		key:    key,
		client: http.DefaultClient,
	}
}

func (a *Api) GetJurisdiction(latitude, longitude float64) (*entity.Jurisdiction, error) {
	response, err := a.client.Get(
		fmt.Sprintf("%s?location.latitude=%f&location.longitude=%f&key=%s", url, latitude, longitude, a.key),
	)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var geocodingResponse GeocodingResponse
	if err := json.NewDecoder(response.Body).Decode(&geocodingResponse); err != nil {
		return nil, err
	}

	state, city, county, err := a.extractJurisdiction(geocodingResponse.Results)
	if err != nil {
		return nil, err
	}

	return entity.NewJurisdiction(state, city, county, ""), nil
}

func (a *Api) extractJurisdiction(results []Result) (string, string, string, error) {
	if len(results) == 0 {
		return "", "", "", fmt.Errorf("no results found")
	}

	var state, city, county string
	for _, component := range results[0].AddressComponents {
		for _, t := range component.Types {
			switch t {
			case "locality":
				county = component.LongText
			case "administrative_area_level_1":
				state = component.LongText
			case "administrative_area_level_2":
				city = strings.TrimSuffix(component.LongText, " County")
			}
		}
	}

	return state, city, county, nil
}
