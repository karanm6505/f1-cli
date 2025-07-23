package commands

import (
	"flag"
	"fmt"
	"strings"

	"f1cli/data"
)

// ANSI color codes
const (
	Reset    = "\033[0m"
	Bold     = "\033[1m"
	Red      = "\033[31m"
	Green    = "\033[32m"
	Yellow   = "\033[33m"
	Blue     = "\033[34m"
	Magenta  = "\033[35m"
	Cyan     = "\033[36m"
	White    = "\033[37m"
	BgRed    = "\033[41m"
	BgGreen  = "\033[42m"
	BgYellow = "\033[43m"
)

// Standings shows driver or constructor championship standings
func Standings(args []string, dataService *data.DataService) {
	fs := flag.NewFlagSet("standings", flag.ExitOnError)

	constructor := fs.Bool("constructor", false, "Show constructor standings")
	constructorShort := fs.Bool("c", false, "Show constructor standings")
	verbose := fs.Bool("verbose", false, "Show detailed points breakdown")
	verboseShort := fs.Bool("v", false, "Show detailed points breakdown")
	helpFlag := fs.Bool("help", false, "Show help for standings command")

	fs.Parse(args)

	if *helpFlag {
		ShowStandingsHelp()
		return
	}

	showConstructor := *constructor || *constructorShort
	showVerbose := *verbose || *verboseShort

	if showConstructor {
		showConstructorStandings(dataService)
	} else {
		showDriverStandings(dataService, showVerbose)
	}
}

func showDriverStandings(dataService *data.DataService, verbose bool) {
	fmt.Printf("F1 2025 Driver Championship (%s)\n",
		dataService.GetSourceName())
	fmt.Printf("%s%s%s\n", Bold, strings.Repeat("═", 80), Reset)

	standings, err := dataService.GetDriverStandings()
	if err != nil {
		fmt.Printf("%s❌ Error fetching driver standings: %v%s\n", Red, err, Reset)
		return
	}

	if len(standings) == 0 {
		fmt.Printf("%s⚠️  No standings data available%s\n", Yellow, Reset)
		return
	}

	// Header with colors
	fmt.Printf("%s%-3s %-25s %-20s %6s %4s %s%s\n",
		Bold+White, "POS", "DRIVER", "TEAM", "POINTS", "WINS", "GAP", Reset)
	fmt.Printf("%s%s%s\n", Bold, strings.Repeat("─", 80), Reset)

	for i, standing := range standings {
		// Color coding for positions
		var posColor string
		switch {
		case standing.Position == 1:
			posColor = Bold + Yellow
		case standing.Position <= 3:
			posColor = Bold + White
		case standing.Position <= 10:
			posColor = Green
		default:
			posColor = Reset
		}

		teamColor := getTeamColor(standing.Team)

		// Points emphasis
		pointsColor := ""
		if standing.Points > 200 {
			pointsColor = Bold + Yellow
		} else if standing.Points > 100 {
			pointsColor = Bold + Green
		} else if standing.Points > 50 {
			pointsColor = Green
		}

		fmt.Printf("%s%-3d%s %-25s %s%-20s%s %s%6d%s %4d %s\n",
			posColor, standing.Position, Reset,
			standing.Driver,
			teamColor, standing.Team, Reset,
			pointsColor, standing.Points, Reset,
			standing.Wins,
			standing.Gap)

		// Add separator lines for visual grouping
		if i == 2 { // After podium
			fmt.Printf("%s%s%s\n", Cyan, strings.Repeat("┄", 80), Reset)
		} else if i == 9 { // After points positions
			fmt.Printf("%s%s%s\n", Magenta, strings.Repeat("┄", 80), Reset)
		}
	}

	fmt.Printf("\n%sTotal drivers: %d%s", Bold+Cyan, len(standings), Reset)

	// Championship leader info
	if len(standings) > 1 {
		leader := standings[0]
		fmt.Printf("\nChampionship Leader: %s%s %s(%d points, %d wins)%s",
			leader.Driver, Reset, Green, leader.Points, leader.Wins, Reset)
	}
	fmt.Println()

	if verbose {
		fmt.Printf("\nPoints System Information:%s\n", Reset)
		fmt.Printf("   Race Points: 25-18-15-12-10-8-6-4-2-1 (positions 1-10)\n")
		fmt.Printf("   Sprint Points: 8-7-6-5-4-3-2-1 (positions 1-8)\n")
		fmt.Printf("   Wins count: Only main races (not sprints)\n")
	}
}

// getTeamColor returns ANSI color codes for different F1 teams
func getTeamColor(team string) string {
	switch team {
	case "McLaren":
		return "\033[38;5;208m"
	case "Red Bull Racing":
		return "\033[38;5;27m"
	case "Ferrari":
		return Red
	case "Mercedes":
		return "\033[38;5;51m"
	case "Aston Martin":
		return Green
	case "Alpine":
		return "\033[38;5;129m"
	case "Williams":
		return Blue
	case "Haas F1 Team", "Haas":
		return "\033[38;5;245m"
	case "Kick Sauber":
		return "\033[38;5;46m"
	case "Racing Bulls":
		return "\033[38;5;63m"
	default:
		return Reset
	}
}

func showConstructorStandings(dataService *data.DataService) {
	fmt.Printf("F1 2025 Constructor Championship (%s)\n",
		dataService.GetSourceName())
	fmt.Printf("%s%s%s\n", Bold, strings.Repeat("═", 80), Reset)

	standings, err := dataService.GetConstructorStandings()
	if err != nil {
		fmt.Printf("%s❌ Error fetching constructor standings: %v%s\n", Red, err, Reset)
		return
	}

	if len(standings) == 0 {
		fmt.Printf("%s⚠️  No standings data available%s\n", Yellow, Reset)
		return
	}

	fmt.Printf("%s%-3s %-25s %-15s %6s %4s %s%s\n",
		Bold+White, "POS", "CONSTRUCTOR", "COUNTRY", "POINTS", "WINS", "GAP", Reset)
	fmt.Printf("%s%s%s\n", Bold, strings.Repeat("─", 80), Reset)

	for i, standing := range standings {
		// Color coding for positions
		var posColor string
		switch {
		case standing.Position == 1:
			posColor = Bold + Yellow // Gold for 1st
		case standing.Position <= 3:
			posColor = Bold + White // Silver/Bronze for podium
		default:
			posColor = Reset
		}

		// Team colors
		teamColor := getTeamColor(standing.Driver) // Constructor name is in Driver field

		// Points emphasis
		pointsColor := ""
		if standing.Points > 400 {
			pointsColor = Bold + Yellow
		} else if standing.Points > 200 {
			pointsColor = Bold + Green
		} else if standing.Points > 100 {
			pointsColor = Green
		}

		fmt.Printf("%s%-3d%s %s%-25s%s %-15s %s%6d%s %4d %s\n",
			posColor, standing.Position, Reset,
			teamColor, standing.Driver, Reset, // Constructor name
			standing.Team, // Country
			pointsColor, standing.Points, Reset,
			standing.Wins,
			standing.Gap)

		// Add separator after podium
		if i == 2 {
			fmt.Printf("%s%s%s\n", Cyan, strings.Repeat("┄", 80), Reset)
		}
	}

	fmt.Printf("\n%sTotal constructors: %d%s", Bold+Cyan, len(standings), Reset)

	// Championship leader info
	if len(standings) > 1 {
		leader := standings[0]
		fmt.Printf("\nConstructor Champion: %s%s %s(%d points, %d wins)%s",
			leader.Driver, Reset, Green, leader.Points, leader.Wins, Reset)
	}
	fmt.Println()
}

func ShowStandingsHelp() {
	fmt.Printf("%sF1 Championship Standings%s\n", Bold+Yellow, Reset)
	fmt.Printf("%s%s%s\n", Bold, strings.Repeat("═", 50), Reset)
	fmt.Println()
	fmt.Printf("%sUsage:%s\n", Bold+Green, Reset)
	fmt.Printf("  %sf1 standings [flags]%s\n", Cyan, Reset)
	fmt.Println()
	fmt.Printf("%sFlags:%s\n", Bold+Green, Reset)
	fmt.Printf("  %s-c, -constructor%s   Show constructor standings instead of driver standings\n", Yellow, Reset)
	fmt.Printf("  %s-v, -verbose%s       Show detailed points system information\n", Yellow, Reset)
	fmt.Printf("  %s-help%s              Show help for standings command\n", Yellow, Reset)
	fmt.Println()
	fmt.Printf("%sExamples:%s\n", Bold+Green, Reset)
	fmt.Printf("  %sf1 standings%s                  # Show driver championship standings\n", Cyan, Reset)
	fmt.Printf("  %sf1 standings -c%s               # Show constructor championship standings\n", Cyan, Reset)
	fmt.Printf("  %sf1 standings -v%s               # Show driver standings with points system info\n", Cyan, Reset)
	fmt.Printf("  %sf1 standings -constructor%s     # Show constructor championship standings\n", Cyan, Reset)
	fmt.Println()
	fmt.Printf("%sPoints Systems:%s\n", Bold+Blue, Reset)
	fmt.Printf("  Race: 25-18-15-12-10-8-6-4-2-1 points (positions 1-10)\n")
	fmt.Printf("  Sprint: 8-7-6-5-4-3-2-1 points (positions 1-8)\n")
	fmt.Println()
	fmt.Printf("%sNote:%s Standings are calculated from real race results using %sOpenF1 API%s\n",
		Bold+Magenta, Reset, Bold+Cyan, Reset)
}
