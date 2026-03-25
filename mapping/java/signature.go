package java

import (
	"bytes"
	"fmt"
	"strings"
)

func MethodToByteCodeSignature(javaSignature string, keepName bool) (string, error) {
	// 去除首尾空格
	javaSignature = strings.TrimSpace(javaSignature)

	// 检查是否包含括号
	leftParen := strings.Index(javaSignature, "(")
	rightParen := strings.Index(javaSignature, ")")

	if leftParen == -1 || rightParen == -1 || leftParen > rightParen {
		return "", fmt.Errorf("无效的方法签名: 缺少括号或括号位置不正确")
	}

	// 提取返回类型和参数部分
	returnTypePart := strings.TrimSpace(javaSignature[:leftParen])
	paramsPart := javaSignature[leftParen+1 : rightParen]

	// 解析返回类型
	returnType := ClassToByteCodeSignature(strings.Split(returnTypePart, " ")[0])

	// 解析参数
	var paramTypes []string
	if paramsPart != "" {
		params := strings.Split(paramsPart, ",")
		for _, param := range params {
			param = strings.TrimSpace(param)
			if param == "" {
				continue
			}

			// 处理带变量名的参数（如 "int x"）
			parts := strings.Fields(param)
			if len(parts) == 0 {
				continue
			}

			paramType := parts[0] // 取类型部分
			parsedType := ClassToByteCodeSignature(paramType)
			paramTypes = append(paramTypes, parsedType)
		}
	}

	// 构建字节码签名
	var result bytes.Buffer
	result.WriteByte('(')
	for _, paramType := range paramTypes {
		result.WriteString(paramType)
	}
	result.WriteByte(')')
	result.WriteString(returnType)

	if keepName {
		result.WriteString(" ")
		result.WriteString(strings.Split(returnTypePart, " ")[1])
		return result.String(), nil
	} else {
		return result.String(), nil
	}
}

func ClassToByteCodeSignature(javaType string) string {
	javaType = strings.TrimSpace(javaType)

	// 处理数组类型
	if elementType, ok := strings.CutSuffix(javaType, "[]"); ok {
		return "[" + ClassToByteCodeSignature(elementType)
	}

	// 处理基本类型
	switch javaType {
	case "void":
		return "V"
	case "boolean":
		return "Z"
	case "byte":
		return "B"
	case "char":
		return "C"
	case "short":
		return "S"
	case "int":
		return "I"
	case "long":
		return "J"
	case "float":
		return "F"
	case "double":
		return "D"
	default:
		// 处理对象类型
		if strings.ContainsRune(javaType, '.') {
			// 转换点号为斜杠
			javaType = strings.ReplaceAll(javaType, ".", "/")
		}
		return "L" + javaType + ";"
	}
}

func FullToClassName(full string) string {
	split := strings.Split(full, ".")
	return split[len(split)-1]
}
