# Cut JSON 规则使用示例

本文档演示如何使用 Cut JSON 工具的规则功能来裁剪 JSON 结构体。

## 规则类型

Cut JSON 支持三种规则类型：

1. **规则1: 保留指定路径** - 保留 JSON 中指定路径的值
2. **规则2: 条件保留父路径** - 如果指定路径的值等于配置值，则保留其父路径
3. **规则3: 数组元素条件过滤** - 保留数组中满足特定条件的元素

## 示例数据

以下示例基于 `rules_example.json` 文件中的数据：

```json
{
  "user": {
    "name": "John Doe",
    "age": 30,
    "verified": true,
    "address": {
      "street": "123 Main St",
      "city": "Anytown",
      "zipcode": "12345"
    },
    "preferences": {
      "theme": "dark",
      "notifications": true
    }
  },
  "products": [...],
  "orders": [...],
  "settings": {...}
}
```

## 规则1: 保留指定路径

### 命令示例

```bash
# 保留用户名称和年龄
cut_json -file rules_example.json -path "user.name,user.age" -pretty
```

### 预期输出

```json
{
  "user": {
    "name": "John Doe",
    "age": 30
  }
}
```

## 规则2: 条件保留父路径

### 命令示例

```bash
# 如果用户的主题是"dark"，则保留整个preferences对象
cut_json -file rules_example.json -keep-if-value "user.preferences.theme=dark" -pretty
```

### 预期输出

```json
{
  "user": {
    "preferences": {
      "theme": "dark",
      "notifications": true
    }
  }
}
```

## 规则3: 数组元素条件过滤

### 命令示例

```bash
# 保留库存中的电子产品
cut_json -file rules_example.json -keep-array-match "products:category=\"electronics\"" -pretty
```

### 预期输出

```json
{
  "products": [
    {
      "id": 101,
      "name": "Laptop",
      "category": "electronics",
      "price": 1299.99,
      "inStock": true
    },
    {
      "id": 102,
      "name": "Smartphone",
      "category": "electronics",
      "price": 899.99,
      "inStock": false
    }
  ]
}
```

## 组合多个规则

规则可以组合使用，以实现更复杂的裁剪逻辑：

```bash
# 组合多个规则
cut_json -file rules_example.json \
  -path "user.name" \
  -keep-if-value "user.verified=true" \
  -keep-array-match "products:inStock=true" \
  -pretty
```

### 预期输出

```json
{
  "user": {
    "name": "John Doe",
    "verified": true
  },
  "products": [
    {
      "id": 101,
      "name": "Laptop",
      "category": "electronics",
      "price": 1299.99,
      "inStock": true
    },
    {
      "id": 103,
      "name": "Headphones",
      "category": "accessories",
      "price": 199.99,
      "inStock": true
    }
  ]
}
```

## 注意事项

1. 对于规则2和规则3，值比较支持各种JSON类型（字符串、数字、布尔值、null）
2. 字符串值在命令行中需要正确转义
3. 多个规则的结果会合并到一个JSON对象中
4. 如果没有找到匹配的路径或条件，相应的部分将不会出现在结果中