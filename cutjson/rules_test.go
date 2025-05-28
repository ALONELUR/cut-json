package cutjson

import (
	"encoding/json"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRules(t *testing.T) {
	// 读取测试数据
	jsonData, err := os.ReadFile("../examples/rules_example.json")
	if err != nil {
		t.Fatalf("无法读取测试文件: %v", err)
	}

	Convey("测试JSON裁剪规则", t, func() {
		Convey("规则1: 保留指定路径", func() {
			rules := []Rule{
				NewKeepPathRule("user.name"),
				NewKeepPathRule("user.age"),
			}

			result, err := CutWithRules(jsonData, rules)

			So(err, ShouldBeNil)

			// 解析结果进行验证
			var resultObj map[string]interface{}
			err = json.Unmarshal(result, &resultObj)
			So(err, ShouldBeNil)

			// 验证结果包含预期的字段
			So(resultObj, ShouldContainKey, "user")
			userObj := resultObj["user"].(map[string]interface{})
			So(userObj, ShouldContainKey, "name")
			So(userObj, ShouldContainKey, "age")
			So(userObj["name"], ShouldEqual, "John Doe")
			So(userObj["age"], ShouldEqual, float64(30))

			// 验证结果不包含其他字段
			So(userObj, ShouldNotContainKey, "verified")
			So(userObj, ShouldNotContainKey, "address")
		})

		Convey("规则2: 如果值匹配，保留父路径", func() {
			rules := []Rule{
				NewKeepParentIfValueMatchesRule("user.preferences.theme", "dark"),
			}

			result, err := CutWithRules(jsonData, rules)

			So(err, ShouldBeNil)

			// 解析结果进行验证
			var resultObj map[string]interface{}
			err = json.Unmarshal(result, &resultObj)
			So(err, ShouldBeNil)

			// 验证结果包含预期的字段
			So(resultObj, ShouldContainKey, "user")
			userObj := resultObj["user"].(map[string]interface{})
			So(userObj, ShouldContainKey, "preferences")

			prefsObj := userObj["preferences"].(map[string]interface{})
			So(prefsObj, ShouldContainKey, "theme")
			So(prefsObj, ShouldContainKey, "notifications")
			So(prefsObj["theme"], ShouldEqual, "dark")
			So(prefsObj["notifications"], ShouldEqual, true)

			// 验证结果不包含其他字段
			So(userObj, ShouldNotContainKey, "name")
			So(userObj, ShouldNotContainKey, "age")
		})

		Convey("规则3: 保留数组中满足条件的元素", func() {
			rules := []Rule{
				NewKeepArrayElementsIfChildValueMatchesRule("products", "category", "electronics"),
			}

			result, err := CutWithRules(jsonData, rules)

			So(err, ShouldBeNil)

			// 解析结果进行验证
			var resultObj map[string]interface{}
			err = json.Unmarshal(result, &resultObj)
			So(err, ShouldBeNil)

			// 验证结果包含预期的字段
			So(resultObj, ShouldContainKey, "products")
			products := resultObj["products"].([]interface{})
			So(len(products), ShouldEqual, 2) // 应该只有两个电子产品

			// 验证第一个产品
			product1 := products[0].(map[string]interface{})
			So(product1["id"], ShouldEqual, float64(101))
			So(product1["name"], ShouldEqual, "Laptop")
			So(product1["category"], ShouldEqual, "electronics")

			// 验证第二个产品
			product2 := products[1].(map[string]interface{})
			So(product2["id"], ShouldEqual, float64(102))
			So(product2["name"], ShouldEqual, "Smartphone")
			So(product2["category"], ShouldEqual, "electronics")
		})

		Convey("组合多个规则", func() {
			rules := []Rule{
				NewKeepPathRule("user.name"),
				NewKeepParentIfValueMatchesRule("user.verified", true),
				NewKeepArrayElementsIfChildValueMatchesRule("products", "inStock", true),
			}

			result, err := CutWithRules(jsonData, rules)

			So(err, ShouldBeNil)

			// 解析结果进行验证
			var resultObj map[string]interface{}
			err = json.Unmarshal(result, &resultObj)
			So(err, ShouldBeNil)

			// 验证用户部分
			So(resultObj, ShouldContainKey, "user")
			userObj := resultObj["user"].(map[string]interface{})
			So(userObj, ShouldContainKey, "name")
			So(userObj, ShouldContainKey, "verified")
			So(userObj["name"], ShouldEqual, "John Doe")
			So(userObj["verified"], ShouldEqual, true)

			// 验证产品部分
			So(resultObj, ShouldContainKey, "products")
			products := resultObj["products"].([]interface{})
			So(len(products), ShouldEqual, 2) // 应该有两个库存中的产品

			// 验证第一个产品
			product1 := products[0].(map[string]interface{})
			So(product1["id"], ShouldEqual, float64(101))
			So(product1["inStock"], ShouldEqual, true)

			// 验证第二个产品
			product2 := products[1].(map[string]interface{})
			So(product2["id"], ShouldEqual, float64(103))
			So(product2["inStock"], ShouldEqual, true)
		})
	})
}
