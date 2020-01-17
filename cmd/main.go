package main

import (
	"fmt"
	"github.com/Joker/jade"
	"github.com/go-chi/chi"
	"github.com/joeshaw/envdecode"
	"github.com/thehungrysmurf/new-in-town/directions"
	"github.com/thehungrysmurf/new-in-town/gpx"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

type server struct {
	publicDir string
	templatesDir string
	apiKey string
	directionsAPIUrl string
}

func main() {
	var cfg Config
	err := envdecode.StrictDecode(&cfg)
	if err != nil {
		fmt.Printf("Error decoding env: %s", err)
	}

	workDir, _ := os.Getwd()

	s := server{
		publicDir:    filepath.Join(workDir, cfg.PublicDir),
		templatesDir: filepath.Join(workDir, cfg.TemplatesDir),
		apiKey:       cfg.GmapsAPIKey,
		directionsAPIUrl: cfg.DirectionsAPIUrl,
	}

	r := chi.NewRouter()

	// define handler that serves HTTP requests with the content of the static assets
	staticAssetsDir := filepath.Join(workDir, cfg.PublicDir)
	fs := http.FileServer(http.Dir(staticAssetsDir))
	// strip the prefix so the path isn't duplicated, which would return an error
	fs = http.StripPrefix("/"+cfg.PublicDir, fs)

	// serve static assets
	r.Get("/static/public/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))

	r.Get("/", s.handleIndex)
	r.Get("/about", s.handleAbout)
	r.Get("/faq", s.handleFAQ)
	r.Get("/attractions", s.handleAttractions)
	r.Post("/directions", s.handleDirections)

	r.Get("/download/{filename}", s.handleDownload)

	r.Get("/you_are_here", s.handleYouAreHere)
	r.Get("/directions_map/{starting_point_id}/{waypoints}", s.handleDirectionsMap)

	// start listener
	err = http.ListenAndServe(":"+cfg.Port, r)
	if err != nil {
		log.Fatal(err)
	}
}

func(s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	htmlTpl, err := pugToHTML(filepath.Join(s.templatesDir, "index.pug"))
	if err != nil {
		fmt.Printf("Error creating HTML template: %s", err)
		return
	}

	err = htmlTpl.Execute(w, nil)
	if err != nil {
		fmt.Printf("Error executing HTML template: %s", err)
		return
	}
}

func(s *server) handleAbout(w http.ResponseWriter, r *http.Request) {
	htmlTpl, err := pugToHTML(filepath.Join(s.templatesDir, "about.pug"))
	if err != nil {
		fmt.Printf("Error creating HTML template: %s", err)
		return
	}

	err = htmlTpl.Execute(w, nil)
	if err != nil {
		fmt.Printf("Error executing HTML template: %s", err)
		return
	}
}

func(s *server) handleFAQ(w http.ResponseWriter, r *http.Request) {
	htmlTpl, err := pugToHTML(filepath.Join(s.templatesDir, "faq.pug"))
	if err != nil {
		fmt.Printf("Error creating HTML template: %s", err)
		return
	}

	err = htmlTpl.Execute(w, nil)
	if err != nil {
		fmt.Printf("Error executing HTML template: %s", err)
		return
	}
}

func(s *server) handleAttractions(w http.ResponseWriter, r *http.Request) {
	htmlTpl, err := pugToHTML(filepath.Join(s.templatesDir, "attractions.pug"))
	if err != nil {
		fmt.Printf("Error creating HTML template: %s", err)
		return
	}

	err = htmlTpl.Execute(w, s.apiKey)
	if err != nil {
		fmt.Printf("Error executing HTML template: %s", err)
	}
}

type DirectionsViewData struct {
	RouteInfo *directions.OptimizedRouteInfo
	GmapsAPIKey string
	GpxFilename string
}

func(s *server) handleDirections(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Error reading request body: %s", err)
		return
	}

	queryMap, err := url.ParseQuery(string(b))
	if err != nil {
		fmt.Printf("Error parsing request query: %s", err)
		return
	}

	routeInfo, err := directions.GetRoute(queryMap, s.directionsAPIUrl, s.apiKey)
	if err != nil {
		fmt.Printf("Error retrieving directions: %s", err)
		return
	}

	g := gpx.Create(routeInfo.APIResponse)
	gpxFilename, err := gpx.WriteFile(g)
	if err != nil {
		fmt.Printf("Error writing gpx file: %s", err)
		return
	}

	viewData := DirectionsViewData{
		RouteInfo: routeInfo,
		GmapsAPIKey:        s.apiKey,
		GpxFilename:        gpxFilename,
	}

	htmlTpl, err := pugToHTML(filepath.Join(s.templatesDir, "directions.pug"))
	if err != nil {
		fmt.Printf("Error creating HTML template: %s", err)
		return
	}

	err = htmlTpl.Execute(w, &viewData)
	if err != nil {
		fmt.Printf("Error executing HTML template: %s", err)
		return
	}
}

func(s *server) handleDownload(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))

	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening gpx file: %s", err)
		return
	}
	_, err = io.Copy(w, f)
	if err != nil {
		fmt.Printf("Error writing gpx file: %s", err)
		return
	}
}

func(s *server) handleYouAreHere(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, fmt.Sprintf("https://www.google.com/maps/embed/v1/place?key=%s&q=current+location"), http.StatusMovedPermanently)
}

func(s *server) handleDirectionsMap(w http.ResponseWriter, r *http.Request) {
	startingPt := chi.URLParam(r, "starting_point_id")
	waypts := chi.URLParam(r, "waypoints")

	url := fmt.Sprintf("https://www.google.com/maps/embed/v1/directions?key=%s&origin=place_id:%s&destination=place_id:%s&mode=walking&waypoints=place_id:%s", s.apiKey, startingPt, startingPt, waypts)

	http.Redirect(w, r, url, http.StatusMovedPermanently)
}

func pugToHTML(inputPath string) (*template.Template, error) {
	strTpl, err := jade.ParseFile(inputPath)
	if err != nil {
		fmt.Printf("Error parsing pug template: %s", err)
		return nil, err
	}

	htmlTpl, err := template.New("template").Parse(strTpl)
	if err != nil {
		fmt.Printf("Error creating html template: %s", err)
		return nil, err
	}

	return htmlTpl, nil
}
