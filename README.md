# Cut JSON

一个用于裁剪JSON内容的Go库，支持通过路径表达式提取JSON对象的特定部分，以及基于规则裁剪JSON结构体。

## 功能特点

- 支持通过点分隔的路径表达式提取JSON内容
- 支持嵌套对象和数组访问
- 支持负数索引访问数组元素（从末尾计数）
- 支持一次提取多个路径的内容
- 支持基于规则的JSON裁剪：
  - 规则1: 保留指定JSON路径
  - 规则2: 如果指定路径的值等于配置值，保留父路径
  - 规则3: 保留数组中满足特定条件的元素
- 提供清晰的错误处理

## 安装

```bash
go get github.com/ALONELUR/cut_json
```

## 使用示例

### 基本使用

```go
package main

import (
	"fmt"
	"log"

	"github.com/ALONELUR/cut_json"
)

func main() {
	jsonData := []byte(`{
	"user": {
		"name": "John Doe",
		"age": 30,
		"address": {
			"street": "123 Main St",
			"city": "Anytown"
		}
	},
	"orders": [
		{"id": 1, "product": "Laptop"},
		{"id": 2, "product": "Phone"}
	]
}`)

	// 提取用户名
	name, err := cutjson.Cut(jsonData, "user.name")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("User name: %s\n", name)

	// 提取地址信息
	address, err := cutjson.Cut(jsonData, "user.address")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("Address: %s\n", address)

	// 提取第一个订单
	firstOrder, err := cutjson.Cut(jsonData, "orders.0")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("First order: %s\n", firstOrder)

	// 提取最后一个订单（使用负索引）
	lastOrder, err := cutjson.Cut(jsonData, "orders.-1")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("Last order: %s\n", lastOrder)
}
```

### 提取多个路径

```go
package main

import (
	"fmt"
	"log"

	"github.com/ALONELUR/cut_json"
)

func main() {
	jsonData := []byte(`{
	"user": {
		"name": "John Doe",
		"age": 30,
		"address": {
			"street": "123 Main St",
			"city": "Anytown"
		}
	},
	"orders": [
		{"id": 1, "product": "Laptop"},
		{"id": 2, "product": "Phone"}
	]
}`)

	// 一次提取多个路径
	paths := []string{
		"user.name",
		"user.age",
		"orders.0.product",
		"orders.1.id",
	}

	results, err := cutjson.CutMultiple(jsonData, paths)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// 打印结果
	for path, value := range results {
		fmt.Printf("%s: %s\n", path, value)
	}
}
```

## 路径表达式语法

- 使用点（`.`）分隔路径段
- 对象属性直接使用属性名，如 `user.name`
- 数组元素使用索引，如 `orders.0`
- 支持负索引访问数组元素，如 `orders.-1` 表示最后一个元素

## 错误处理

库提供了以下错误类型：

- `ErrInvalidJSON`: 输入的JSON格式无效
- `ErrPathNotFound`: 指定的路径在JSON中不存在
- `ErrInvalidPath`: 路径格式无效（例如，数组索引不是数字）

## 基于规则的JSON裁剪

```go
package main

import (
	"fmt"
	"log"

	"github.com/ALONELUR/cut_json"
)

func main() {
	jsonData := []byte(`{
  "user": {
    "name": "John Doe",
    "age": 30,
    "verified": true,
    "preferences": {
      "theme": "dark",
      "notifications": true
    }
  },
  "products": [
    {
      "id": 101,
      "name": "Laptop",
      "category": "electronics",
      "inStock": true
    },
    {
      "id": 102,
      "name": "Smartphone",
      "category": "electronics",
      "inStock": false
    },
    {
      "id": 103,
      "name": "Headphones",
      "category": "accessories",
      "inStock": true
    }
  ]
}`)

	// 创建规则列表
	rules := []cutjson.Rule{
		// 规则1: 保留指定路径
		cutjson.NewKeepPathRule("user.name"),

		// 规则2: 如果用户已验证，保留verified字段及其父路径
		cutjson.NewKeepParentIfValueMatchesRule("user.verified", true),

		// 规则3: 保留库存中的电子产品
		cutjson.NewKeepArrayElementsIfChildValueMatchesRule(
			"products", "category", "electronics",
		),
	}

	// 应用规则
	result, err := cutjson.CutWithRules(jsonData, rules)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("Result: %s\n", result)
	// 输出包含用户名称、verified字段和库存中的电子产品
}
```

更多规则使用示例，请参阅 [examples/rules_usage.md](examples/rules_usage.md)。

## 命令行工具

库附带了一个命令行工具，可以直接在命令行中使用规则功能：

```bash
# 安装命令行工具
go install github.com/ALONELUR/cut_json/cmd/cut_json@latest

# 使用规则1: 保留指定路径
cut_json -file data.json -path "user.name,user.age" -pretty

# 使用规则2: 条件保留父路径
cut_json -file data.json -keep-if-value "user.preferences.theme=dark" -pretty

# 使用规则3: 数组元素条件过滤
cut_json -file data.json -keep-array-match "products:category=\"electronics\"" -pretty

# 组合多个规则
cut_json -file data.json \
  -path "user.name" \
  -keep-if-value "user.verified=true" \
  -keep-array-match "products:inStock=true" \
  -pretty

# 使用JSON配置文件定义规则
cut_json -file data.json -config rules_config.json -pretty
```

### 使用JSON配置文件

除了通过命令行参数定义规则外，还可以通过JSON配置文件定义规则：

```json
{
  "rules": [
    {
      "type": "keep_path",
      "where": "user.name"
    },
    {
      "type": "keep_parent_if_value_matches",
      "where": "user.preferences.theme",
      "op": "equals",
      "value": "dark"
    },
    {
      "type": "keep_array_elements_if_child_value_matches",
      "where": "products",
      "child_path": "category",
      "op": "equals",
      "value": "electronics"
    }
  ]
}
```

更多关于配置文件的详细信息，请参阅 [examples/config_usage.md](examples/config_usage.md)。

## 许可证

MIT