package core

import "strings"

// CommonMethods 定义了"*"通配符支持的方法
var CommonMethods = []string{"GET", "POST", "PUT", "DELETE"}

// ShouldMatchMethod 检查请求方法是否匹配任务配置
func (task *ProfilingTask) ShouldMatchMethod(requestMethod string) bool {
	// 如果没有指定方法，默认为GET
	if len(task.Methods) == 0 {
		return requestMethod == "GET"
	}
	
	// 如果包含"*"，匹配常用方法
	if contains(task.Methods, "*") {
		return contains(CommonMethods, requestMethod)
	}
	
	// 检查是否包含指定的方法
	return contains(task.Methods, requestMethod)
}

// GetEffectiveMethods 返回此任务将匹配的所有方法
func (task *ProfilingTask) GetEffectiveMethods() []string {
	// 如果没有指定方法，默认为GET
	if len(task.Methods) == 0 {
		return []string{"GET"}
	}
	
	// 如果包含"*"，展开为常用方法
	var result []string
	for _, method := range task.Methods {
		if method == "*" {
			result = append(result, CommonMethods...)
		} else {
			result = append(result, method)
		}
	}
	
	return result
}

// contains 检查切片是否包含指定字符串
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) { // 不区分大小写比较
			return true
		}
	}
	return false
}