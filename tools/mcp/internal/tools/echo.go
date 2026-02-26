package tools

import (
	"fmt"
	"strings"
	"time"

	mcp "github.com/metoro-io/mcp-golang"
)

// EchoArguments echo工具的参数结构
type EchoArguments struct {
	Text      string `json:"text" jsonschema:"required,description=要处理的文本内容"`
	Uppercase *bool  `json:"uppercase" jsonschema:"description=是否转换为大写,default=false"`
	Lowercase *bool  `json:"lowercase" jsonschema:"description=是否转换为小写,default=false"`
	Prefix    *string `json:"prefix" jsonschema:"description=添加到文本前面的前缀"`
	Suffix    *string `json:"suffix" jsonschema:"description=添加到文本后面的后缀"`
	Repeat    *int   `json:"repeat" jsonschema:"description=重复次数(1-10),default=1,minimum=1,maximum=10"`
}

// EchoHandler echo工具的处理函数
func EchoHandler(arguments EchoArguments) (*mcp.ToolResponse, error) {
	startTime := time.Now()
	
	// 处理文本
	finalText := arguments.Text
	
	// 大小写转换（大写优先）
	if arguments.Uppercase != nil && *arguments.Uppercase {
		finalText = strings.ToUpper(finalText)
	} else if arguments.Lowercase != nil && *arguments.Lowercase {
		finalText = strings.ToLower(finalText)
	}
	
	// 添加前缀
	if arguments.Prefix != nil && *arguments.Prefix != "" {
		finalText = *arguments.Prefix + finalText
	}
	
	// 添加后缀
	if arguments.Suffix != nil && *arguments.Suffix != "" {
		finalText = finalText + *arguments.Suffix
	}
	
	// 重复处理
	repeat := 1
	if arguments.Repeat != nil {
		if *arguments.Repeat >= 1 && *arguments.Repeat <= 10 {
			repeat = *arguments.Repeat
		}
	}
	
	if repeat > 1 {
		repeated := make([]string, repeat)
		for i := 0; i < repeat; i++ {
			repeated[i] = finalText
		}
		finalText = strings.Join(repeated, " ")
	}
	
	// 创建响应内容
	responseText := fmt.Sprintf("Echo result: %s\n\nTransformations applied:\n- Uppercase: %t\n- Lowercase: %t\n- Prefix: '%s'\n- Suffix: '%s'\n- Repeat: %d times\n\nOriginal length: %d, Final length: %d, Duration: %dms",
		finalText, 
		arguments.Uppercase != nil && *arguments.Uppercase,
		arguments.Lowercase != nil && *arguments.Lowercase,
		func() string { if arguments.Prefix != nil { return *arguments.Prefix }; return "" }(),
		func() string { if arguments.Suffix != nil { return *arguments.Suffix }; return "" }(),
		repeat,
		len(arguments.Text), 
		len(finalText),
		time.Since(startTime).Milliseconds())
	
	return mcp.NewToolResponse(mcp.NewTextContent(responseText)), nil
}