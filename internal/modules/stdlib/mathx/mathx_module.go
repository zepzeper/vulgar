package mathx

import (
	"math"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
)

const ModuleName = "stdlib.mathx"

// collectNumbers collects numbers from arguments or a table
func collectNumbers(L *lua.LState) []float64 {
	var numbers []float64

	// Check if first argument is a table
	if L.GetTop() == 1 {
		if tbl, ok := L.Get(1).(*lua.LTable); ok {
			tbl.ForEach(func(_, v lua.LValue) {
				if num, ok := v.(lua.LNumber); ok {
					numbers = append(numbers, float64(num))
				}
			})
			return numbers
		}
	}

	// Otherwise collect from all arguments
	n := L.GetTop()
	for i := 1; i <= n; i++ {
		if num, ok := L.Get(i).(lua.LNumber); ok {
			numbers = append(numbers, float64(num))
		}
	}
	return numbers
}

// luaRound rounds a number to the specified decimal places
// Usage: local result = mathx.round(3.14159, 2) -- returns 3.14
func luaRound(L *lua.LState) int {
	value := float64(L.CheckNumber(1))
	places := L.CheckInt(2)

	multiplier := math.Pow(10, float64(places))
	result := math.Round(value*multiplier) / multiplier

	L.Push(lua.LNumber(result))
	return 1
}

// luaClamp clamps a value between min and max
// Usage: local result = mathx.clamp(value, min, max)
func luaClamp(L *lua.LState) int {
	value := float64(L.CheckNumber(1))
	min := float64(L.CheckNumber(2))
	max := float64(L.CheckNumber(3))

	if value < min {
		value = min
	} else if value > max {
		value = max
	}

	L.Push(lua.LNumber(value))
	return 1
}

// luaLerp performs linear interpolation between two values
// Usage: local result = mathx.lerp(a, b, t)
func luaLerp(L *lua.LState) int {
	a := float64(L.CheckNumber(1))
	b := float64(L.CheckNumber(2))
	t := float64(L.CheckNumber(3))

	result := a + t*(b-a)
	L.Push(lua.LNumber(result))
	return 1
}

// luaMap maps a value from one range to another
// Usage: local result = mathx.map(value, in_min, in_max, out_min, out_max)
func luaMap(L *lua.LState) int {
	value := float64(L.CheckNumber(1))
	inMin := float64(L.CheckNumber(2))
	inMax := float64(L.CheckNumber(3))
	outMin := float64(L.CheckNumber(4))
	outMax := float64(L.CheckNumber(5))

	// Normalize to [0, 1] range
	normalized := (value - inMin) / (inMax - inMin)
	// Map to output range
	result := outMin + normalized*(outMax-outMin)

	L.Push(lua.LNumber(result))
	return 1
}

// luaSign returns the sign of a number (-1, 0, or 1)
// Usage: local result = mathx.sign(value)
func luaSign(L *lua.LState) int {
	value := float64(L.CheckNumber(1))

	var sign float64
	if value > 0 {
		sign = 1
	} else if value < 0 {
		sign = -1
	} else {
		sign = 0
	}

	L.Push(lua.LNumber(sign))
	return 1
}

// luaSum returns the sum of all arguments or table values
// Usage: local result = mathx.sum(1, 2, 3) or mathx.sum({1, 2, 3})
func luaSum(L *lua.LState) int {
	numbers := collectNumbers(L)

	var sum float64
	for _, num := range numbers {
		sum += num
	}

	L.Push(lua.LNumber(sum))
	return 1
}

// luaAvg returns the average of all arguments or table values
// Usage: local result = mathx.avg(1, 2, 3) or mathx.avg({1, 2, 3})
func luaAvg(L *lua.LState) int {
	numbers := collectNumbers(L)

	if len(numbers) == 0 {
		L.Push(lua.LNumber(0))
		return 1
	}

	var sum float64
	for _, num := range numbers {
		sum += num
	}

	avg := sum / float64(len(numbers))
	L.Push(lua.LNumber(avg))
	return 1
}

// luaMin returns the minimum value from arguments or table
// Usage: local result = mathx.min(1, 2, 3) or mathx.min({1, 2, 3})
func luaMin(L *lua.LState) int {
	numbers := collectNumbers(L)

	if len(numbers) == 0 {
		L.Push(lua.LNumber(0))
		return 1
	}

	min := numbers[0]
	for _, num := range numbers[1:] {
		if num < min {
			min = num
		}
	}

	L.Push(lua.LNumber(min))
	return 1
}

// luaMax returns the maximum value from arguments or table
// Usage: local result = mathx.max(1, 2, 3) or mathx.max({1, 2, 3})
func luaMax(L *lua.LState) int {
	numbers := collectNumbers(L)

	if len(numbers) == 0 {
		L.Push(lua.LNumber(0))
		return 1
	}

	max := numbers[0]
	for _, num := range numbers[1:] {
		if num > max {
			max = num
		}
	}

	L.Push(lua.LNumber(max))
	return 1
}

var exports = map[string]lua.LGFunction{
	"round": luaRound,
	"clamp": luaClamp,
	"lerp":  luaLerp,
	"map":   luaMap,
	"sign":  luaSign,
	"sum":   luaSum,
	"avg":   luaAvg,
	"min":   luaMin,
	"max":   luaMax,
}

// Loader is called when the module is required via require("mathx")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
