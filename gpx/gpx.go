package gpx

import (
	"bytes"
	"encoding/xml"
	"github.com/thehungrysmurf/new-in-town/directions"
	"github.com/twpayne/go-gpx"
	"io/ioutil"
	"strconv"
	"time"
)

// Create assembles a gpx struct from data returned by the Directions API
func Create(dr *directions.APIResponse) *gpx.GPX {
	var mainWpts []*gpx.WptType
	var trkWpts []*gpx.WptType

	for _, leg := range dr.Routes[0].Legs {
		nw := gpx.WptType{
			Lat: leg.StartLocation.Lat,
			Lon: leg.StartLocation.Lng,
		}
		mainWpts = append(mainWpts, &nw)
	}

	for _, leg := range dr.Routes[0].Legs {
		for _, step := range leg.Steps {
			nw := gpx.WptType{
				Lat: step.EndLocation.Lat,
				Lon: step.EndLocation.Lng,
			}
			trkWpts = append(trkWpts, &nw)
		}
	}

	g := &gpx.GPX{
		Version: "1.1",
		Creator: "New In Town",
		Wpt:     mainWpts,
		Trk: []*gpx.TrkType{
			{
				TrkSeg: []*gpx.TrkSegType{
					{
						TrkPt: trkWpts,
					},
				},
			},
		},
	}

	return g
}

// WriteFile writes a gpx struct to a local file with open permissions
func WriteFile(g *gpx.GPX) (string, error) {
	var xmlBuffer bytes.Buffer

	_, err := xmlBuffer.Write([]byte(xml.Header))
	if err != nil {
		return "", err
	}

	err = g.WriteIndent(&xmlBuffer, "", " ")
	if err != nil {
		return "", err
	}

	t := time.Now().UnixNano()
	gpxFilename := "new_in_town_" + strconv.FormatInt(t, 10) + ".gpx"
	err = ioutil.WriteFile(gpxFilename, xmlBuffer.Bytes(), 777)
	if err != nil {
		return "", err
	}

	return gpxFilename, nil
}
