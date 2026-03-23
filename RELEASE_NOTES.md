# 更新日志

## Version 0.3.0 (2026-03-23)

### 🧪 测试基础设施
- ✨ 新增 BDD (行为驱动开发) 测试框架，基于 Godog
- 📊 改进测试覆盖率和报告机制
- 🔧 重构测试基础设施，提升可维护性

### 📖 API 文档
- 🌐 新增 OpenAPI 3.0.3 规范文档
- 🔍 集成 Swagger UI 用于 API 探索和测试
- 📝 完善发布流程文档 (RELEASE.md)

### 📦 依赖更新
- 🔄 pgx/v5: 5.8.0 → 5.9.1 (server, integration)
- 🔄 oapi-codegen/runtime: 1.2.0 → 1.3.0 (member, integration)
- 🔄 modernc.org/sqlite: 1.46.1 → 1.47.0 (member)
- 🔄 vue-i18n: 11.1.12 → 11.3.0
- 🔄 vue-tsc: 3.1.3 → 3.2.5
- 🔄 @mui/x-date-pickers: 7.29.4 → 8.27.2
- 🔄 @tailwindcss/postcss: 4.1.17 → 4.2.1
- 🔄 react/react-dom: 小版本更新
- 🔄 axios, dayjs, eslint-plugin-react-refresh 等其他依赖

### 🐛 Bug 修复
- 🛠️ 修复 Reports 组件中的代码质量问题 (Copilot Autofix)

### 📚 文档改进
- 📝 Server Relay 文档更加工具无关，支持任意 AI 工具
- 🔧 修正 release workflow 路径以适配 monorepo 结构

---

## Version 0.2.0

### RBAC 权限系统
- 🔐 基于 Casbin 实现基于角色的访问控制 (RBAC)
- 👥 新增双层权限系统：**member** (只读) 和 **manager** (读写)
- 🛡️ 所有 API 端点实施服务端权限控制
- 🔑 基于 JWT 的身份认证，支持主动令牌刷新

### 测试与质量
- ✅ 使用 testify/mock 实现全面的单元测试
- 🧪 基于接口的依赖注入，提升可测试性
- 📊 处理器和中间件采用表驱动测试模式
- 🔍 所有 HTTP 处理器使用模拟测试

### 架构改进
- 🏗️ 基于 PostgreSQL 的多租户数据隔离
- 🔗 Member 客户端完整的服务器集成
- 🔄 自动从服务器同步配置
- 📊 使用量跟踪和统计聚合

### 开发体验
- 📝 增强文档 (每个模块的 CLAUDE.md)
- 🔧 统一的 Makefile 构建所有组件
- 🚀 Member 和 Server 都支持热重载
- 🧪 改进的集成测试

### API 增强
- 🌐 RESTful API 覆盖全面的端点
- 📋 团队级别的供应商管理
- 🔐 基于角色访问的受保护端点
- 📊 使用量统计和分析端点

详细权限说明请参考 [server/README.md](server/README.md)。

---

## Version 0.1.9

### 供应商管理
- 🎯 供应商级别同步和排序
- 🔄 改进的供应商中继故障转移逻辑
- 🔧 新增 OpenCode 中继支持
- 🗺️ 模型映射持久化修复

### 配置管理
- 📝 增强的配置导入/导出
- 🔍 更好的配置验证
- 🛠️ 改进的设置界面/体验

### Bug 修复
- 🐛 修复供应商同步的数据结构不匹配
- 🔧 解决模型映射持久化问题
- 🚀 改进缓存令牌处理

---

# Code Switch v0.1.8

## 更新亮点
- 🚀 **供应商路由更稳**：移除 Level 分组后按列表顺序重试，日志与失败提示更清晰，成功率统计也不再误将失败计为成功。
- ⚙️ **开机自启动**：新增“开机自启动”开关，安装后可一键配置后台随系统启动。
- 📥 **cc-switch 导入优化**：即使未检测到默认配置也可直接导入，旁路上传提示更友好。
- 🛠️ **构建与 CI 提升**：完善缺失的模型编辑组件，补齐 Node/NSIS 配置并新增自动发布工作流，减少跨平台构建故障。
