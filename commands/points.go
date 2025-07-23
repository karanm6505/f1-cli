package commands

import (
	"fmt"
	"strings"
	"time"

	"f1cli/data"
)

// ANSI color codes for points command
const (
	PointsReset   = "\033[0m"
	PointsBold    = "\033[1m"
	PointsRed     = "\033[31m"
	PointsGreen   = "\033[32m"
	PointsYellow  = "\033[33m"
	PointsBlue    = "\033[34m"
	PointsMagenta = "\033[35m"
	PointsCyan    = "\033[36m"
	PointsWhite   = "\033[37m"
)

// PointsBreakdown represents points scored by a driver in a specific session
type PointsBreakdown struct {
	RaceName    string
	Location    string
	Date        time.Time
	SessionType string
	Position    int
	Points      int
	IsAdjusted  bool // True if position was adjusted due to DSQ
}

// Points displays detailed race-by-race points breakdown for a specific driver
func Points(args []string, dataService *data.DataService) {
	if len(args) == 0 {
		fmt.Println("❌ Error: Please specify a driver name")
		fmt.Println("Usage: f1 points \"<driver_name>\"")
		fmt.Println("Example: f1 points \"Oscar Piastri\"")
		return
	}

	targetDriver := strings.Join(args, " ")
	client := dataService.GetAPIClient()

	// Get all drivers to find the target driver
	drivers, err := client.GetDrivers()
	if err != nil {
		fmt.Printf("❌ Error fetching drivers: %v\n", err)
		return
	}

	var driverNumber int
	var driverTeam string
	found := false

	for _, driver := range drivers {
		if strings.EqualFold(driver.Name, targetDriver) {
			driverNumber = driver.Number
			driverTeam = driver.Team
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("❌ Driver '%s' not found\n", targetDriver)
		fmt.Println("\nAvailable drivers:")
		for _, driver := range drivers {
			fmt.Printf("  - %s\n", driver.Name)
		}
		return
	}

	// Get all race and sprint sessions
	sessions, err := client.GetAllRaceAndSprintSessions()
	if err != nil {
		fmt.Printf("❌ Error fetching sessions: %v\n", err)
		return
	}

	var pointsBreakdown []PointsBreakdown
	totalPoints := 0
	totalWins := 0

	// Process each completed session
	for _, session := range sessions {
		// Only process completed sessions
		if session.DateStart.After(time.Now()) {
			continue
		}

		results, err := client.GetSessionResults(session.SessionKey)
		if err != nil {
			continue // Skip sessions without results
		}

		// Find this driver's result in this session
		for _, result := range results {
			if result.DriverNumber == driverNumber {
				// Check if driver was disqualified
				dsqDrivers, sessionHasDSQ := data.DisqualifiedDrivers[session.SessionKey]
				isDisqualified := false
				if sessionHasDSQ {
					for _, dsqDriver := range dsqDrivers {
						if dsqDriver == result.DriverNumber {
							isDisqualified = true
							break
						}
					}
				}

				if !isDisqualified {
					// Determine points system
					var pointsSystem map[int]int
					if session.SessionName == "Sprint" {
						pointsSystem = data.SprintPointsSystem
					} else {
						pointsSystem = data.PointsSystem
					}

					finalPosition := result.Position
					isAdjusted := false

					// Check for position adjustments
					if adjustments, hasAdjustments := data.PositionAdjustments[session.SessionKey]; hasAdjustments {
						if newPosition, wasAdjusted := adjustments[result.DriverNumber]; wasAdjusted {
							finalPosition = newPosition
							isAdjusted = true
						}
					}

					points := 0
					if p, hasPoints := pointsSystem[finalPosition]; hasPoints {
						points = p
					}

					if points > 0 || isAdjusted {
						pointsBreakdown = append(pointsBreakdown, PointsBreakdown{
							RaceName:    session.Location,
							Location:    session.Location,
							Date:        session.DateStart,
							SessionType: session.SessionName,
							Position:    finalPosition,
							Points:      points,
							IsAdjusted:  isAdjusted,
						})

						totalPoints += points
					}

					// Count wins (only for main races)
					if finalPosition == 1 && session.SessionName == "Race" {
						totalWins++
					}
				}
				break
			}
		}
	}

	// Display results with enhanced formatting
	fmt.Printf("%sPoints Breakdown - %s%s %s(#%d)%s\n",
		PointsBold+PointsYellow, targetDriver, PointsReset, PointsCyan, driverNumber, PointsReset)

	// Team color
	teamColor := getPointsTeamColor(driverTeam)
	fmt.Printf("%sTeam:%s %s%s%s\n", PointsBlue, PointsReset, teamColor, driverTeam, PointsReset)

	// Points summary with colors
	pointsColor := ""
	if totalPoints > 200 {
		pointsColor = PointsBold + PointsYellow
	} else if totalPoints > 100 {
		pointsColor = PointsBold + PointsGreen
	} else if totalPoints > 50 {
		pointsColor = PointsGreen
	}

	winsColor := ""
	if totalWins > 0 {
		winsColor = PointsBold + PointsYellow
	}

	fmt.Printf("%sTotal Points:%s %s%d%s %s| Wins:%s %s%d%s\n",
		PointsBold+PointsBlue, PointsReset, pointsColor, totalPoints, PointsReset,
		PointsBold+PointsBlue, PointsReset, winsColor, totalWins, PointsReset)
	fmt.Printf("%s%s%s\n", PointsBold, strings.Repeat("═", 80), PointsReset)

	if len(pointsBreakdown) == 0 {
		fmt.Printf("%sNo points scored yet in the 2025 season.%s\n", PointsYellow, PointsReset)
		return
	}

	fmt.Printf("%s%-15s %-10s %-8s %-3s %-6s %s%s\n",
		PointsBold+PointsWhite, "RACE", "DATE", "TYPE", "POS", "POINTS", "NOTES", PointsReset)
	fmt.Printf("%s%s%s\n", PointsBold, strings.Repeat("─", 80), PointsReset)

	for i, breakdown := range pointsBreakdown {
		sessionType := "Race"
		sessionColor := PointsGreen
		if breakdown.SessionType == "Sprint" {
			sessionType = "Sprint"
			sessionColor = PointsYellow
		}

		notes := ""
		noteColor := PointsReset
		if breakdown.IsAdjusted {
			notes = "Promoted due to DSQ"
			noteColor = PointsCyan
		}

		// Position colors
		posColor := PointsReset
		if breakdown.Position == 1 {
			posColor = PointsBold + PointsYellow // Gold for win
		} else if breakdown.Position <= 3 {
			posColor = PointsBold + PointsWhite // Podium
		} else if breakdown.Points > 0 {
			posColor = PointsGreen // Points
		}

		// Points colors
		pointColor := PointsReset
		if breakdown.Points >= 15 {
			pointColor = PointsBold + PointsYellow
		} else if breakdown.Points >= 8 {
			pointColor = PointsBold + PointsGreen
		} else if breakdown.Points > 0 {
			pointColor = PointsGreen
		}

		fmt.Printf("%-15s %s%-10s%s %s%-8s%s %sP%-2d%s %s%-6d%s %s%s%s\n",
			truncateStringPoints(breakdown.RaceName, 15),
			PointsCyan, breakdown.Date.Format("2006-01-02"), PointsReset,
			sessionColor, sessionType, PointsReset,
			posColor, breakdown.Position, PointsReset,
			pointColor, breakdown.Points, PointsReset,
			noteColor, notes, PointsReset)

		// Add separator after every 5 races for readability
		if (i+1)%5 == 0 && i < len(pointsBreakdown)-1 {
			fmt.Printf("%s%s%s\n", PointsMagenta, strings.Repeat("┄", 80), PointsReset)
		}
	}

	fmt.Printf("%s%s%s\n", PointsBold, strings.Repeat("─", 80), PointsReset)
	fmt.Printf("%sPoints scored in %d/%d sessions%s\n",
		PointsBold+PointsCyan, len(pointsBreakdown), countCompletedSessions(sessions), PointsReset)

	// Show points system info with colors
	fmt.Printf("\n%sPoints Systems:%s\n", PointsBold+PointsBlue, PointsReset)
	fmt.Printf("   %sRace:%s   25-18-15-12-10-8-6-4-2-1 (positions 1-10)\n", PointsGreen, PointsReset)
	fmt.Printf("   %sSprint:%s 8-7-6-5-4-3-2-1 (positions 1-8)\n", PointsYellow, PointsReset)
}

// getPointsTeamColor returns ANSI color codes for different F1 teams
func getPointsTeamColor(team string) string {
	switch team {
	case "McLaren":
		return "\033[38;5;208m"
	case "Red Bull Racing":
		return "\033[38;5;27m"
	case "Ferrari":
		return PointsRed
	case "Mercedes":
		return "\033[38;5;51m"
	case "Aston Martin":
		return PointsGreen
	case "Alpine":
		return "\033[38;5;129m"
	case "Williams":
		return PointsBlue
	case "Haas F1 Team", "Haas":
		return "\033[38;5;245m"
	case "Kick Sauber":
		return "\033[38;5;46m"
	case "Racing Bulls":
		return "\033[38;5;63m"
	default:
		return PointsReset
	}
}

// truncateStringPoints truncates strings for points display
func truncateStringPoints(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// countCompletedSessions counts how many sessions have been completed
func countCompletedSessions(sessions []data.OpenF1Session) int {
	count := 0
	for _, session := range sessions {
		if session.DateStart.Before(time.Now()) {
			count++
		}
	}
	return count
}

// ShowPointsHelp displays help information for the points command
func ShowPointsHelp() {
	fmt.Printf("%sPoints Breakdown Command%s\n", PointsBold+PointsYellow, PointsReset)
	fmt.Printf("%s%s%s\n", PointsBold, strings.Repeat("═", 50), PointsReset)
	fmt.Println()
	fmt.Printf("%sUsage:%s\n", PointsBold+PointsGreen, PointsReset)
	fmt.Printf("  %sf1 points \"<driver_name>\"%s\n", PointsCyan, PointsReset)
	fmt.Println()
	fmt.Printf("%sDescription:%s\n", PointsBold+PointsGreen, PointsReset)
	fmt.Printf("  Shows a detailed breakdown of points scored by a specific driver\n")
	fmt.Printf("  in each race and sprint session of the current season.\n")
	fmt.Println()
	fmt.Printf("%sFeatures:%s\n", PointsBold+PointsGreen, PointsReset)
	fmt.Printf("  %s•%s Race-by-race points breakdown\n", PointsBlue, PointsReset)
	fmt.Printf("  %s•%s Sprint session points\n", PointsYellow, PointsReset)
	fmt.Printf("  %s•%s Position adjustments due to disqualifications\n", PointsRed, PointsReset)
	fmt.Printf("  %s•%s Total points and wins summary\n", PointsGreen, PointsReset)
	fmt.Printf("  %s•%s Points system information\n", PointsMagenta, PointsReset)
	fmt.Println()
	fmt.Printf("%sExamples:%s\n", PointsBold+PointsGreen, PointsReset)
	fmt.Printf("  %sf1 points \"Oscar Piastri\"%s     # Show Oscar's points breakdown\n", PointsCyan, PointsReset)
	fmt.Printf("  %sf1 points \"Lewis Hamilton\"%s    # Show Lewis's points breakdown\n", PointsCyan, PointsReset)
	fmt.Printf("  %sf1 points \"Max Verstappen\"%s    # Show Max's points breakdown\n", PointsCyan, PointsReset)
	fmt.Println()
	fmt.Printf("%sNote:%s\n", PointsBold+PointsMagenta, PointsReset)
	fmt.Printf("  Driver names are case-insensitive and should match the full name\n")
	fmt.Printf("  as shown in the drivers list (%sf1 drivers%s).\n", PointsCyan, PointsReset)
}
