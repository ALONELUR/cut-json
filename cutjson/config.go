package cutjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// RuleConfig 表示JSON配置文件中的单个规则配置
type RuleConfig struct {
	Type      string      `json:"type"`
	Where     string      `json:"where"`
	ChildPath string      `json:"child_path,omitempty"`
	Op        string      `json:"op,omitempty"`
	Value     interface{} `json:"value,omitempty"`
}

// RulesConfig 表示整个JSON配置文件的结构
type RulesConfig struct {
	Rules []RuleConfig `json:"rules"`
}

// LoadRulesFromConfig 从JSON配置文件加载规则
func LoadRulesFromConfig(configPath string) ([]Rule, error) {
	// 读取配置文件
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("无法读取配置文件: %w", err)
	}

	// 解析配置文件
	var config RulesConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("无法解析配置文件: %w", err)
	}

	// 构建规则列表
	rules := make([]Rule, 0, len(config.Rules))
	for _, ruleConfig := range config.Rules {
		rule, err := buildRuleFromConfig(ruleConfig)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

// buildRuleFromConfig 根据配置构建规则
func buildRuleFromConfig(config RuleConfig) (Rule, error) {
	switch config.Type {
	case "keep_path":
		return NewKeepPathRule(config.Where), nil

	case "keep_parent_if_value_matches":
		if config.Where == "" {
			return Rule{}, errors.New("keep_parent_if_value_matches规则必须指定where字段")
		}
		if config.Op != "equals" {
			return Rule{}, errors.New("keep_parent_if_value_matches规则目前只支持equals操作符")
		}
		return NewKeepParentIfValueMatchesRule(config.Where, config.Value), nil

	case "keep_array_elements_if_child_value_matches":
		if config.Where == "" {
			return Rule{}, errors.New("keep_array_elements_if_child_value_matches规则必须指定where字段")
		}
		if config.ChildPath == "" {
			return Rule{}, errors.New("keep_array_elements_if_child_value_matches规则必须指定child_path字段")
		}
		if config.Op != "equals" {
			return Rule{}, errors.New("keep_array_elements_if_child_value_matches规则目前只支持equals操作符")
		}
		return NewKeepArrayElementsIfChildValueMatchesRule(config.Where, config.ChildPath, config.Value), nil

	default:
		return Rule{}, fmt.Errorf("未知的规则类型: %s", config.Type)
	}
}
