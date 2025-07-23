package main

import (
	"fmt"
	"os"

	"f1cli/commands"
	"f1cli/data"
)

func main() {
	dataService := data.NewDataService()

	if len(os.Args) < 2 {
		showWelcomeAndHelp()
		return
	}

	userCommand := os.Args[1]

	if userCommand == "--help" || userCommand == "-h" {
		showWelcomeAndHelp()
		return
	}

	switch userCommand {
	case "drivers":
		commands.DriversWithService(os.Args[2:], dataService)
	case "standings":
		commands.Standings(os.Args[2:], dataService)
	case "results":
		commands.Results(os.Args[2:], dataService)
	case "points":
		commands.Points(os.Args[2:], dataService)
	case "status":
		commands.Status(os.Args[2:], dataService)
	case "help":
		if len(os.Args) > 2 {
			showSpecificCommandHelp(os.Args[2])
		} else {
			showWelcomeAndHelp()
		}
	default:
		fmt.Printf("❌ Sorry, I don't know how to handle the command: %s\n", userCommand)
		fmt.Println("Try 'f1 help' to see what commands are available")
		os.Exit(1)
	}
}

func showWelcomeAndHelp() {
	fmt.Println("F1 CLI - Your Formula 1 Command Line Companion!")
	fmt.Println()
	fmt.Println("This tool helps you explore the exciting world of Formula 1 with real-time data.")
	fmt.Println("You can check driver standings, race results, and much more!")
	fmt.Println()
	fmt.Println("How to use this tool:")
	fmt.Println("  f1 <command> [options] [arguments]")
	fmt.Println()
	fmt.Println("Available Commands:")
	fmt.Println("  drivers      Discover information about F1 drivers")
	fmt.Println("  standings    View championship standings (drivers or teams)")
	fmt.Println("  results      See race and sprint results from specific locations")
	fmt.Println("  points       Deep dive into a driver's points breakdown")
	fmt.Println("  status       Check if our data source is working properly")
	fmt.Println("  help         Get help (you're looking at it now!)")
	fmt.Println()
	fmt.Println("Quick Examples to Get Started:")
	fmt.Println("  f1 drivers                     → See all current F1 drivers")
	fmt.Println("  f1 standings                   → View the driver championship")
	fmt.Println("  f1 standings -c                → View the constructor championship")
	fmt.Println("  f1 results Shanghai            → See Shanghai Grand Prix results")
	fmt.Println("  f1 points \"Oscar Piastri\"      → See how Oscar earned his points")
	fmt.Println("  f1 drivers -d                  → Get detailed driver information")
	fmt.Println("  f1 drivers \"Lewis Hamilton\"    → Focus on a specific driver")
	fmt.Println("  f1 status                      → Make sure everything is working")
	fmt.Println("  f1 help drivers                → Learn more about the drivers command")
	fmt.Println()
	fmt.Println("Pro Tip: Use 'f1 help <command>' to learn more about any specific command!")
	fmt.Println("Data is provided by the OpenF1 API for the most up-to-date information.")
}

func showSpecificCommandHelp(commandName string) {
	switch commandName {
	case "drivers":
		fmt.Println("Getting help for the 'drivers' command...")
		fmt.Println()
		commands.ShowDriversHelp()
	case "standings":
		fmt.Println("Getting help for the 'standings' command...")
		fmt.Println()
		commands.ShowStandingsHelp()
	case "results":
		fmt.Println("Getting help for the 'results' command...")
		fmt.Println()
		commands.ShowResultsHelp()
	case "points":
		fmt.Println("Getting help for the 'points' command...")
		fmt.Println()
		commands.ShowPointsHelp()
	default:
		fmt.Printf("❌ Sorry, I don't have specific help for the command: %s\n", commandName)
		fmt.Println("Try one of these commands for help: drivers, standings, results, points")
		fmt.Println("Or use 'f1 help' to see all available commands")
	}
}
