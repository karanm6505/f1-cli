package data

import "time"

// Driver represents a Formula 1 driver
type Driver struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Number        int    `json:"number"`
	Team          string `json:"team"`
	Country       string `json:"country"`
	Points        int    `json:"points"`
	Wins          int    `json:"wins"`
	Podiums       int    `json:"podiums"`
	Championships int    `json:"championships"`
}

type Team struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	Country       string   `json:"country"`
	Points        int      `json:"points"`
	Drivers       []string `json:"drivers"`
	Founded       int      `json:"founded"`
	Championships int      `json:"championships"`
}

type Race struct {
	Round        int       `json:"round"`
	Name         string    `json:"name"`
	Circuit      string    `json:"circuit"`
	Country      string    `json:"country"`
	Date         time.Time `json:"date"`
	Time         string    `json:"time"`
	Status       string    `json:"status"`
	Winner       string    `json:"winner,omitempty"`
	PolePosition string    `json:"pole_position,omitempty"`
	FastestLap   string    `json:"fastest_lap,omitempty"`
}

type Circuit struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Country      string  `json:"country"`
	Length       float64 `json:"length_km"`
	Turns        int     `json:"turns"`
	LapRecord    string  `json:"lap_record"`
	RecordHolder string  `json:"record_holder"`
	FirstRace    int     `json:"first_race_year"`
}

type StandingEntry struct {
	Position int    `json:"position"`
	Driver   string `json:"driver"`
	Team     string `json:"team"`
	Points   int    `json:"points"`
	Wins     int    `json:"wins"`
	Gap      string `json:"gap"`
}
