# 使用JSON配置文件定义规则

除了通过命令行参数定义规则外，Cut JSON 工具还支持通过JSON配置文件定义规则。这种方式更加灵活，特别是在规则较多或较复杂的情况下。

## 配置文件格式

配置文件是一个JSON文件，包含一个`rules`数组，每个规则定义为一个对象，具有以下字段：

- `type`: 规则类型，可以是以下值之一：
  - `keep_path`: 保留指定路径（规则1）
  - `keep_parent_if_value_matches`: 如果值匹配，保留父路径（规则2）
  - `keep_array_elements_if_child_value_matches`: 保留数组中满足条件的元素（规则3）
- `where`: 指定JSON路径
- `child_path`: 子路径（仅用于规则3）
- `op`: 操作符，目前仅支持`equals`（仅用于规则2和规则3）
- `value`: 用于比较的值（仅用于规则2和规则3）

## 示例配置文件

```json
{
  "rules": [
    {
      "type": "keep_path",
      "where": "user.name"
    },
    {
      "type": "keep_path",
      "where": "user.age"
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
    },
    {
      "type": "keep_array_elements_if_child_value_matches",
      "where": "products",
      "child_path": "inStock",
      "op": "equals",
      "value": true
    }
  ]
}
```

## 使用配置文件

使用`-config`参数指定配置文件路径：

```bash
cut_json -file data.json -config rules_config.json -pretty
```

## 组合使用配置文件和命令行规则

您可以同时使用配置文件和命令行规则。在这种情况下，两种方式定义的规则将被合并：

```bash
cut_json -file data.json -config rules_config.json -path "orders.0" -pretty
```

这将应用配置文件中定义的所有规则，以及命令行中通过`-path`参数定义的额外规则。

## 值类型

在配置文件中，`value`字段可以是任何有效的JSON值：

- 字符串: `"value": "dark"`
- 数字: `"value": 42`
- 布尔值: `"value": true`
- null: `"value": null`
- 数组: `"value": [1, 2, 3]`
- 对象: `"value": {"key": "value"}`

## 注意事项

1. 目前，`op`字段仅支持`equals`操作符，表示精确匹配。
2. 配置文件中的规则将按照定义的顺序应用。
3. 如果配置文件格式不正确，程序将报错并退出。