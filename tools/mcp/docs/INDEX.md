# AI SRE MCP Server 文档索引

##  文档概览

本目录包含AI SRE MCP Server的完整文档集合，按照不同用户需求和使用场景组织。

##  快速导航

### 新用户入门
1. **[README.md](../README.md)** - 项目概述和快速开始
2. **[快速参考](QUICK_REFERENCE.md)** - 常用命令和配置速查
3. **[使用示例](../examples/usage-examples.md)** - 实际使用场景和代码示例

### 详细使用指南
4. **[用户指南](USER_GUIDE.md)** - 完整的使用说明和最佳实践
5. **[API参考](API_REFERENCE.md)** - 详细的API接口文档

## 📖 文档结构

```
tools/mcp/
├── README.md                    # 项目主页，快速开始
├── docs/
│   ├── INDEX.md                # 本文档索引
│   ├── USER_GUIDE.md           # 完整用户指南
│   ├── API_REFERENCE.md        # API接口参考
│   └── QUICK_REFERENCE.md      # 快速参考卡片
└── examples/
    ├── usage-examples.md       # 使用示例集合
    ├── test-all-modes.sh       # 完整功能测试
    ├── test-startup.sh         # 启动测试
    └── test-server.sh          # 服务器测试
```

##  按需求选择文档

### 我想要...

####  快速上手
**推荐阅读顺序**:
1. [README.md](../README.md) - 了解项目和基本使用
2. [快速参考](QUICK_REFERENCE.md) - 查看常用命令
3. [使用示例](../examples/usage-examples.md) - 参考实际案例

####  深入配置
**推荐阅读顺序**:
1. [用户指南](USER_GUIDE.md) - 完整配置说明
2. [API参考](API_REFERENCE.md) - 接口详细信息
3. [快速参考](QUICK_REFERENCE.md) - 作为速查手册

####  开发集成
**推荐阅读顺序**:
1. [API参考](API_REFERENCE.md) - 了解接口规范
2. [用户指南](USER_GUIDE.md) - 理解认证和安全
3. [使用示例](../examples/usage-examples.md) - 参考集成代码

#### 故障排除
**推荐阅读顺序**:
1. [用户指南 - 故障排除](USER_GUIDE.md#故障排除)
2. [快速参考 - 常见错误](QUICK_REFERENCE.md#常见错误)
3. [API参考 - 错误处理](API_REFERENCE.md#-错误处理)

##  文档详细说明

### [README.md](../README.md)
- **目标用户**: 所有用户
- **内容**: 项目概述、主要特性、快速开始、基本使用
- **长度**: 中等（~200行）
- **更新频率**: 随版本更新

### [用户指南](USER_GUIDE.md)
- **目标用户**: 系统管理员、运维工程师
- **内容**: 完整的安装、配置、使用、故障排除指南
- **长度**: 详细（~800行）
- **更新频率**: 功能变更时更新

### [API参考](API_REFERENCE.md)
- **目标用户**: 开发者、集成工程师
- **内容**: 详细的API接口、参数、响应格式、错误码
- **长度**: 详细（~600行）
- **更新频率**: API变更时更新

### [快速参考](QUICK_REFERENCE.md)
- **目标用户**: 有经验的用户
- **内容**: 命令速查、配置模板、常见问题
- **长度**: 简洁（~150行）
- **更新频率**: 定期同步更新

### [使用示例](../examples/usage-examples.md)
- **目标用户**: 所有用户
- **内容**: 实际使用场景、代码示例、最佳实践
- **长度**: 中等（~400行）
- **更新频率**: 新功能添加时更新

##  文档版本管理

### 版本策略
- **主版本更新**: 重大功能变更时，所有文档同步更新
- **次版本更新**: 新功能添加时，相关文档更新
- **补丁版本**: 错误修复时，相关文档修正

### 版本标识
每个文档底部包含版本信息：
```
版本: v1.0.0
最后更新: 2026-02-12
```

##  文档反馈

### 如何贡献
1. **发现错误**: 提交Issue描述问题
2. **改进建议**: 提交PR或Issue说明改进点
3. **新增内容**: 提交PR添加新的使用示例或说明

### 文档质量标准
-  准确性：所有示例都经过测试验证
-  完整性：覆盖所有主要功能和使用场景
-  易读性：清晰的结构和简洁的语言
-  实用性：提供可直接使用的示例和模板

##  外部资源

### 相关技术文档
- [Model Context Protocol 规范](https://modelcontextprotocol.io/)
- [Go语言官方文档](https://golang.org/doc/)
- [Docker官方文档](https://docs.docker.com/)

### 社区资源
- [项目GitHub仓库](https://github.com/your-org/ai-sre)
- [问题追踪](https://github.com/your-org/ai-sre/issues)
- [讨论区](https://github.com/your-org/ai-sre/discussions)

##  文档使用统计

### 推荐阅读路径
1. **初学者**: README → 快速参考 → 使用示例
2. **管理员**: 用户指南 → API参考 → 快速参考
3. **开发者**: API参考 → 使用示例 → 用户指南

### 常见查询
- 如何启动服务器？ → [README.md](../README.md#快速开始)
- 如何配置认证？ → [用户指南](USER_GUIDE.md#认证系统)
- API接口格式？ → [API参考](API_REFERENCE.md#http管理接口)
- 故障排除？ → [用户指南](USER_GUIDE.md#故障排除)

---

**文档维护**: AI SRE Team  
**最后更新**: 2026-02-12  
**文档版本**: v1.0.0