# 配置约定

后端在启动时读取一个 YAML 文件（`-f` 指定，默认 `configs/config.yaml`），之后不再变化：改了配置要重启。

1. **部署相关的值都写在部署配置里，代码里不留默认值。** 地址、账号、密钥、租户、用户组之类的值因部署而异，
   写在代码里会让配置文件看不出实际用了什么，改一次还要发版。示例配置（`configs/config*.yaml`）里给空值或
   `CHANGE_ME_…` 占位。
2. **凭据只放在不进 git、不进镜像的文件里。** 部署专用的配置命名为 `backend/configs/*.local.yaml`（已被
   `.gitignore` 与 `backend/.dockerignore` 忽略），或由部署脚本生成到私有目录（如 `deploy/local/runtime/`），
   运行时挂载进容器。含凭据的字段加 `json:"-"`，不出现在状态接口和日志里。
3. **平台必需的配置缺了，启动失败。** JWT、数据库、Redis 等在 `config.ValidateServer` 里校验，错误信息写明是哪一项。
4. **可选能力缺了必填项，只是这个能力“未配置”。** 它的接口明确返回未配置（不伪造数据），平台其他功能照常；
   启动日志写明缺了哪几项（只写名字，不写值）。用 `config.Missing` 写这类判断：

   ```go
   func (c PaymentConfig) Missing() []string {
   	return config.Missing(map[string]bool{"api_base": c.APIBase == "", "app_key": c.AppKey == ""})
   }
   func (c PaymentConfig) Configured() bool { return len(c.Missing()) == 0 }
   ```

5. **不用的配置项删掉。** 示例配置里每一项都应该有代码在用；保留没人读的配置会让人以为它有效。
