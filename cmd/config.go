package main

type Config struct {
	DirectionsAPIUrl string `env:"DIRECTIONS_API_URL,default=https://maps.googleapis.com/maps/api/directions/json"`
	GmapsAPIKey string `env:"GMAPS_API_KEY,required"`

	PublicDir string `env:"STATIC_DIR,default=static/public"`
	TemplatesDir string `env:"STATIC_DIR,default=static/views"`
}
