package data

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// APIClient connects to the OpenF1 API.
// Note: OpenF1 provides raw race results - we calculate championship standings ourselves.
type APIClient struct {
	BaseURL string
	Client  *http.Client
}

func NewAPIClient() *APIClient {
	return &APIClient{
		BaseURL: "https://api.openf1.org/v1",
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// OpenF1 API response structures - field names must match API exactly for JSON unmarshaling
type OpenF1Driver struct {
	DriverNumber  int    `json:"driver_number"`
	BroadcastName string `json:"broadcast_name"`
	FullName      string `json:"full_name"`
	NameAcronym   string `json:"name_acronym"`
	TeamName      string `json:"team_name"`
	TeamColour    string `json:"team_colour"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	HeadshotUrl   string `json:"headshot_url"`
	CountryCode   string `json:"country_code"`
}

type OpenF1Session struct {
	SessionKey       int       `json:"session_key"`
	SessionName      string    `json:"session_name"`
	DateStart        time.Time `json:"date_start"`
	DateEnd          time.Time `json:"date_end"`
	GmtOffset        string    `json:"gmt_offset"`
	SessionType      string    `json:"session_type"`
	MeetingKey       int       `json:"meeting_key"`
	Location         string    `json:"location"`
	CountryKey       int       `json:"country_key"`
	CountryCode      string    `json:"country_code"`
	CountryName      string    `json:"country_name"`
	CircuitKey       int       `json:"circuit_key"`
	CircuitShortName string    `json:"circuit_short_name"`
	Year             int       `json:"year"`
}

type OpenF1Meeting struct {
	CircuitKey          int       `json:"circuit_key"`
	CircuitShortName    string    `json:"circuit_short_name"`
	CountryCode         string    `json:"country_code"`
	CountryKey          int       `json:"country_key"`
	CountryName         string    `json:"country_name"`
	DateStart           time.Time `json:"date_start"`
	GmtOffset           string    `json:"gmt_offset"`
	Location            string    `json:"location"`
	MeetingKey          int       `json:"meeting_key"`
	MeetingName         string    `json:"meeting_name"`
	MeetingOfficialName string    `json:"meeting_official_name"`
	Year                int       `json:"year"`
}

type OpenF1Position struct {
	Date         time.Time `json:"date"`
	DriverNumber int       `json:"driver_number"`
	MeetingKey   int       `json:"meeting_key"`
	Position     int       `json:"position"`
	SessionKey   int       `json:"session_key"`
}

type OpenF1RaceResult struct {
	DriverNumber int    `json:"driver_number"`
	Position     int    `json:"position"`
	Points       int    `json:"points"`
	Status       string `json:"status"`
	SessionKey   int    `json:"session_key"`
	MeetingKey   int    `json:"meeting_key"`
}

func (c *APIClient) makeRequest(endpoint string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", c.BaseURL, endpoint)

	resp, err := c.Client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return body, nil
}

func (c *APIClient) GetDrivers() ([]Driver, error) {
	data, err := c.makeRequest("drivers?session_key=latest")
	if err != nil {
		return nil, err
	}

	var openF1Drivers []OpenF1Driver
	if err := json.Unmarshal(data, &openF1Drivers); err != nil {
		return nil, fmt.Errorf("failed to parse drivers response: %w", err)
	}

	// Remove duplicates and convert to our Driver struct
	seen := make(map[int]bool)
	result := []Driver{}

	for _, driver := range openF1Drivers {
		if seen[driver.DriverNumber] {
			continue
		}
		seen[driver.DriverNumber] = true

		result = append(result, Driver{
			ID:            driver.DriverNumber,
			Name:          driver.FullName,
			Number:        driver.DriverNumber,
			Team:          driver.TeamName,
			Country:       driver.CountryCode,
			Points:        0,
			Wins:          0,
			Podiums:       0,
			Championships: 0,
		})
	}

	return result, nil
}

// F1 Points Systems
var PointsSystem = map[int]int{
	1: 25, 2: 18, 3: 15, 4: 12, 5: 10, 6: 8, 7: 6, 8: 4, 9: 2, 10: 1,
}

var SprintPointsSystem = map[int]int{
	1: 8, 2: 7, 3: 6, 4: 5, 5: 4, 6: 3, 7: 2, 8: 1,
}

// Disqualified drivers for specific sessions
var DisqualifiedDrivers = map[int][]int{
	9998:  {16, 44, 10}, // Shanghai Race: Leclerc, Hamilton, Gasly
	10028: {23, 30, 87}, // Miami Sprint: Albon, Lawson, Bearman
}

// Position adjustments due to disqualifications - when drivers are DSQ'd,
// everyone behind them moves up and gets points for their new position
var PositionAdjustments = map[int]map[int]int{
	9693: {
		12: 4, // Antonelli: 5th → 4th
		23: 5, // Albon: 4th → 5th
	},
	9998: {
		31: 5,  // Ocon: 7th → 5th
		12: 6,  // Antonelli: 8th → 6th
		23: 7,  // Albon: 9th → 7th
		87: 8,  // Bearman: 10th → 8th
		18: 9,  // Stroll: 12th → 9th
		55: 10, // Sainz: 13th → 10th
	},
	10028: {
		63: 4, // Russell: 5th → 4th
		18: 5, // Stroll: 6th → 5th
		22: 6, // Tsunoda: 7th → 6th
		12: 7, // Antonelli: 8th → 7th
		10: 8, // Gasly: 9th → 8th
	},
}

// StandingData holds points and wins for standings calculation
type StandingData struct {
	Points int
	Wins   int
}

func (c *APIClient) GetRaceSessions() ([]OpenF1Session, error) {
	data, err := c.makeRequest("sessions?session_type=Race&year=2025")
	if err != nil {
		return nil, err
	}

	var allSessions []OpenF1Session
	if err := json.Unmarshal(data, &allSessions); err != nil {
		return nil, fmt.Errorf("failed to parse sessions response: %w", err)
	}

	// Filter for sessions with session_name "Race" (exclude "Sprint")
	var raceSessions []OpenF1Session
	for _, session := range allSessions {
		if session.SessionName == "Race" {
			raceSessions = append(raceSessions, session)
		}
	}

	return raceSessions, nil
}

func (c *APIClient) GetSprintSessions() ([]OpenF1Session, error) {
	data, err := c.makeRequest("sessions?session_type=Race&year=2025")
	if err != nil {
		return nil, err
	}

	var allSessions []OpenF1Session
	if err := json.Unmarshal(data, &allSessions); err != nil {
		return nil, fmt.Errorf("failed to parse sessions response: %w", err)
	}

	// Filter for sessions with session_name "Sprint"
	var sprintSessions []OpenF1Session
	for _, session := range allSessions {
		if session.SessionName == "Sprint" {
			sprintSessions = append(sprintSessions, session)
		}
	}

	return sprintSessions, nil
}

func (c *APIClient) GetAllRaceAndSprintSessions() ([]OpenF1Session, error) {
	raceSessions, err := c.GetRaceSessions()
	if err != nil {
		return nil, err
	}

	sprintSessions, err := c.GetSprintSessions()
	if err != nil {
		return raceSessions, nil
	}

	allSessions := append(raceSessions, sprintSessions...)
	return allSessions, nil
}

func (c *APIClient) GetSessionResults(sessionKey int) ([]OpenF1Position, error) {
	endpoint := fmt.Sprintf("position?session_key=%d", sessionKey)
	data, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var positions []OpenF1Position
	if err := json.Unmarshal(data, &positions); err != nil {
		return nil, fmt.Errorf("failed to parse positions response: %w", err)
	}

	// Get final positions only (latest timestamp for each driver)
	finalPositions := make(map[int]OpenF1Position)
	for _, pos := range positions {
		if existing, ok := finalPositions[pos.DriverNumber]; !ok || pos.Date.After(existing.Date) {
			finalPositions[pos.DriverNumber] = pos
		}
	}

	result := make([]OpenF1Position, 0, len(finalPositions))
	for _, pos := range finalPositions {
		result = append(result, pos)
	}

	return result, nil
}

func (c *APIClient) GetCurrentRaceSchedule() ([]Race, error) {
	data, err := c.makeRequest("meetings?year=2025")
	if err != nil {
		return nil, err
	}

	var meetings []OpenF1Meeting
	if err := json.Unmarshal(data, &meetings); err != nil {
		return nil, fmt.Errorf("failed to parse meetings response: %w", err)
	}

	result := make([]Race, len(meetings))

	for i, meeting := range meetings {
		status := "upcoming"
		if meeting.DateStart.Before(time.Now()) {
			status = "completed"
		}

		result[i] = Race{
			Round:        i + 1,
			Name:         meeting.MeetingOfficialName,
			Circuit:      meeting.CircuitShortName,
			Country:      meeting.CountryName,
			Date:         meeting.DateStart,
			Time:         meeting.DateStart.Format("15:04"),
			Status:       status,
			Winner:       "",
			PolePosition: "",
			FastestLap:   "",
		}
	}

	return result, nil
}

// GetCurrentDriverStandings calculates real driver standings from race results
func (c *APIClient) GetCurrentDriverStandings() ([]StandingEntry, error) {
	// Get all race and sprint sessions for current year
	sessions, err := c.GetAllRaceAndSprintSessions()
	if err != nil {
		return nil, err
	}

	// Get all drivers first
	drivers, err := c.GetDrivers()
	if err != nil {
		return nil, err
	}

	// Initialize driver standings map
	driverPoints := make(map[int]*StandingData)
	driverNames := make(map[int]string)
	driverTeams := make(map[int]string)

	for _, driver := range drivers {
		driverPoints[driver.Number] = &StandingData{
			Points: 0,
			Wins:   0,
		}
		driverNames[driver.Number] = driver.Name
		driverTeams[driver.Number] = driver.Team
	}

	// Process each completed session (race or sprint)
	for _, session := range sessions {
		// Only process completed sessions (before current time)
		if session.DateStart.After(time.Now()) {
			continue
		}

		results, err := c.GetSessionResults(session.SessionKey)
		if err != nil {
			// Silently skip sessions that don't have results yet
			continue
		}

		// Determine which points system to use based on session name
		var pointsSystem map[int]int
		if session.SessionName == "Sprint" {
			pointsSystem = SprintPointsSystem
		} else {
			pointsSystem = PointsSystem
		}

		// Award points based on finishing position
		for _, result := range results {
			if standing, exists := driverPoints[result.DriverNumber]; exists {
				// Check if driver was disqualified from this session
				dsqDrivers, sessionHasDSQ := DisqualifiedDrivers[session.SessionKey]
				isDisqualified := false
				if sessionHasDSQ {
					for _, dsqDriver := range dsqDrivers {
						if dsqDriver == result.DriverNumber {
							isDisqualified = true
							break
						}
					}
				}

				// Only award points if not disqualified
				if !isDisqualified {
					finalPosition := result.Position

					// Check if driver's position was adjusted due to DSQs
					if adjustments, hasAdjustments := PositionAdjustments[session.SessionKey]; hasAdjustments {
						if newPosition, wasAdjusted := adjustments[result.DriverNumber]; wasAdjusted {
							finalPosition = newPosition
						}
					}

					if points, hasPoints := pointsSystem[finalPosition]; hasPoints {
						standing.Points += points
					}
					// Only count wins for main races, not sprints
					if finalPosition == 1 && session.SessionName == "Race" {
						standing.Wins++
					}
				}
			}
		}
	}

	// Convert to sorted standings
	var standings []StandingEntry
	for driverNumber, standing := range driverPoints {
		if driverName, exists := driverNames[driverNumber]; exists {
			standings = append(standings, StandingEntry{
				Driver: driverName,
				Team:   driverTeams[driverNumber],
				Points: standing.Points,
				Wins:   standing.Wins,
			})
		}
	}

	// Sort by points (highest first), then by wins
	for i := 0; i < len(standings)-1; i++ {
		for j := i + 1; j < len(standings); j++ {
			if standings[j].Points > standings[i].Points ||
				(standings[j].Points == standings[i].Points && standings[j].Wins > standings[i].Wins) {
				standings[i], standings[j] = standings[j], standings[i]
			}
		}
	}

	// Add position and gap information
	for i := range standings {
		standings[i].Position = i + 1
		if i == 0 {
			standings[i].Gap = "Leader"
		} else {
			standings[i].Gap = fmt.Sprintf("-%d", standings[0].Points-standings[i].Points)
		}
	}

	return standings, nil
}

// GetCurrentConstructorStandings calculates constructor standings from driver standings
func (c *APIClient) GetCurrentConstructorStandings() ([]StandingEntry, error) {
	// Get driver standings first
	driverStandings, err := c.GetCurrentDriverStandings()
	if err != nil {
		return nil, err
	}

	// Group drivers by team and sum points
	teamPoints := make(map[string]int)
	teamWins := make(map[string]int)
	teamCountries := make(map[string]string)

	for _, standing := range driverStandings {
		teamPoints[standing.Team] += standing.Points
		teamWins[standing.Team] += standing.Wins
		// Use first driver's team to determine country (approximate)
		if _, exists := teamCountries[standing.Team]; !exists {
			teamCountries[standing.Team] = standing.Team // Fallback to team name
		}
	}

	// Convert to sorted standings
	var standings []StandingEntry
	for team, points := range teamPoints {
		standings = append(standings, StandingEntry{
			Driver: team, // Using Driver field for team name
			Team:   teamCountries[team],
			Points: points,
			Wins:   teamWins[team],
		})
	}

	// Sort by points (highest first), then by wins
	for i := 0; i < len(standings)-1; i++ {
		for j := i + 1; j < len(standings); j++ {
			if standings[j].Points > standings[i].Points ||
				(standings[j].Points == standings[i].Points && standings[j].Wins > standings[i].Wins) {
				standings[i], standings[j] = standings[j], standings[i]
			}
		}
	}

	// Add position and gap information
	for i := range standings {
		standings[i].Position = i + 1
		if i == 0 {
			standings[i].Gap = "Leader"
		} else {
			standings[i].Gap = fmt.Sprintf("-%d", standings[0].Points-standings[i].Points)
		}
	}

	return standings, nil
}
