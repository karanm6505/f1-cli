package data

import (
	"fmt"
	"log"
	"strings"
)

// DataService provides F1 data from OpenF1 API
type DataService struct {
	apiClient *APIClient
}

func NewDataService() *DataService {
	return &DataService{
		apiClient: NewAPIClient(),
	}
}

func (ds *DataService) GetDriverStandings() ([]StandingEntry, error) {
	return ds.apiClient.GetCurrentDriverStandings()
}

func (ds *DataService) GetConstructorStandings() ([]StandingEntry, error) {
	return ds.apiClient.GetCurrentConstructorStandings()
}

func (ds *DataService) GetDrivers() ([]Driver, error) {
	drivers, err := ds.apiClient.GetDrivers()
	if err != nil {
		return nil, err
	}

	// Enrich with standings data
	standings, err := ds.apiClient.GetCurrentDriverStandings()
	if err != nil {
		log.Printf("Warning: Could not get standings data: %v", err)
		return drivers, nil
	}

	// Merge driver info with standings
	for i := range drivers {
		for _, standing := range standings {
			if drivers[i].Name == standing.Driver {
				drivers[i].Points = standing.Points
				drivers[i].Wins = standing.Wins
				drivers[i].Team = standing.Team
				break
			}
		}
	}

	return drivers, nil
}

// GetDriverByName finds a driver by name from OpenF1 API
func (ds *DataService) GetDriverByName(name string) (*Driver, error) {
	drivers, err := ds.GetDrivers()
	if err != nil {
		return nil, err
	}

	for _, driver := range drivers {
		if strings.EqualFold(driver.Name, name) {
			return &driver, nil
		}
	}

	return nil, fmt.Errorf("driver '%s' not found", name)
}

// GetRaceSchedule returns race schedule from OpenF1 API
func (ds *DataService) GetRaceSchedule() ([]Race, error) {
	return ds.apiClient.GetCurrentRaceSchedule()
}

// GetNextRace returns the next upcoming race from OpenF1 API
func (ds *DataService) GetNextRace() (*Race, error) {
	races, err := ds.GetRaceSchedule()
	if err != nil {
		return nil, err
	}

	for _, race := range races {
		if race.Status == "upcoming" {
			return &race, nil
		}
	}

	return nil, fmt.Errorf("no upcoming races found")
}

// GetLastRace returns the most recent completed race from OpenF1 API
func (ds *DataService) GetLastRace() (*Race, error) {
	races, err := ds.GetRaceSchedule()
	if err != nil {
		return nil, err
	}

	var lastRace *Race
	for _, race := range races {
		if race.Status == "completed" {
			lastRace = &race
		}
	}

	if lastRace != nil {
		return lastRace, nil
	}

	return nil, fmt.Errorf("no completed races found")
}

// GetSourceName returns the name of the data source
func (ds *DataService) GetSourceName() string {
	return "OpenF1 API"
}

// IsOnline checks if the API source is accessible
func (ds *DataService) IsOnline() bool {
	_, err := ds.apiClient.makeRequest("drivers?session_key=latest")
	return err == nil
}

// GetAPIClient returns the underlying API client for advanced operations
func (ds *DataService) GetAPIClient() *APIClient {
	return ds.apiClient
}
