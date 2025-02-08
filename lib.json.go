package main

import (
	"encoding/json"
	"strconv"

	lua "github.com/yuin/gopher-lua"
)

// 注册JSON模块
func jsonLoader(L *lua.LState) int {
	api := map[string]lua.LGFunction{
		"encode": jsonEncode,
		"decode": jsonDecode,
	}
	L.Push(L.SetFuncs(L.NewTable(), api))
	return 1
}

// json.encode 实现
func jsonEncode(L *lua.LState) int {
	value := L.CheckAny(1)
	data, err := json.Marshal(lValueToJson(value))
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LString(data))
	return 1
}

// json.decode 实现
func jsonDecode(L *lua.LState) int {
	str := L.CheckString(1)
	var goValue interface{}
	if err := json.Unmarshal([]byte(str), &goValue); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	luaValue := jsonToLValue(L, goValue)
	L.Push(luaValue)
	return 1
}

func lValueToJson(lv lua.LValue) interface{} {
	switch v := lv.(type) {
	case *lua.LTable:
		return tableToJson(v)
	case lua.LString:
		return string(v)
	case lua.LNumber:
		return float64(v)
	case lua.LBool:
		return bool(v)
	case *lua.LNilType:
		return []byte(`null`)
	default:
		return v.String() // 处理函数等其他类型
	}
}

// Table深度转换（处理数组/对象）
func tableToJson(tbl *lua.LTable) interface{} {
	if _isArray(tbl) {
		return convertJsonArray(tbl)
	}
	return convertJsonObject(tbl)
}

// 判断是否为数组
func _isArray(tbl *lua.LTable) bool {
	maxIndex := tbl.Len()
	if maxIndex == 0 {
		return false
	}

	// 检查所有键是否为连续整数
	isArr := true
	tbl.ForEach(func(k, _ lua.LValue) {
		if num, ok := k.(lua.LNumber); ok {
			idx := int(num)
			if idx < 1 || idx > maxIndex || float64(idx) != float64(num) {
				isArr = false
			}
		} else {
			isArr = false
		}
	})

	// 检查是否有空洞
	if isArr {
		for i := 1; i <= maxIndex; i++ {
			if tbl.RawGetInt(i) == lua.LNil {
				isArr = false
				break
			}
		}
	}
	return isArr
}

// 转换数组
func convertJsonArray(tbl *lua.LTable) []interface{} {
	arr := make([]interface{}, tbl.Len())
	for i := 1; i <= len(arr); i++ {
		arr[i-1] = lValueToJson(tbl.RawGetInt(i))
	}
	return arr
}

// 转换对象
func convertJsonObject(tbl *lua.LTable) map[string]interface{} {
	obj := make(map[string]interface{})
	tbl.ForEach(func(k, v lua.LValue) {
		key := lValueToKey(k)
		obj[key] = lValueToJson(v)
	})
	return obj
}

// Lua值转map键
func lValueToKey(lv lua.LValue) string {
	switch v := lv.(type) {
	case lua.LNumber:
		// 如果是整数则转换为数字形式
		if float64(int(v)) == float64(v) {
			return strconv.Itoa(int(v))
		}
		return strconv.FormatFloat(float64(v), 'f', -1, 64)
	case lua.LString:
		return string(v)
	default:
		return v.String()
	}
}

// JSON值转Lua值（用于json.decode）
func jsonToLValue(L *lua.LState, value interface{}) lua.LValue {
	switch v := value.(type) {
	case map[string]interface{}:
		return jsonObjectToTable(L, v)
	case []interface{}:
		return jsonArrayToTable(L, v)
	case float64:
		return lua.LNumber(v)
	case string:
		return lua.LString(v)
	case bool:
		return lua.LBool(v)
	case nil:
		return lua.LNil
	default:
		return lua.LNil
	}
}

// JSON对象转Lua Table
func jsonObjectToTable(L *lua.LState, obj map[string]interface{}) *lua.LTable {
	tbl := L.NewTable()
	for key, val := range obj {
		tbl.RawSetString(key, jsonToLValue(L, val))
	}
	return tbl
}

// JSON数组转Lua Table
func jsonArrayToTable(L *lua.LState, arr []interface{}) *lua.LTable {
	tbl := L.NewTable()
	for i, val := range arr {
		tbl.RawSetInt(i+1, jsonToLValue(L, val)) // Lua数组索引从1开始
	}
	return tbl
}
