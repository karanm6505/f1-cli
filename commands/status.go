package commands

import (
	"fmt"

	"f1cli/data"
)

// Status shows the status of data sources
func Status(args []string, dataService *data.DataService) {
	fmt.Println("F1 CLI Data Source Status")
	fmt.Println("══════════════════════════════════════════════")

	fmt.Printf("Current Source: %s\n", dataService.GetSourceName())

	fmt.Print("API Connectivity: ")
	if dataService.IsOnline() {
		fmt.Println("✅ Online")
	} else {
		fmt.Println("❌ Offline or unreachable")
		fmt.Println("\nThe OpenF1 API might be:")
		fmt.Println("   • Temporarily down")
		fmt.Println("   • Blocked by firewall")
		fmt.Println("   • Rate limited")
		fmt.Println("\nPlease check your internet connection and try again")
	}

	fmt.Println("\nData Source Information:")
	fmt.Println("   • OpenF1 API (https://api.openf1.org/v1)")
	fmt.Println("   • Real-time F1 data and telemetry")
	fmt.Println("   • Session information and results")
	fmt.Println("   • Driver and team information")
}
