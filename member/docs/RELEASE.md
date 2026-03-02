# 0.6.0

## 新增功能

### 1. 新增 hooks 收集 API 端点

- 新增 `POST /collect/{tool_name}` API 端点，用于收集 hooks 事件
- 用于收集 hooks 事件的工具名称，目前支持 `claude`, `opencode`.
- 请求体为 JSON 格式，包含 `session_id`, `hook_event_name`, `data` 字段。

## 修复问题

### 1. 去掉 Codex 支持

- 去掉 Codex 支持，只保留 Claude 和 OpenCode 支持。并选择 Open Code 作为首选工具


### 2. 修复 Claude 问题

- 在 Toggle On 按钮点击后，不会自动合并已有配置的问题。 -- 已解决

### 3. 重名供应商的问题

- 在前端限制增加已有供应商名字，造成数据混乱的问题。 -- 已解决
- 并且优化前端的输入框，增加只读提示，供应商名字一旦提交，无法修改。

## 兼容性

兼容 0.5.2 版本服务器，无需升级