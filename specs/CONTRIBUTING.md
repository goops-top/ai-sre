# 规范贡献指南

本文档描述了如何为AI SRE项目贡献API规范、数据模型和协议定义。

##  规范设计原则

### 1. API First 设计
- **规范先行**: 所有API必须先定义OpenAPI规范，再实现代码
- **契约驱动**: 使用规范作为前后端开发的契约
- **版本管理**: 严格的API版本控制和向后兼容性

### 2. 一致性原则
- **命名规范**: 统一的命名约定和术语
- **数据格式**: 标准化的请求/响应格式
- **错误处理**: 统一的错误码和错误响应

### 3. 可扩展性
- **模块化设计**: 规范按功能模块组织
- **插件化架构**: 支持新工具和服务的扩展
- **多语言支持**: 支持多种编程语言的代码生成

##  规范类型

### OpenAPI 3.0 规范
- **用途**: REST API接口定义
- **位置**: `specs/openapi/`
- **命名**: `{service-name}-api.yaml`
- **验证**: Spectral + Swagger CLI

### Protocol Buffers
- **用途**: gRPC服务定义和高性能通信
- **位置**: `specs/proto/`
- **命名**: `{service-name}.proto`
- **验证**: buf lint + buf breaking

### JSON Schema
- **用途**: 数据模型和配置验证
- **位置**: `specs/schemas/`
- **命名**: `{category}/{name}.json`
- **验证**: ajv validator

##  开发工作流

### 1. 规范设计阶段

```mermaid
graph LR
    A[需求分析] --> B[API设计]
    B --> C[编写规范]
    C --> D[规范评审]
    D --> E[规范发布]
```

**步骤说明**:
1. **需求分析**: 明确API功能需求和使用场景
2. **API设计**: 设计RESTful API或gRPC服务接口
3. **编写规范**: 使用标准格式编写规范文件
4. **规范评审**: 提交PR进行代码评审
5. **规范发布**: 合并后自动生成代码和文档

### 2. 规范编写规范

#### OpenAPI规范编写
```yaml
# 基本信息
openapi: 3.0.3
info:
  title: Service Name API
  description: |
    详细的API描述
    包括功能说明和使用场景
  version: 1.0.0
  contact:
    name: AI SRE Team
    email: ai-sre@your-org.com

# 服务器配置
servers:
  - url: http://localhost:8080
    description: 开发环境
  - url: https://api.ai-sre.com
    description: 生产环境

# 安全配置
security:
  - bearerAuth: []

# 路径定义
paths:
  /api/v1/resources:
    get:
      tags:
        - Resources
      summary: 获取资源列表
      description: 获取所有可用资源的列表
      operationId: listResources
      parameters:
        - $ref: '#/components/parameters/PageParam'
        - $ref: '#/components/parameters/LimitParam'
      responses:
        '200':
          description: 成功返回资源列表
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ResourceListResponse'
        '400':
          $ref: '#/components/responses/BadRequest'
        '500':
          $ref: '#/components/responses/InternalServerError'
```

#### Protocol Buffers编写
```protobuf
syntax = "proto3";

package ai_sre.service.v1;

option go_package = "github.com/your-org/ai-sre/specs/proto/service/v1;servicev1";

import "google/protobuf/timestamp.proto";
import "google/protobuf/struct.proto";

// 服务定义
service ResourceService {
  // 获取资源列表
  rpc ListResources(ListResourcesRequest) returns (ListResourcesResponse);
  
  // 获取资源详情
  rpc GetResource(GetResourceRequest) returns (Resource);
}

// 消息定义
message Resource {
  string id = 1;
  string name = 2;
  ResourceType type = 3;
  ResourceStatus status = 4;
  google.protobuf.Timestamp created_at = 5;
  google.protobuf.Timestamp updated_at = 6;
}

enum ResourceType {
  RESOURCE_TYPE_UNSPECIFIED = 0;
  RESOURCE_TYPE_COMPUTE = 1;
  RESOURCE_TYPE_STORAGE = 2;
  RESOURCE_TYPE_NETWORK = 3;
}

enum ResourceStatus {
  RESOURCE_STATUS_UNSPECIFIED = 0;
  RESOURCE_STATUS_ACTIVE = 1;
  RESOURCE_STATUS_INACTIVE = 2;
}
```

### 3. 规范验证

#### 本地验证
```bash
# 验证OpenAPI规范
make validate-openapi

# 验证Protocol Buffers
make validate-proto

# 验证所有规范
make validate
```

#### CI/CD验证
- 每次提交自动运行规范验证
- 规范变更检查向后兼容性
- 自动生成代码和文档

##  编写指南

### OpenAPI编写最佳实践

#### 1. 基本结构
```yaml
# 必需字段
openapi: 3.0.3
info:
  title: 明确的API标题
  description: 详细的功能描述
  version: 语义化版本号
  
# 推荐字段
servers:
  - url: 开发环境URL
  - url: 生产环境URL
  
security:
  - 认证方式: []
```

#### 2. 路径设计
```yaml
paths:
  # 使用复数名词
  /api/v1/users:
    get: # 获取列表
    post: # 创建资源
  
  /api/v1/users/{userId}:
    get: # 获取详情
    put: # 更新资源
    delete: # 删除资源
  
  # 子资源
  /api/v1/users/{userId}/tasks:
    get: # 获取用户任务列表
```

#### 3. 响应设计
```yaml
responses:
  '200':
    description: 成功响应
    content:
      application/json:
        schema:
          allOf:
            - $ref: '#/components/schemas/BaseResponse'
            - type: object
              properties:
                data:
                  $ref: '#/components/schemas/ResourceData'
  
  '400':
    $ref: '#/components/responses/BadRequest'
  
  '500':
    $ref: '#/components/responses/InternalServerError'
```

#### 4. 数据模型
```yaml
components:
  schemas:
    BaseResponse:
      type: object
      required:
        - success
        - timestamp
      properties:
        success:
          type: boolean
        timestamp:
          type: string
          format: date-time
        request_id:
          type: string
    
    Resource:
      type: object
      required:
        - id
        - name
        - type
      properties:
        id:
          type: string
          description: 资源唯一标识符
        name:
          type: string
          minLength: 1
          maxLength: 100
          description: 资源名称
        type:
          $ref: '#/components/schemas/ResourceType'
```

### Protocol Buffers编写最佳实践

#### 1. 文件结构
```protobuf
// 文件头
syntax = "proto3";
package ai_sre.service.v1;
option go_package = "path/to/package";

// 导入
import "google/protobuf/timestamp.proto";
import "common/types.proto";

// 服务定义
service ServiceName {
  rpc MethodName(Request) returns (Response);
}

// 消息定义
message MessageName {
  // 字段定义
}

// 枚举定义
enum EnumName {
  ENUM_NAME_UNSPECIFIED = 0;
  ENUM_NAME_VALUE1 = 1;
}
```

#### 2. 命名规范
```protobuf
// 服务名：PascalCase + Service后缀
service UserService {}

// 方法名：PascalCase，动词开头
rpc CreateUser(CreateUserRequest) returns (User);
rpc GetUser(GetUserRequest) returns (User);
rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);

// 消息名：PascalCase
message User {}
message CreateUserRequest {}

// 字段名：snake_case
message User {
  string user_id = 1;
  string full_name = 2;
  google.protobuf.Timestamp created_at = 3;
}

// 枚举：SCREAMING_SNAKE_CASE，以UNSPECIFIED开头
enum UserStatus {
  USER_STATUS_UNSPECIFIED = 0;
  USER_STATUS_ACTIVE = 1;
  USER_STATUS_INACTIVE = 2;
}
```

#### 3. 字段编号
```protobuf
message User {
  // 1-15: 常用字段（1字节编码）
  string id = 1;
  string name = 2;
  UserStatus status = 3;
  
  // 16-2047: 次要字段（2字节编码）
  string description = 16;
  map<string, string> metadata = 17;
  
  // 预留字段编号
  reserved 4 to 10;
  reserved "old_field_name";
}
```

##  规范验证规则

### OpenAPI验证规则

#### 必需规则
-  所有操作必须有`operationId`
-  所有操作必须有`summary`
-  所有操作必须有`tags`
-  所有路径必须遵循RESTful约定
-  所有响应必须有描述

#### 推荐规则
-  所有操作应该有`description`
-  所有模型应该有`description`
-  列表API应该支持分页
-  所有API应该有错误响应

#### 自定义规则
-  API路径必须以`/api/v{version}/`开头
-  所有API必须有标准错误响应
-  Agent API标题必须包含"Agent"
-  MCP API标题必须包含"MCP"

### Protocol Buffers验证规则

#### 必需规则
-  所有枚举必须有`UNSPECIFIED`值
-  所有字段必须有注释
-  服务方法必须遵循命名约定
-  不能有向后不兼容的变更

#### 推荐规则
-  使用标准类型（timestamp, struct等）
-  合理使用字段编号
-  预留字段编号用于扩展

##  工具使用

### 代码生成
```bash
# 生成所有代码
make generate

# 只生成OpenAPI代码
make generate-openapi

# 只生成Protocol Buffers代码
make generate-proto
```

### 规范验证
```bash
# 验证所有规范
make validate

# 验证OpenAPI规范
spectral lint specs/openapi/*.yaml

# 验证Protocol Buffers
buf lint specs/proto/
```

### 文档生成
```bash
# 生成API文档
make docs-generate

# 启动文档服务器
make docs-serve
```

##  提交检查清单

### 新增API规范
- [ ] 编写完整的OpenAPI规范
- [ ] 添加所有必需的响应状态码
- [ ] 包含详细的参数和响应描述
- [ ] 添加使用示例
- [ ] 通过Spectral验证
- [ ] 生成客户端代码测试

### 新增gRPC服务
- [ ] 编写完整的Protocol Buffers定义
- [ ] 包含详细的服务和消息注释
- [ ] 遵循命名约定
- [ ] 通过buf验证
- [ ] 检查向后兼容性
- [ ] 生成服务端和客户端代码

### 规范变更
- [ ] 检查向后兼容性
- [ ] 更新版本号
- [ ] 添加变更日志
- [ ] 更新相关文档
- [ ] 通知相关团队

##  评审流程

### 规范评审要点
1. **功能完整性**: 规范是否完整覆盖需求
2. **一致性**: 是否遵循项目约定
3. **可用性**: 是否易于理解和使用
4. **扩展性**: 是否支持未来扩展
5. **性能**: 是否考虑性能影响

### 评审角色
- **规范作者**: 编写和维护规范
- **架构师**: 审查架构一致性
- **开发者**: 审查实现可行性
- **测试工程师**: 审查测试覆盖度

##  参考资源

### 官方文档
- [OpenAPI 3.0 规范](https://swagger.io/specification/)
- [Protocol Buffers 指南](https://developers.google.com/protocol-buffers)
- [JSON Schema 规范](https://json-schema.org/)

### 工具文档
- [Spectral 规则](https://meta.stoplight.io/docs/spectral/)
- [buf 配置](https://docs.buf.build/configuration/overview)
- [OpenAPI Generator](https://openapi-generator.tech/)

### 最佳实践
- [RESTful API 设计指南](https://restfulapi.net/)
- [gRPC 最佳实践](https://grpc.io/docs/guides/best-practices/)
- [API 版本管理](https://blog.stoplight.io/api-versioning)