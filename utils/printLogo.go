package utils

import "fmt"

func PrintAsciiLogoAndInfo() {
	PrintAsciiShadowLogo()
	PrintVersion()
	PrintUrl()
	fmt.Println("")
}

func PrintTitle(newline bool, title string) {
	if newline {
		fmt.Println("")
	}
	fmt.Println(title)
	fmt.Println("==================")
}

func PrintVersion() {
	fmt.Println("   " + GetVersion())
}

func PrintUrl() {
	fmt.Println("   " + GetUrl())
}

func GetUrl() string {
	return ""
}

func GetVersion() string {
	base := "0.0.1"
	version := base + "-" + GetTimeSuffixString()
	return version
}

func PrintAsciiShadowLogo() {
	fmt.Println("")
	fmt.Println("   ████████╗██████╗  █████╗ ██████╗ ███████╗")
	fmt.Println("   ╚══██╔══╝██╔══██╗██╔══██╗██╔══██╗██╔════╝")
	fmt.Println("      ██║   ██████╔╝███████║██║  ██║█████╗  ")
	fmt.Println("      ██║   ██╔══██╗██╔══██║██║  ██║██╔══╝  ")
	fmt.Println("      ██║   ██║  ██║██║  ██║██████╔╝███████╗")
	fmt.Println("      ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝ ╚══════╝")
}
