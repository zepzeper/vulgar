package cli

import (
	"fmt"
	"os"
	"strings"
)

func PrintSuccess(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(StyleSuccess.Render(SymbolSuccess + " " + msg))
}

func PrintError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, StyleError.Render(SymbolError+" "+msg))
}

func PrintWarning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(StyleWarning.Render(SymbolWarning + " " + msg))
}

func PrintInfo(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(StyleInfo.Render(SymbolInfo + " " + msg))
}

func PrintHeader(title string) {
	fmt.Println()
	fmt.Println(StyleTitle.Render(title))
	fmt.Println(StyleMuted.Render(strings.Repeat("-", len(title)+2)))
}

func PrintSubheader(title string) {
	fmt.Println()
	fmt.Println(StylePrimary.Render(title))
}

func PrintTip(tip string) {
	fmt.Println()
	fmt.Println(StyleSubtitle.Render("Tip: " + tip))
}

func PrintBox(title, content string) {
	if title != "" {
		fmt.Println(StyleTitle.Render("  " + title))
	}
	fmt.Println(StyleBox.Render(content))
}

func PrintBoxSuccess(title, content string) {
	if title != "" {
		fmt.Println(StyleSuccess.Render("  " + title))
	}
	fmt.Println(StyleBoxSuccess.Render(content))
}

func PrintBoxError(title, content string) {
	if title != "" {
		fmt.Println(StyleError.Render("  " + title))
	}
	fmt.Println(StyleBoxError.Render(content))
}

func PrintBoxWarning(title, content string) {
	if title != "" {
		fmt.Println(StyleWarning.Render("  " + title))
	}
	fmt.Println(StyleBoxWarning.Render(content))
}

func PrintKeyValue(key, value string) {
	fmt.Printf("  %s: %s\n", StyleMuted.Render(key), value)
}

func PrintList(items []string) {
	for _, item := range items {
		fmt.Printf("  %s %s\n", StyleMuted.Render(SymbolBullet), item)
	}
}

func PrintNumberedList(items []string) {
	for i, item := range items {
		fmt.Printf("  %s %s\n", StyleMuted.Render(fmt.Sprintf("%d.", i+1)), item)
	}
}

func PrintDivider() {
	fmt.Println(StyleMuted.Render(strings.Repeat("-", 40)))
}

func PrintNewline() {
	fmt.Println()
}

func Print(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func Println(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func PrintCode(code string) {
	fmt.Println(StyleCode.Render(code))
}

func PrintUsageHint(moduleName, funcName, idVar, idValue string) {
	fmt.Println()
	fmt.Println(StyleMuted.Render("Use in your workflow:"))
	code := fmt.Sprintf(`local %s = require("%s")
local result = %s.%s(client, "%s")`,
		getModuleShortName(moduleName),
		moduleName,
		getModuleShortName(moduleName),
		funcName,
		idValue,
	)
	fmt.Println(StyleCode.Render(code))
}

func getModuleShortName(moduleName string) string {
	parts := strings.Split(moduleName, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return moduleName
}
