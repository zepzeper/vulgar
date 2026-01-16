package validator

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
)

const ModuleName = "stdlib.validator"

// emailRegex is a simple email validation regex
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// phoneRegex is a simple phone validation regex
var phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

// slugRegex validates URL slug format
var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// luhnCheck validates credit card using Luhn algorithm
func luhnCheck(cardNumber string) bool {
	cardNumber = strings.ReplaceAll(cardNumber, " ", "")
	cardNumber = strings.ReplaceAll(cardNumber, "-", "")

	sum := 0
	alternate := false

	for i := len(cardNumber) - 1; i >= 0; i-- {
		digit := int(cardNumber[i] - '0')
		if digit < 0 || digit > 9 {
			return false
		}

		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0 && len(cardNumber) >= 13
}

// luaIsEmail validates email format
// Usage: local valid = validator.is_email("user@example.com")
func luaIsEmail(L *lua.LState) int {
	email := L.CheckString(1)
	valid := emailRegex.MatchString(email)
	L.Push(lua.LBool(valid))
	return 1
}

// luaIsURL validates URL format
// Usage: local valid = validator.is_url("https://example.com")
func luaIsURL(L *lua.LState) int {
	urlStr := L.CheckString(1)
	_, err := url.ParseRequestURI(urlStr)
	L.Push(lua.LBool(err == nil && urlStr != ""))
	return 1
}

// luaIsIP validates IP address
// Usage: local valid = validator.is_ip("192.168.1.1")
func luaIsIP(L *lua.LState) int {
	ipStr := L.CheckString(1)
	ip := net.ParseIP(ipStr)
	L.Push(lua.LBool(ip != nil))
	return 1
}

// luaIsIPv4 validates IPv4 address
// Usage: local valid = validator.is_ipv4("192.168.1.1")
func luaIsIPv4(L *lua.LState) int {
	ipStr := L.CheckString(1)
	ip := net.ParseIP(ipStr)
	L.Push(lua.LBool(ip != nil && ip.To4() != nil))
	return 1
}

// luaIsIPv6 validates IPv6 address
// Usage: local valid = validator.is_ipv6("::1")
func luaIsIPv6(L *lua.LState) int {
	ipStr := L.CheckString(1)
	ip := net.ParseIP(ipStr)
	L.Push(lua.LBool(ip != nil && ip.To4() == nil))
	return 1
}

// luaIsUUID validates UUID format
// Usage: local valid = validator.is_uuid("550e8400-e29b-41d4-a716-446655440000")
func luaIsUUID(L *lua.LState) int {
	uuidStr := L.CheckString(1)
	_, err := uuid.Parse(uuidStr)
	L.Push(lua.LBool(err == nil))
	return 1
}

// luaIsJSON validates JSON string
// Usage: local valid = validator.is_json('{"key": "value"}')
func luaIsJSON(L *lua.LState) int {
	jsonStr := L.CheckString(1)
	var js interface{}
	err := json.Unmarshal([]byte(jsonStr), &js)
	L.Push(lua.LBool(err == nil))
	return 1
}

// luaIsNumeric checks if string is numeric (integers only, no decimals)
// Usage: local valid = validator.is_numeric("12345")
func luaIsNumeric(L *lua.LState) int {
	str := L.CheckString(1)
	// Reject if contains decimal point
	if strings.Contains(str, ".") {
		L.Push(lua.LBool(false))
		return 1
	}
	_, err := strconv.ParseInt(str, 10, 64)
	L.Push(lua.LBool(err == nil))
	return 1
}

// luaIsAlpha checks if string is alphabetic
// Usage: local valid = validator.is_alpha("hello")
func luaIsAlpha(L *lua.LState) int {
	str := L.CheckString(1)
	if str == "" {
		L.Push(lua.LBool(false))
		return 1
	}
	for _, r := range str {
		if !unicode.IsLetter(r) {
			L.Push(lua.LBool(false))
			return 1
		}
	}
	L.Push(lua.LBool(true))
	return 1
}

// luaIsAlphanumeric checks if string is alphanumeric
// Usage: local valid = validator.is_alphanumeric("hello123")
func luaIsAlphanumeric(L *lua.LState) int {
	str := L.CheckString(1)
	if str == "" {
		L.Push(lua.LBool(false))
		return 1
	}
	for _, r := range str {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			L.Push(lua.LBool(false))
			return 1
		}
	}
	L.Push(lua.LBool(true))
	return 1
}

// luaIsCreditCard validates credit card number (Luhn algorithm)
// Usage: local valid = validator.is_credit_card("4111111111111111")
func luaIsCreditCard(L *lua.LState) int {
	cardNumber := L.CheckString(1)
	valid := luhnCheck(cardNumber)
	L.Push(lua.LBool(valid))
	return 1
}

// luaIsPhone validates phone number format
// Usage: local valid = validator.is_phone("+1-555-555-5555")
func luaIsPhone(L *lua.LState) int {
	phone := L.CheckString(1)
	// Remove common separators for validation
	cleaned := strings.ReplaceAll(phone, "-", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")
	valid := phoneRegex.MatchString(cleaned) && len(cleaned) >= 10
	L.Push(lua.LBool(valid))
	return 1
}

// luaIsDate validates date format
// Usage: local valid = validator.is_date("2024-01-15", "YYYY-MM-DD")
func luaIsDate(L *lua.LState) int {
	dateStr := L.CheckString(1)
	format := L.CheckString(2)

	// Convert format to Go time format
	goFormat := strings.ReplaceAll(format, "YYYY", "2006")
	goFormat = strings.ReplaceAll(goFormat, "MM", "01")
	goFormat = strings.ReplaceAll(goFormat, "DD", "02")
	goFormat = strings.ReplaceAll(goFormat, "HH", "15")
	goFormat = strings.ReplaceAll(goFormat, "mm", "04")
	goFormat = strings.ReplaceAll(goFormat, "ss", "05")

	_, err := time.Parse(goFormat, dateStr)
	L.Push(lua.LBool(err == nil))
	return 1
}

// luaIsBase64 validates base64 string
// Usage: local valid = validator.is_base64("aGVsbG8=")
func luaIsBase64(L *lua.LState) int {
	str := L.CheckString(1)
	_, err := base64.StdEncoding.DecodeString(str)
	L.Push(lua.LBool(err == nil))
	return 1
}

// luaIsHex validates hexadecimal string
// Usage: local valid = validator.is_hex("deadbeef")
func luaIsHex(L *lua.LState) int {
	str := L.CheckString(1)
	if str == "" {
		L.Push(lua.LBool(false))
		return 1
	}
	_, err := hex.DecodeString(str)
	L.Push(lua.LBool(err == nil))
	return 1
}

// luaIsSlug validates URL slug
// Usage: local valid = validator.is_slug("hello-world")
func luaIsSlug(L *lua.LState) int {
	slug := L.CheckString(1)
	valid := slugRegex.MatchString(strings.ToLower(slug))
	L.Push(lua.LBool(valid))
	return 1
}

// luaMatches checks if string matches regex pattern
// Usage: local valid = validator.matches("hello123", "^[a-z]+[0-9]+$")
func luaMatches(L *lua.LState) int {
	str := L.CheckString(1)
	pattern := L.CheckString(2)

	matched, err := regexp.MatchString(pattern, str)
	if err != nil {
		L.Push(lua.LBool(false))
		return 1
	}
	L.Push(lua.LBool(matched))
	return 1
}

// luaValidateSchema validates data against JSON schema
// Usage: local valid, errors = validator.schema(data, schema)
func luaValidateSchema(L *lua.LState) int {
	_ = L.CheckTable(1) // dataTable
	_ = L.CheckTable(2) // schemaTable

	// For now, return true with no errors
	// Full implementation would require a JSON Schema library
	L.Push(lua.LBool(true))
	L.Push(lua.LNil)
	return 2
}

// luaLength checks string length constraints
// Usage: local valid = validator.length("hello", {min = 1, max = 10})
func luaLength(L *lua.LState) int {
	str := L.CheckString(1)
	opts := L.CheckTable(2)

	length := len(str)
	min := -1
	max := -1

	if minVal := opts.RawGetString("min"); minVal != lua.LNil {
		if num, ok := minVal.(lua.LNumber); ok {
			min = int(num)
		}
	}
	if maxVal := opts.RawGetString("max"); maxVal != lua.LNil {
		if num, ok := maxVal.(lua.LNumber); ok {
			max = int(num)
		}
	}

	valid := (min < 0 || length >= min) && (max < 0 || length <= max)

	L.Push(lua.LBool(valid))
	return 1
}

// luaRange checks numeric range
// Usage: local valid = validator.range(5, {min = 1, max = 10})
func luaRange(L *lua.LState) int {
	value := float64(L.CheckNumber(1))
	opts := L.CheckTable(2)

	min := -1.0
	max := -1.0

	if minVal := opts.RawGetString("min"); minVal != lua.LNil {
		if num, ok := minVal.(lua.LNumber); ok {
			min = float64(num)
		}
	}
	if maxVal := opts.RawGetString("max"); maxVal != lua.LNil {
		if num, ok := maxVal.(lua.LNumber); ok {
			max = float64(num)
		}
	}

	valid := (min < 0 || value >= min) && (max < 0 || value <= max)

	L.Push(lua.LBool(valid))
	return 1
}

var exports = map[string]lua.LGFunction{
	"is_email":        luaIsEmail,
	"is_url":          luaIsURL,
	"is_ip":           luaIsIP,
	"is_ipv4":         luaIsIPv4,
	"is_ipv6":         luaIsIPv6,
	"is_uuid":         luaIsUUID,
	"is_json":         luaIsJSON,
	"is_numeric":      luaIsNumeric,
	"is_alpha":        luaIsAlpha,
	"is_alphanumeric": luaIsAlphanumeric,
	"is_credit_card":  luaIsCreditCard,
	"is_phone":        luaIsPhone,
	"is_date":         luaIsDate,
	"is_base64":       luaIsBase64,
	"is_hex":          luaIsHex,
	"is_slug":         luaIsSlug,
	"matches":         luaMatches,
	"schema":          luaValidateSchema,
	"length":          luaLength,
	"range":           luaRange,
}

// Loader is called when the module is required via require("validator")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
