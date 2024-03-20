package types

type Place struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Address  string  `json:"address"`
	Phone    string  `json:"phone"`
	Location GeoJSON `json:"location"`
}

type GeoJSON struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
}

type PageData struct {
	Total     int
	Places    []Place
	HasPrev   bool
	PrevPage  int
	HasNext   bool
	NextPage  int
	LastPage  int
	FirstPage int
}

type Recommendation struct {
	Name   string  `json:"name"`
	Places []Place `json:"places"`
	Total  int     `json:"total"`
}
