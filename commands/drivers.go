package commands

import (
	"flag"
	"fmt"
	"strings"

	"f1cli/data"
)

// DriversWithService shows driver information using the provided data service
func DriversWithService(args []string, dataService *data.DataService) {
	fs := flag.NewFlagSet("drivers", flag.ExitOnError)

	detailed := fs.Bool("detailed", false, "Show detailed driver information")
	detailedShort := fs.Bool("d", false, "Show detailed driver information")
	team := fs.String("team", "", "Filter drivers by team")
	teamShort := fs.String("t", "", "Filter drivers by team")
	helpFlag := fs.Bool("help", false, "Show help for drivers command")

	fs.Parse(args)

	if *helpFlag {
		ShowDriversHelp()
		return
	}

	remaining := fs.Args()

	// If specific driver requested
	if len(remaining) > 0 {
		driverName := strings.Join(remaining, " ")
		driver, err := dataService.GetDriverByName(driverName)
		if err != nil {
			fmt.Printf("❌ Driver '%s' not found: %v\n", driverName, err)
			return
		}
		showDriverDetail(driver)
		return
	}

	showDetailed := *detailed || *detailedShort
	teamFilter := *team
	if teamFilter == "" {
		teamFilter = *teamShort
	}

	drivers, err := dataService.GetDrivers()
	if err != nil {
		fmt.Printf("❌ Error fetching drivers: %v\n", err)
		return
	}

	fmt.Printf("F1 2025 Drivers (%s)\n", dataService.GetSourceName())
	fmt.Println("══════════════════════════════════════════════")

	for _, driver := range drivers {
		// Apply team filter if specified
		if teamFilter != "" && !strings.EqualFold(driver.Team, teamFilter) {
			continue
		}

		if showDetailed {
			fmt.Printf("\n%d. %s (#%d) - %s\n", driver.ID, driver.Name, driver.Number, driver.Team)
			fmt.Printf("   Country: %s\n", driver.Country)
			fmt.Printf("   Points: %d | Wins: %d | Podiums: %d\n", driver.Points, driver.Wins, driver.Podiums)
			if driver.Championships > 0 {
				fmt.Printf("   Championships: %d\n", driver.Championships)
			}
		} else {
			fmt.Printf("%2d. %-20s #%-2d %-20s %3d pts\n",
				driver.ID, driver.Name, driver.Number, driver.Team, driver.Points)
		}
	}

	if teamFilter != "" {
		fmt.Printf("\n📋 Filtered by team: %s\n", teamFilter)
	}

	if dataService.GetSourceName() == "Ergast F1 API" {
		if !dataService.IsOnline() {
			fmt.Println("\n⚠️  API appears to be offline or unreachable")
		}
	}
}

func showDriverDetail(driver *data.Driver) {
	fmt.Printf("\n%s (#%d)\n", driver.Name, driver.Number)
	fmt.Println("══════════════════════════════════════════════")
	fmt.Printf("Country: %s\n", driver.Country)
	fmt.Printf("Team: %s\n", driver.Team)
	fmt.Printf("Championship Points: %d\n", driver.Points)
	fmt.Printf("Race Wins: %d\n", driver.Wins)
	fmt.Printf("Podium Finishes: %d\n", driver.Podiums)
	if driver.Championships > 0 {
		fmt.Printf("World Championships: %d\n", driver.Championships)
	}
}

func ShowDriversHelp() {
	fmt.Println("Display F1 driver information")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  f1 drivers [flags] [driver_name]")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  driver_name    Show detailed info for specific driver")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -d, -detailed      Show detailed information for all drivers")
	fmt.Println("  -t, -team <name>   Filter drivers by team")
	fmt.Println("  -help              Show help for drivers command")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  f1 drivers                    # List all drivers")
	fmt.Println("  f1 drivers -d                 # Detailed list")
	fmt.Println("  f1 drivers -t McLaren         # McLaren drivers only")
	fmt.Println("  f1 drivers \"Max Verstappen\"   # Specific driver info")
}
