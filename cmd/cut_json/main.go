package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/ALONELUR/cut_json/cutjson"
)

func main() {
	// 定义命令行参数
	var (
		filePath       string
		paths          string
		keepIfValue    string
		keepArrayMatch string
		configPath     string
		prettyOut      bool
	)

	flag.StringVar(&filePath, "file", "", "JSON文件路径 (如果不提供，则从标准输入读取)")
	flag.StringVar(&paths, "path", "", "规则1: 要保留的路径，多个路径用逗号分隔")
	flag.StringVar(&keepIfValue, "keep-if-value", "", "规则2: 格式为'路径=值'，如果指定路径的值等于配置值，则保留父路径")
	flag.StringVar(&keepArrayMatch, "keep-array-match", "", "规则3: 格式为'数组路径:子路径=值'，保留数组中满足子路径值为配置值的元素")
	flag.StringVar(&configPath, "config", "", "JSON配置文件路径，用于从配置文件加载规则")
	flag.BoolVar(&prettyOut, "pretty", false, "是否美化输出的JSON")
	flag.Parse()

	// 检查是否提供了至少一个规则或配置文件
	if paths == "" && keepIfValue == "" && keepArrayMatch == "" && configPath == "" {
		fmt.Println("错误: 必须提供至少一个规则参数或配置文件")
		flag.Usage()
		os.Exit(1)
	}

	// 读取JSON数据
	var jsonData []byte
	var err error

	if filePath != "" {
		// 从文件读取
		jsonData, err = os.ReadFile(filePath)
		if err != nil {
			log.Fatalf("无法读取文件 %s: %v", filePath, err)
		}
	} else {
		// 从标准输入读取
		jsonData, err = io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("无法从标准输入读取: %v", err)
		}
	}

	// 构建规则列表
	var rules []cutjson.Rule

	if configPath != "" {
		// 从配置文件加载规则
		rules, err = cutjson.LoadRulesFromConfig(configPath)
		if err != nil {
			log.Fatalf("从配置文件加载规则时出错: %v", err)
		}

		// 如果同时提供了命令行规则，则合并规则
		if paths != "" || keepIfValue != "" || keepArrayMatch != "" {
			cmdRules := buildRules(paths, keepIfValue, keepArrayMatch)
			rules = append(rules, cmdRules...)
		}
	} else {
		// 仅使用命令行规则
		rules = buildRules(paths, keepIfValue, keepArrayMatch)
	}

	// 应用规则
	result, err := cutjson.CutWithRules(jsonData, rules)
	if err != nil {
		log.Fatalf("应用规则时出错: %v", err)
	}

	// 输出结果
	printJSON(result, prettyOut)
}

// buildRules 根据命令行参数构建规则列表
func buildRules(paths, keepIfValue, keepArrayMatch string) []cutjson.Rule {
	rules := []cutjson.Rule{}

	// 处理规则1: 保留指定路径
	if paths != "" {
		pathList := strings.Split(paths, ",")
		for _, path := range pathList {
			path = strings.TrimSpace(path)
			if path != "" {
				rules = append(rules, cutjson.NewKeepPathRule(path))
			}
		}
	}

	// 处理规则2: 如果值匹配，保留父路径
	if keepIfValue != "" {
		pairs := strings.Split(keepIfValue, ",")
		for _, pair := range pairs {
			pair = strings.TrimSpace(pair)
			if pair == "" {
				continue
			}

			parts := strings.SplitN(pair, "=", 2)
			if len(parts) != 2 {
				log.Printf("警告: 忽略无效的规则2格式: %s", pair)
				continue
			}

			path := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// 尝试将值解析为JSON
			var parsedValue interface{}
			if err := json.Unmarshal([]byte(value), &parsedValue); err != nil {
				// 如果不是有效的JSON，则视为字符串
				parsedValue = value
			}

			rules = append(rules, cutjson.NewKeepParentIfValueMatchesRule(path, parsedValue))
		}
	}

	// 处理规则3: 保留数组中满足条件的元素
	if keepArrayMatch != "" {
		pairs := strings.Split(keepArrayMatch, ",")
		for _, pair := range pairs {
			pair = strings.TrimSpace(pair)
			if pair == "" {
				continue
			}

			// 分割数组路径和条件
			pathParts := strings.SplitN(pair, ":", 2)
			if len(pathParts) != 2 {
				log.Printf("警告: 忽略无效的规则3格式: %s", pair)
				continue
			}

			arrayPath := strings.TrimSpace(pathParts[0])
			condition := strings.TrimSpace(pathParts[1])

			// 分割子路径和值
			condParts := strings.SplitN(condition, "=", 2)
			if len(condParts) != 2 {
				log.Printf("警告: 忽略无效的规则3条件格式: %s", condition)
				continue
			}

			childPath := strings.TrimSpace(condParts[0])
			value := strings.TrimSpace(condParts[1])

			// 尝试将值解析为JSON
			var parsedValue interface{}
			if err := json.Unmarshal([]byte(value), &parsedValue); err != nil {
				// 如果不是有效的JSON，则视为字符串
				parsedValue = value
			}

			rules = append(rules, cutjson.NewKeepArrayElementsIfChildValueMatchesRule(arrayPath, childPath, parsedValue))
		}
	}

	return rules
}

// printJSON 根据是否需要美化输出JSON
func printJSON(data []byte, pretty bool) {
	if !pretty {
		fmt.Println(string(data))
		return
	}

	// 美化JSON输出
	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		// 如果无法解析为JSON对象，直接输出原始内容
		fmt.Println(string(data))
		return
	}

	prettyJSON, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		// 如果美化失败，输出原始内容
		fmt.Println(string(data))
		return
	}

	fmt.Println(string(prettyJSON))
}
