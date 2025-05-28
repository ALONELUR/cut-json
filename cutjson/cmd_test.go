package cutjson

import (
	"encoding/json"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// 测试从命令行参数构建规则的函数
func TestBuildRules(t *testing.T) {
	Convey("测试从命令行参数构建规则", t, func() {
		Convey("规则1: 保留指定路径", func() {
			// 测试单个路径
			rules := buildRules("user.name", "", "")
			So(len(rules), ShouldEqual, 1)
			So(rules[0].Type, ShouldEqual, KeepPath)

			// 测试多个路径
			rules = buildRules("user.name,user.age", "", "")
			So(len(rules), ShouldEqual, 2)
			So(rules[0].Type, ShouldEqual, KeepPath)
			So(rules[1].Type, ShouldEqual, KeepPath)
		})

		Convey("规则2: 如果值匹配，保留父路径", func() {
			// 测试字符串值
			rules := buildRules("", "user.preferences.theme=dark", "")
			So(len(rules), ShouldEqual, 1)
			So(rules[0].Type, ShouldEqual, KeepParentIfValueMatches)

			// 测试布尔值
			rules = buildRules("", "user.verified=true", "")
			So(len(rules), ShouldEqual, 1)
			So(rules[0].Type, ShouldEqual, KeepParentIfValueMatches)

			// 测试数字值
			rules = buildRules("", "user.age=30", "")
			So(len(rules), ShouldEqual, 1)
			So(rules[0].Type, ShouldEqual, KeepParentIfValueMatches)

			// 测试多个条件
			rules = buildRules("", "user.verified=true,user.preferences.theme=dark", "")
			So(len(rules), ShouldEqual, 2)
			So(rules[0].Type, ShouldEqual, KeepParentIfValueMatches)
			So(rules[1].Type, ShouldEqual, KeepParentIfValueMatches)
		})

		Convey("规则3: 保留数组中满足条件的元素", func() {
			// 测试字符串值
			rules := buildRules("", "", "products:category=electronics")
			So(len(rules), ShouldEqual, 1)
			So(rules[0].Type, ShouldEqual, KeepArrayElementsIfChildValueMatches)

			// 测试布尔值
			rules = buildRules("", "", "products:inStock=true")
			So(len(rules), ShouldEqual, 1)
			So(rules[0].Type, ShouldEqual, KeepArrayElementsIfChildValueMatches)

			// 测试数字值
			rules = buildRules("", "", "products:price=199.99")
			So(len(rules), ShouldEqual, 1)
			So(rules[0].Type, ShouldEqual, KeepArrayElementsIfChildValueMatches)

			// 测试多个条件
			rules = buildRules("", "", "products:category=electronics,orders:status=shipped")
			So(len(rules), ShouldEqual, 2)
			So(rules[0].Type, ShouldEqual, KeepArrayElementsIfChildValueMatches)
			So(rules[1].Type, ShouldEqual, KeepArrayElementsIfChildValueMatches)
		})

		Convey("组合多个规则", func() {
			rules := buildRules("user.name", "user.verified=true", "products:inStock=true")
			So(len(rules), ShouldEqual, 3)
			So(rules[0].Type, ShouldEqual, KeepPath)
			So(rules[1].Type, ShouldEqual, KeepParentIfValueMatches)
			So(rules[2].Type, ShouldEqual, KeepArrayElementsIfChildValueMatches)
		})

		Convey("处理无效的规则格式", func() {
			// 无效的规则2格式
			rules := buildRules("", "user.preferences.theme", "")
			So(len(rules), ShouldEqual, 0)

			// 无效的规则3格式
			rules = buildRules("", "", "products")
			So(len(rules), ShouldEqual, 0)

			// 无效的规则3条件格式
			rules = buildRules("", "", "products:category")
			So(len(rules), ShouldEqual, 0)
		})
	})
}

// buildRules 根据命令行参数构建规则列表
func buildRules(paths, keepIfValue, keepArrayMatch string) []Rule {
	rules := []Rule{}

	// 处理规则1: 保留指定路径
	if paths != "" {
		pathList := strings.Split(paths, ",")
		for _, path := range pathList {
			path = strings.TrimSpace(path)
			if path != "" {
				rules = append(rules, NewKeepPathRule(path))
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
				// 忽略无效的格式
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

			rules = append(rules, NewKeepParentIfValueMatchesRule(path, parsedValue))
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
				// 忽略无效的格式
				continue
			}

			arrayPath := strings.TrimSpace(pathParts[0])
			condition := strings.TrimSpace(pathParts[1])

			// 分割子路径和值
			condParts := strings.SplitN(condition, "=", 2)
			if len(condParts) != 2 {
				// 忽略无效的条件格式
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

			rules = append(rules, NewKeepArrayElementsIfChildValueMatchesRule(arrayPath, childPath, parsedValue))
		}
	}

	return rules
}
