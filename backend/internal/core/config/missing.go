package config

import "slices"

// Missing 返回没填的必填项（配置里的名字，已排序）。empty 的键是配置项的名字，值表示它是否为空。
//
// 可选能力（某个外部服务等）缺了必填项时按“未配置”处理：它的接口明确返回未配置，平台其他功能照常；
// 启动日志写明缺了哪几项（只写名字，不写值）。例：
//
//	func (c PaymentConfig) Missing() []string {
//		return config.Missing(map[string]bool{"api_base": c.APIBase == "", "app_key": c.AppKey == ""})
//	}
//	func (c PaymentConfig) Configured() bool { return len(c.Missing()) == 0 }
func Missing(empty map[string]bool) []string {
	out := []string{}
	for name, isEmpty := range empty {
		if isEmpty {
			out = append(out, name)
		}
	}
	slices.Sort(out)
	return out
}
