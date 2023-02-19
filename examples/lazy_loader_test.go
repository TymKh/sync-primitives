package examples

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"async-task-manager/spike"
)

// {"postal_code":"1010","country_code":"AT","place_name":"Wien, Innere Stadt","state":"Wien","lat":"48.20770000","lng":"16.37050000"}
type ZipCodeInfo struct {
	PostalCode  string `json:"postal_code"`
	CountryCode string `json:"country_code"`
	PlaceName   string `json:"place_name"`
	State       string `json:"state"`
	Lat         string `json:"lat"`
	Lng         string `json:"lng"`
}

// Sometime you don't want to bother with preloading information, but rather catch it on the fly
// Given that this information is static, you could use `spike.Manager` with expiration set to -1 which
// would handle everything gracefully for you, requesting once on the flight and caching for lifetime of your service
func TestLazyLoader(t *testing.T) {
	//if we pass -1 as a cache duration, then data received will be cached forever
	fetchZipCodeInfo := func(ctx context.Context, zipCode string) (*ZipCodeInfo, error) {
		resp, err := http.Get("https://zip-api.eu/api/v1/info/" + zipCode)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		zipCodeInfo := ZipCodeInfo{}
		err = json.NewDecoder(resp.Body).Decode(&zipCodeInfo)
		if err != nil {
			return nil, err
		}
		return &zipCodeInfo, nil
	}

	sm := spike.NewManager(fetchZipCodeInfo, -1)
	zci, err := sm.GetResult(context.Background(), "AT-1010")
	if err != nil {
		t.Errorf("error: %v", err)
		return
	}
	t.Logf("zip code info: %+v", zci)

	zci, err = sm.GetResult(context.Background(), "DE-01067")
	if err != nil {
		t.Errorf("error: %v", err)
		return
	}
	t.Logf("zip code info: %+v", zci)

	zci, err = sm.GetResult(context.Background(), "FR-31390")
	if err != nil {
		t.Errorf("error: %v", err)
		return
	}
	t.Logf("zip code info: %+v", zci)

}
