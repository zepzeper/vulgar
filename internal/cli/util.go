package cli

import (
	"strings"
	"time"
)

func FormatTimeAgo(timeStr string) string {
	if timeStr == "" {
		return ""
	}

	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02 15:04:05",
	}

	var t time.Time
	var err error
	for _, format := range formats {
		t, err = time.Parse(format, timeStr)
		if err == nil {
			break
		}
	}

	if err != nil {
		return timeStr // Return original if parsing fails
	}

	return TimeAgo(t)
}

func TimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1m ago"
		}
		return Sprintf("%dm ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1h ago"
		}
		return Sprintf("%dh ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1d ago"
		}
		return Sprintf("%dd ago", days)
	case diff < 30*24*time.Hour:
		weeks := int(diff.Hours() / 24 / 7)
		if weeks == 1 {
			return "1w ago"
		}
		return Sprintf("%dw ago", weeks)
	case diff < 365*24*time.Hour:
		months := int(diff.Hours() / 24 / 30)
		if months == 1 {
			return "1mo ago"
		}
		return Sprintf("%dmo ago", months)
	default:
		years := int(diff.Hours() / 24 / 365)
		if years == 1 {
			return "1y ago"
		}
		return Sprintf("%dy ago", years)
	}
}

func Sprintf(format string, args ...interface{}) string {
	result := format
	for _, arg := range args {
		switch v := arg.(type) {
		case int:
			result = strings.Replace(result, "%d", intToString(v), 1)
		case string:
			result = strings.Replace(result, "%s", v, 1)
		}
	}
	return result
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	if negative {
		return "-" + string(digits)
	}
	return string(digits)
}

func GetString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func GetBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

func GetFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0
}

func GetInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	if v, ok := m[key].(int); ok {
		return v
	}
	return 0
}

func GetMap(m map[string]interface{}, key string) map[string]interface{} {
	if v, ok := m[key].(map[string]interface{}); ok {
		return v
	}
	return nil
}

func GetSlice(m map[string]interface{}, key string) []interface{} {
	if v, ok := m[key].([]interface{}); ok {
		return v
	}
	return nil
}

func FileTypeIcon(mimeType string) string {
	switch {
	case strings.Contains(mimeType, "folder"):
		return "[DIR]"
	case strings.Contains(mimeType, "spreadsheet"), strings.Contains(mimeType, "excel"):
		return "[XLS]"
	case strings.Contains(mimeType, "document"), strings.Contains(mimeType, "word"):
		return "[DOC]"
	case strings.Contains(mimeType, "presentation"), strings.Contains(mimeType, "powerpoint"):
		return "[PPT]"
	case strings.Contains(mimeType, "pdf"):
		return "[PDF]"
	case strings.Contains(mimeType, "image"):
		return "[IMG]"
	case strings.Contains(mimeType, "video"):
		return "[VID]"
	case strings.Contains(mimeType, "audio"):
		return "[AUD]"
	case strings.Contains(mimeType, "zip"), strings.Contains(mimeType, "archive"):
		return "[ZIP]"
	case strings.Contains(mimeType, "text"):
		return "[TXT]"
	default:
		return "[---]"
	}
}

func StatusIcon(status string) string {
	switch strings.ToLower(status) {
	case "success", "passed", "complete", "completed", "done", "merged":
		return SymbolSuccess
	case "failed", "failure", "error":
		return SymbolError
	case "warning", "warn":
		return SymbolWarning
	case "running", "pending", "in_progress", "in progress":
		return SymbolLoading
	case "open", "opened":
		return SymbolArrow
	case "closed":
		return SymbolDash
	default:
		return SymbolBullet
	}
}

func FormatStatus(status string) string {
	icon := StatusIcon(status)
	return icon + " " + status
}
