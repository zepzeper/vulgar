package strings

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func newTestState() *lua.LState {
	L := lua.NewState()
	L.PreloadModule(ModuleName, Loader)
	return L
}

// =============================================================================
// trim tests
// =============================================================================

func TestTrimWhitespace(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local result = strings.trim("  hello  ")
		assert(result == "hello", "should trim whitespace from both ends")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestTrimNoWhitespace(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local result = strings.trim("hello")
		assert(result == "hello", "should return unchanged string")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestTrimTabs(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local result = strings.trim("\t\thello\t\t")
		assert(result == "hello", "should trim tabs")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// trim_left / trim_right tests
// =============================================================================

func TestTrimLeft(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local result = strings.trim_left("  hello  ")
		assert(result == "hello  ", "should trim left whitespace only")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestTrimRight(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local result = strings.trim_right("  hello  ")
		assert(result == "  hello", "should trim right whitespace only")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// split tests
// =============================================================================

func TestSplitByComma(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local parts = strings.split("a,b,c", ",")
		assert(#parts == 3, "should have 3 parts")
		assert(parts[1] == "a", "first part should be 'a'")
		assert(parts[2] == "b", "second part should be 'b'")
		assert(parts[3] == "c", "third part should be 'c'")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSplitNoDelimiter(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local parts = strings.split("hello", ",")
		assert(#parts == 1, "should have 1 part")
		assert(parts[1] == "hello", "part should be original string")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSplitEmpty(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local parts = strings.split("", ",")
		assert(#parts == 1, "should have 1 part")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// join tests
// =============================================================================

func TestJoinWithComma(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local result = strings.join({"a", "b", "c"}, ",")
		assert(result == "a,b,c", "should join with comma")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestJoinSingleElement(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local result = strings.join({"hello"}, ",")
		assert(result == "hello", "single element should not have separator")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestJoinEmpty(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local result = strings.join({}, ",")
		assert(result == "", "empty table should produce empty string")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// replace tests
// =============================================================================

func TestReplaceFirst(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local result = strings.replace("hello world world", "world", "lua")
		assert(result == "hello lua world", "should replace first occurrence only")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestReplaceNoMatch(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local result = strings.replace("hello world", "xyz", "abc")
		assert(result == "hello world", "should return unchanged string")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// replace_all tests
// =============================================================================

func TestReplaceAllOccurrences(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local result = strings.replace_all("aaa", "a", "b")
		assert(result == "bbb", "should replace all occurrences")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// to_upper / to_lower tests
// =============================================================================

func TestToUpper(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local result = strings.to_upper("hello")
		assert(result == "HELLO", "should convert to uppercase")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestToLower(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local result = strings.to_lower("HELLO")
		assert(result == "hello", "should convert to lowercase")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// capitalize / title tests
// =============================================================================

func TestCapitalize(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local result = strings.capitalize("hello")
		assert(result == "Hello", "should capitalize first letter")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestTitle(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local result = strings.title("hello world")
		assert(result == "Hello World", "should title case all words")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// case conversion tests
// =============================================================================

func TestSnakeCase(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local result = strings.snake_case("helloWorld")
		assert(result == "hello_world", "should convert to snake_case")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestCamelCase(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local result = strings.camel_case("hello_world")
		assert(result == "helloWorld", "should convert to camelCase")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestKebabCase(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local result = strings.kebab_case("hello_world")
		assert(result == "hello-world", "should convert to kebab-case")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestPascalCase(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local result = strings.pascal_case("hello_world")
		assert(result == "HelloWorld", "should convert to PascalCase")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// contains / starts_with / ends_with tests
// =============================================================================

func TestContains(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		assert(strings.contains("hello world", "world") == true, "should find 'world'")
		assert(strings.contains("hello world", "xyz") == false, "should not find 'xyz'")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestStartsWith(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		assert(strings.starts_with("hello world", "hello") == true, "should start with 'hello'")
		assert(strings.starts_with("hello world", "world") == false, "should not start with 'world'")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestEndsWith(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		assert(strings.ends_with("hello world", "world") == true, "should end with 'world'")
		assert(strings.ends_with("hello world", "hello") == false, "should not end with 'hello'")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// pad_left / pad_right tests
// =============================================================================

func TestPadLeft(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local result = strings.pad_left("42", 5, "0")
		assert(result == "00042", "should pad with zeros on left")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestPadRight(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local result = strings.pad_right("hi", 5)
		assert(#result == 5, "should pad to length 5")
		assert(string.sub(result, 1, 2) == "hi", "original string should be at start")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// reverse tests
// =============================================================================

func TestReverse(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local result = strings.reverse("hello")
		assert(result == "olleh", "should reverse string")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestReverseEmpty(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local result = strings.reverse("")
		assert(result == "", "empty string should remain empty")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// truncate tests
// =============================================================================

func TestTruncateWithSuffix(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local result = strings.truncate("hello world", 8, "...")
		assert(#result <= 8, "should be truncated to max length")
		assert(string.sub(result, -3) == "...", "should end with suffix")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestTruncateNoTruncation(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local result = strings.truncate("hello", 10, "...")
		assert(result == "hello", "should not truncate short string")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// slugify tests
// =============================================================================

func TestSlugify(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local result = strings.slugify("Hello World!")
		assert(result == "hello-world", "should create URL-safe slug")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSlugifySpecialChars(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local strings = require("stdlib.strings")
		local result = strings.slugify("Hello & World @ 2024")
		assert(string.find(result, "&") == nil, "should remove special chars")
		assert(string.find(result, "@") == nil, "should remove special chars")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
