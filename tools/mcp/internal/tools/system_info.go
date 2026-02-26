package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	mcp "github.com/metoro-io/mcp-golang"
)

// SystemInfoArguments system_info工具的参数结构
type SystemInfoArguments struct {
	Category *string `json:"category" jsonschema:"description=信息类别: all, runtime, memory, environment, process,default=all,enum=all,enum=runtime,enum=memory,enum=environment,enum=process"`
}

// SystemInfoHandler system_info工具的处理函数
func SystemInfoHandler(arguments SystemInfoArguments) (*mcp.ToolResponse, error) {
	startTime := time.Now()
	
	category := "all"
	if arguments.Category != nil && *arguments.Category != "" {
		category = *arguments.Category
	}
	
	var output []string
	
	switch category {
	case "runtime":
		output = append(output, getRuntimeInfo()...)
	case "memory":
		output = append(output, getMemoryInfo()...)
	case "environment":
		output = append(output, getEnvironmentInfo()...)
	case "process":
		output = append(output, getProcessInfo()...)
	case "all":
		output = append(output, "=== System Information ===")
		output = append(output, "")
		output = append(output, "--- Runtime Information ---")
		output = append(output, getRuntimeInfo()...)
		output = append(output, "")
		output = append(output, "--- Memory Information ---")
		output = append(output, getMemoryInfo()...)
		output = append(output, "")
		output = append(output, "--- Environment Information ---")
		output = append(output, getEnvironmentInfo()...)
		output = append(output, "")
		output = append(output, "--- Process Information ---")
		output = append(output, getProcessInfo()...)
	default:
		return nil, fmt.Errorf("invalid category: %s. Valid categories: all, runtime, memory, environment, process", category)
	}
	
	// 添加执行信息
	output = append(output, "")
	output = append(output, fmt.Sprintf("--- Execution Information ---"))
	output = append(output, fmt.Sprintf("Query Category: %s", category))
	output = append(output, fmt.Sprintf("Execution Time: %dms", time.Since(startTime).Milliseconds()))
	output = append(output, fmt.Sprintf("Timestamp: %s", time.Now().UTC().Format(time.RFC3339)))
	
	result := strings.Join(output, "\n")
	
	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// getRuntimeInfo 获取运行时信息
func getRuntimeInfo() []string {
	var info []string
	info = append(info, fmt.Sprintf("Go Version: %s", runtime.Version()))
	info = append(info, fmt.Sprintf("Go OS: %s", runtime.GOOS))
	info = append(info, fmt.Sprintf("Go Arch: %s", runtime.GOARCH))
	info = append(info, fmt.Sprintf("CPU Count: %d", runtime.NumCPU()))
	info = append(info, fmt.Sprintf("Goroutines: %d", runtime.NumGoroutine()))
	info = append(info, fmt.Sprintf("CGO Calls: %d", runtime.NumCgoCall()))
	return info
}

// getMemoryInfo 获取内存信息
func getMemoryInfo() []string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	var info []string
	info = append(info, fmt.Sprintf("Allocated Memory: %d KB", bToKb(m.Alloc)))
	info = append(info, fmt.Sprintf("Total Allocated: %d KB", bToKb(m.TotalAlloc)))
	info = append(info, fmt.Sprintf("System Memory: %d KB", bToKb(m.Sys)))
	info = append(info, fmt.Sprintf("GC Runs: %d", m.NumGC))
	info = append(info, fmt.Sprintf("Last GC: %s", time.Unix(0, int64(m.LastGC)).Format(time.RFC3339)))
	return info
}

// getEnvironmentInfo 获取环境信息
func getEnvironmentInfo() []string {
	var info []string
	
	// 获取工作目录
	if wd, err := os.Getwd(); err == nil {
		info = append(info, fmt.Sprintf("Working Directory: %s", wd))
	}
	
	// 获取可执行文件路径
	if exe, err := os.Executable(); err == nil {
		info = append(info, fmt.Sprintf("Executable: %s", exe))
		info = append(info, fmt.Sprintf("Executable Dir: %s", filepath.Dir(exe)))
	}
	
	// 获取一些关键环境变量
	envVars := []string{"PATH", "HOME", "USER", "SHELL", "LANG", "GOPATH", "GOROOT"}
	for _, envVar := range envVars {
		if value := os.Getenv(envVar); value != "" {
			// 对于PATH，只显示前100个字符以避免输出过长
			if envVar == "PATH" && len(value) > 100 {
				value = value[:100] + "..."
			}
			info = append(info, fmt.Sprintf("%s: %s", envVar, value))
		}
	}
	
	return info
}

// getProcessInfo 获取进程信息
func getProcessInfo() []string {
	var info []string
	info = append(info, fmt.Sprintf("Process ID: %d", os.Getpid()))
	info = append(info, fmt.Sprintf("Parent Process ID: %d", os.Getppid()))
	
	// 获取命令行参数
	args := os.Args
	if len(args) > 0 {
		info = append(info, fmt.Sprintf("Command: %s", args[0]))
		if len(args) > 1 {
			info = append(info, fmt.Sprintf("Arguments: %s", strings.Join(args[1:], " ")))
		}
	}
	
	return info
}

// bToKb 将字节转换为KB
func bToKb(b uint64) uint64 {
	return b / 1024
}