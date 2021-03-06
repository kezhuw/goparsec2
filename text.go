package goP2

import (
	"fmt"
	"unicode"
)

// Chr 判断下一个字符是否与给定值相等
func Chr(val rune) P {
	return func(state State) (interface{}, error) {
		x, err := state.Next()
		if err != nil {
			return nil, err
		}
		if c, ok := x.(int32); ok {
			if c == val {
				return c, nil
			}
			return nil, state.Trap("Expect '%v' but '%v'", string([]rune{val}), string([]rune{c}))
		}
		return nil, state.Trap("Expect a rune '%s' but x is %v", string([]rune{val}), x)
	}
}

// NChr 判断下一个字符是否与给定值不相等
func NChr(val rune) P {
	return func(state State) (interface{}, error) {
		x, err := state.Next()
		if err != nil {
			return nil, err
		}
		if c, ok := x.(int32); ok {
			if c == val {
				return nil, state.Trap("Expect not '%v' but '%v'", string([]rune{val}), string([]rune{c}))
			}
			return c, nil
		}
		return nil, state.Trap("Expect a rune '%s' but x is %v", string([]rune{val}), x)
	}
}

// RuneOf 检查后续的字符是否是给定值中的某一个
func RuneOf(str string) P {
	data := []rune(str)
	return func(state State) (interface{}, error) {
		x, err := state.Next()
		if err != nil {
			return nil, err
		}
		if c, ok := x.(int32); ok {
			for _, r := range data {
				if c == r {
					return c, nil
				}
			}
			return nil, state.Trap("Expect rune in '%s' but '%s'", str, string([]rune{c}))
		}
		return nil, state.Trap("Expect rune in '%s' but x=%v is %t", str, x, x)
	}
}

// RuneNone 检查后续的字符是否不是给定值中的任一个
func RuneNone(str string) P {
	data := []rune(str)
	return func(state State) (interface{}, error) {
		x, err := state.Next()
		if err != nil {
			return nil, err
		}
		if c, ok := x.(int32); ok {
			for _, r := range data {
				if c == r {
					return nil, state.Trap("Expect rune none of '%s' but '%s'", str, string([]rune{c}))
				}
			}
			return c, nil
		}
		return nil, state.Trap("Expect rune none of '%s' but x=%v is %t", str, x, x)
	}
}

// Str 判断后续的字符串是否匹配给定的串
func Str(str string) P {
	data := []rune(str)
	return func(state State) (interface{}, error) {
		for _, r := range data {
			_, err := Chr(r).Parse(state)
			if err != nil {
				return nil, err
			}
		}
		return str, nil
	}
}

// RuneP 通过一个谓词参数，提供通用的 rune 算子生成判断
func RuneP(name string, pred func(r rune) bool) P {
	return func(state State) (interface{}, error) {
		x, err := state.Next()
		if err != nil {
			return nil, err
		}
		if c, ok := x.(int32); ok {
			r := rune(c)
			if pred(r) {
				return c, nil
			}
			return nil, state.Trap("Expect %s but '%v'", name, string([]rune{r}))
		}
		return nil, state.Trap("Expect %s but x=%v is %t", name, x, x)
	}
}

// Space 构造一个空格校验算子
func Space(state State) (interface{}, error) {
	return RuneP("space", unicode.IsSpace)(state)
}

// Newline 构造一个换行校验算子
func Newline(state State) (interface{}, error) {
	return RuneOf("\r\n")(state)
}

// Letter 构造一个字母校验算子
func Letter(state State) (interface{}, error) {
	return RuneP("letter", unicode.IsLetter)(state)
}

// Number 构造一个 Number 校验算子
func Number(state State) (interface{}, error) {
	return RuneP("number", unicode.IsNumber)(state)
}

// Digit 构造一个数字字符校验算子
func Digit(state State) (interface{}, error) {
	return RuneP("digit", unicode.IsDigit)(state)
}

// UInt 返回一个无符号整型的解析算子
func UInt(state State) (interface{}, error) {
	return Do(func(st State) interface{} {
		digits := Many1(Digit).Exec(st)
		buffer := digits.([]interface{})
		data := make([]rune, 0, len(buffer))
		for _, value := range buffer {
			data = append(data, value.(rune))
		}
		return string(data)
	})(state)
}

// Int 返回一个有符号整型的解析算子
func Int(state State) (interface{}, error) {
	binder := func(value interface{}) P {
		return Return(fmt.Sprintf("-%s", value.(string)))
	}
	return Choice(Try(Chr('-').Then(UInt).Bind(binder)), UInt)(state)
}

// UFloat 返回一个无符号实数的解析算子
func UFloat(state State) (interface{}, error) {
	return Do(func(state State) interface{} {
		left := Choice(Try(P(UInt).Over(Chr('.'))), Chr('.').Then(Return("0"))).Exec(state)
		right := P(UInt).Exec(state)
		return fmt.Sprintf("%s.%s", left, right)
	})(state)
}

// Float 返回一个有符号实数的解析算子
func Float(state State) (interface{}, error) {
	binder := func(value interface{}) P {
		return Return("-" + value.(string))
	}
	return Choice(Try(Chr('-').Then(UFloat).Bind(binder)), UFloat)(state)
}

// ToString 将封装为 interface{} 的 []interface{} 转成 string，如果输入数据与前面提到的规范不符，会 panic
func ToString(input interface{}) string {
	data := input.([]interface{})
	l := len(data)
	buffer := make([]rune, l)
	for index, item := range data {
		buffer[index] = item.(rune)
	}
	return string(buffer)
}

// ToInterfaces 将string 转为 []interface{} ，如果输入数据与前面提到的规范不符，会 panic
func ToInterfaces(input string) []interface{} {
	buf := []rune(input)
	re := make([]interface{}, 0, len(buf))
	for r := range buf {
		re = append(re, r)
	}
	return re
}

// ReturnString 用 Return 包装 ToString，使其适用于组合子表达式。
func ReturnString(input interface{}) P {
	return Return(ToString(input))
}
