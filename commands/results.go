package commands

import (
	"f1cli/data"
	"fmt"
	"strings"
)

// ANSI color codes
const (
	ResultsReset   = "\033[0m"
	ResultsBold    = "\033[1m"
	ResultsRed     = "\033[31m"
	ResultsGreen   = "\033[32m"
	ResultsYellow  = "\033[33m"
	ResultsBlue    = "\033[34m"
	ResultsMagenta = "\033[35m"
	ResultsCyan    = "\033[36m"
	ResultsWhite   = "\033[37m"
)

func Results(args []string, dataService *data.DataService) {
	if len(args) == 0 {
		ShowResultsHelp()
		return
	}

	// Check for help flag
	for _, arg := range args {
		if arg == "-help" || arg == "--help" {
			ShowResultsHelp()
			return
		}
	}

	location := args[0]
	sessionType := "Race" // Default to race

	// Check if sprint is specified
	if len(args) > 1 && args[1] == "sprint" {
		sessionType = "Sprint"
	}

	client := data.NewAPIClient()

	sessions, err := client.GetAllRaceAndSprintSessions()
	if err != nil {
		fmt.Printf("Error getting sessions: %v\n", err)
		return
	}

	var targetSession *data.OpenF1Session
	for _, session := range sessions {
		// Match by location (case-insensitive) and session type
		if contains(session.Location, location) && session.SessionName == sessionType {
			targetSession = &session
			break
		}
	}

	if targetSession == nil {
		fmt.Printf("No %s session found for location: %s\n", sessionType, location)
		fmt.Println("\nAvailable locations:")
		seen := make(map[string]bool)
		for _, session := range sessions {
			if !seen[session.Location] {
				fmt.Printf("  - %s\n", session.Location)
				seen[session.Location] = true
			}
		}
		return
	}

	results, err := client.GetSessionResults(targetSession.SessionKey)
	if err != nil {
		fmt.Printf("Error getting results for %s %s: %v\n", location, sessionType, err)
		return
	}

	// Sort results by position
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Position < results[i].Position {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	drivers, err := client.GetDrivers()
	if err != nil {
		fmt.Printf("Error getting drivers: %v\n", err)
		return
	}

	driverNames := make(map[int]string)
	driverTeams := make(map[int]string)
	for _, driver := range drivers {
		driverNames[driver.Number] = driver.Name
		driverTeams[driver.Number] = driver.Team
	}

	// Display results with enhanced formatting
	sessionIcon := "Race"
	if sessionType == "Sprint" {
		sessionIcon = "Sprint"
	}

	fmt.Printf("%s%s %s %s Results%s - %s%s%s\n",
		ResultsBold+ResultsYellow, sessionIcon, targetSession.Location, sessionType, ResultsReset,
		ResultsCyan, targetSession.DateStart.Format("2006-01-02"), ResultsReset)
	fmt.Printf("%s%s%s\n", ResultsBold, strings.Repeat("═", 80), ResultsReset)

	fmt.Printf("%s%-3s %-25s %-20s %-8s%s\n",
		ResultsBold+ResultsWhite, "POS", "DRIVER", "TEAM", "NUMBER", ResultsReset)
	fmt.Printf("%s%s%s\n", ResultsBold, strings.Repeat("─", 80), ResultsReset)

	// Determine points system
	var pointsSystem map[int]int
	if sessionType == "Sprint" {
		pointsSystem = data.SprintPointsSystem
	} else {
		pointsSystem = data.PointsSystem
	}

	for i, result := range results {
		driverName := driverNames[result.DriverNumber]
		if driverName == "" {
			driverName = fmt.Sprintf("Driver #%d", result.DriverNumber)
		}

		teamName := driverTeams[result.DriverNumber]
		if teamName == "" {
			teamName = "Unknown Team"
		}

		// Check if driver was disqualified
		dsqDrivers, sessionHasDSQ := data.DisqualifiedDrivers[targetSession.SessionKey]
		isDisqualified := false
		if sessionHasDSQ {
			for _, dsqDriver := range dsqDrivers {
				if dsqDriver == result.DriverNumber {
					isDisqualified = true
					break
				}
			}
		}

		points := 0
		if !isDisqualified {
			finalPosition := result.Position

			// Check if driver's position was adjusted due to DSQs
			if adjustments, hasAdjustments := data.PositionAdjustments[targetSession.SessionKey]; hasAdjustments {
				if newPosition, wasAdjusted := adjustments[result.DriverNumber]; wasAdjusted {
					finalPosition = newPosition
				}
			}

			if p, hasPoints := pointsSystem[finalPosition]; hasPoints {
				points = p
			}
		}

		// Color coding for positions
		var posColor string
		switch {
		case result.Position == 1:
			posColor = ResultsBold + ResultsYellow // Gold for winner
		case result.Position <= 3:
			posColor = ResultsBold + ResultsWhite // Silver/Bronze for podium
		case result.Position <= 10 && !isDisqualified:
			posColor = ResultsGreen // Green for points
		case isDisqualified:
			posColor = ResultsRed // Red for DSQ
		default:
			posColor = ResultsReset // Normal for no points
		}

		// Team colors
		teamColor := getResultsTeamColor(teamName)

		fmt.Printf("%s%-3d%s %-25s %s%-20s%s %s#%-6d%s",
			posColor, result.Position, ResultsReset,
			truncateString(driverName, 25),
			teamColor, truncateString(teamName, 20), ResultsReset,
			ResultsCyan, result.DriverNumber, ResultsReset)

		if isDisqualified {
			fmt.Printf(" %s(DSQ)%s", ResultsRed+ResultsBold, ResultsReset)
		} else if points > 0 {
			pointsColor := ""
			if points >= 15 {
				pointsColor = ResultsBold + ResultsYellow
			} else if points >= 8 {
				pointsColor = ResultsBold + ResultsGreen
			} else {
				pointsColor = ResultsGreen
			}
			fmt.Printf(" %s(%d pts)%s", pointsColor, points, ResultsReset)
		}
		fmt.Println()

		// Add visual separators
		if i == 2 { // After podium
			fmt.Printf("%s%s%s\n", ResultsCyan, strings.Repeat("┄", 80), ResultsReset)
		} else if result.Position == 10 && sessionType == "Race" { // After points in race
			fmt.Printf("%s%s%s\n", ResultsMagenta, strings.Repeat("┄", 80), ResultsReset)
		} else if result.Position == 8 && sessionType == "Sprint" { // After points in sprint
			fmt.Printf("%s%s%s\n", ResultsMagenta, strings.Repeat("┄", 80), ResultsReset)
		}
	}

	fmt.Printf("\n%sTotal finishers: %d%s\n", ResultsBold+ResultsCyan, len(results), ResultsReset)
	if sessionType == "Race" {
		fmt.Printf("%sPoints:%s 25-18-15-12-10-8-6-4-2-1 (positions 1-10)\n", ResultsGreen, ResultsReset)
	} else {
		fmt.Printf("%sSprint Points:%s 8-7-6-5-4-3-2-1 (positions 1-8)\n", ResultsYellow, ResultsReset)
	}
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			findInString(s, substr))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Helper function to truncate strings
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// getResultsTeamColor returns ANSI color codes for different F1 teams
func getResultsTeamColor(team string) string {
	switch team {
	case "McLaren":
		return "\033[38;5;208m" // Orange
	case "Red Bull Racing":
		return "\033[38;5;27m" // Blue
	case "Ferrari":
		return ResultsRed
	case "Mercedes":
		return "\033[38;5;51m" // Cyan
	case "Aston Martin":
		return ResultsGreen
	case "Alpine":
		return "\033[38;5;129m" // Pink
	case "Williams":
		return ResultsBlue
	case "Haas F1 Team", "Haas":
		return "\033[38;5;245m" // Gray
	case "Kick Sauber":
		return "\033[38;5;46m" // Bright Green
	case "Racing Bulls":
		return "\033[38;5;63m" // Purple
	default:
		return ResultsReset
	}
}

func ShowResultsHelp() {
	fmt.Printf("%sF1 Race Results%s\n", ResultsBold+ResultsYellow, ResultsReset)
	fmt.Printf("%s%s%s\n", ResultsBold, strings.Repeat("═", 50), ResultsReset)
	fmt.Println()
	fmt.Printf("%sUsage:%s\n", ResultsBold+ResultsGreen, ResultsReset)
	fmt.Printf("  %sf1 results <location> [session_type]%s\n", ResultsCyan, ResultsReset)
	fmt.Println()
	fmt.Printf("%sArguments:%s\n", ResultsBold+ResultsGreen, ResultsReset)
	fmt.Printf("  %slocation%s       Circuit location (e.g., Shanghai, Monaco, Silverstone)\n", ResultsYellow, ResultsReset)
	fmt.Printf("  %ssession_type%s   'race' (default) or 'sprint'\n", ResultsYellow, ResultsReset)
	fmt.Println()
	fmt.Printf("%sExamples:%s\n", ResultsBold+ResultsGreen, ResultsReset)
	fmt.Printf("  %sf1 results Shanghai%s           # Show Shanghai main race results\n", ResultsCyan, ResultsReset)
	fmt.Printf("  %sf1 results Shanghai sprint%s    # Show Shanghai sprint results\n", ResultsCyan, ResultsReset)
	fmt.Printf("  %sf1 results Monaco%s             # Show Monaco race results\n", ResultsCyan, ResultsReset)
	fmt.Printf("  %sf1 results Miami sprint%s       # Show Miami sprint results\n", ResultsCyan, ResultsReset)
	fmt.Println()
	fmt.Printf("%sNote:%s Available locations include Shanghai, Melbourne, Miami, Monaco, etc.\n",
		ResultsBold+ResultsMagenta, ResultsReset)
	fmt.Printf("      %sPoints are shown with DSQ (disqualification) indicators%s\n",
		ResultsGreen, ResultsReset)
}
