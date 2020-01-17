package directions

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

type attraction struct {
	id   string
	name string
}

// Coord is the smallest building block of the Directions response, a single pair of latitude and longitude coordinates
type Coord struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// Step is a directed route between two points defined each by coordinates
type Step struct {
	StartLocation Coord `json:"start_location"`
	EndLocation   Coord `json:"end_location"`
}

// Leg is an ordered series of Steps on the route
type Leg struct {
	StartLocation Coord `json:"start_location"`
	EndLocation   Coord `json:"end_location"`
	Steps         []Step
}

// Route contains a series of Legs and the index order for the original waypoints, to achieve an optimized route
type Route struct {
	Legs          []Leg `json:"legs"`
	WaypointOrder []int `json:"waypoint_order"`
}

// APIResponse is the object representation of the response from Directions API
type APIResponse struct {
	Routes []Route `json:"routes"`
}

// OptimizedRouteInfo collects the data needed by the front end to display the route
type OptimizedRouteInfo struct {
	StartingPointName string
	StartingPointID   string
	OrderedWaypoints  string
	APIResponse       *APIResponse
}

// GetRoute queries the Google Directions API and retrieves an optimized route given a set of unordered waypoints (attractions)
func GetRoute(queryMap url.Values, directionsApiUrl, gmapsApiKey string) (*OptimizedRouteInfo, error) {
	//TODO:
	// error if attraction_id array length and attraction_name array length don't match
	// error if attraction_id or attraction_name keys is missing (or if there are fewer than 2 attractions)

	var attractions []attraction
	var attractionIDs []string

	for index, id := range queryMap["attraction_id"] {
		a := attraction{
			id:   id,
			name: queryMap["attraction_name"][index],
		}
		attractions = append(attractions, a)
		attractionIDs = append(attractionIDs, id)
	}

	startingPtID := attractionIDs[0]
	unorderedWpts := attractionIDs[1:]

	r, err := getDirections(directionsApiUrl, gmapsApiKey, startingPtID, unorderedWpts)
	if err != nil {
		return nil, err
	}

	var orderedWptIDs []string
	for _, position := range r.Routes[0].WaypointOrder {
		orderedWptIDs = append(orderedWptIDs, unorderedWpts[position])
	}

	data := &OptimizedRouteInfo{
		StartingPointName: attractions[0].name,
		StartingPointID:   startingPtID,
		OrderedWaypoints:  strings.Join(orderedWptIDs, "|place_id:"),
		APIResponse:       r,
	}

	return data, nil
}

func getDirections(directionsAPIUrl, apiKey, startingPtID string, unorderedWpts []string) (*APIResponse, error) {
	client := http.Client{}

	req, err := http.NewRequest("POST", directionsAPIUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-type", "application/json")

	qs := req.URL.Query()
	qs.Add("origin", "place_id:"+startingPtID)
	qs.Add("destination", "place_id:"+startingPtID)
	qs.Add("waypoints", "optimize:true|place_id:"+strings.Join(unorderedWpts, "|place_id:"))
	qs.Add("mode", "walking")
	qs.Add("key", apiKey)
	req.URL.RawQuery = qs.Encode()

	resp, err := client.Do(req)

	dr := APIResponse{}
	err = json.NewDecoder(resp.Body).Decode(&dr)
	if err != nil {
		return nil, err
	}

	return &dr, nil
}
