# Go Vue Admin 开发文档

本项目是基于 **Goravel v1.18 + Vue 3 + shadcn-vue** 的通用后台管理平台。项目只建设后台平台能力，不重复实现 Goravel 已经提供的 Web 框架能力。

## 阅读顺序

- [architecture.md](./architecture.md)：总体架构和依赖方向
- [goravel-boundaries.md](./goravel-boundaries.md)：Goravel 能力边界
- [module-system.md](./module-system.md)：模块组织和注册
- [resource-engine.md](./resource-engine.md)：标准 CRUD 资源
- [frontend.md](./frontend.md)：Vue Admin 和 shadcn-vue 规范
- [i18n.md](./i18n.md)：中文、英文和模块化语言包
- [openapi.md](./openapi.md)：前后端 API 契约
- [generator.md](./generator.md)：代码生成器
- [plugin-system.md](./plugin-system.md)：插件系统设计预留
- [security.md](./security.md)：认证、授权和敏感数据
- [authentication.md](./authentication.md)：JWT 和 Refresh Token 认证
- [integrations.md](./integrations.md)：Goravel 驱动和扩展包集成
- [testing.md](./testing.md)：测试和验收
- [roadmap.md](./roadmap.md)：分阶段路线图
- [fastimg-product-design.md](./fastimg-product-design.md)：FastImg 图床产品设计、免费/付费模型和真实模块边界
- [superpowers/plans/2026-09-23-fastimg-repository-development.md](./superpowers/plans/2026-09-23-fastimg-repository-development.md)：基于本仓库结构的 FastImg 开发计划
- [fastimg-requirements-matrix.md](./fastimg-requirements-matrix.md)：FastImg 需求编号、API 目录、页面清单和风险控制
- [fastimg-data-contracts.md](./fastimg-data-contracts.md)：FastImg 数据字段、状态机、幂等和一致性契约
- [fastimg-release-runbook.md](./fastimg-release-runbook.md)：本地开发、迁移、发布、回滚和运行检查
- [fastimg-deployment.md](./fastimg-deployment.md)：Linux 源码部署、Docker 应用部署和外部数据库/Redis约束
- [fastimg-developer-api.md](./fastimg-developer-api.md)：Personal API Token、上传接口和各种图片链接返回规范
- [fastimg-frontend-design.md](./fastimg-frontend-design.md)：用户端和管理端页面、状态与交互设计
- [fastimg-moderation-policy.md](./fastimg-moderation-policy.md)：违规图片、违规账户、举报、处罚与申诉策略
- [fastimg-compliance-policy.md](./fastimg-compliance-policy.md)：服务条款、版权、隐私、数据保留和安全事件规则
- [fastimg-storage-cost-policy.md](./fastimg-storage-cost-policy.md)：存储、流量、回收站、CDN 和免费服务成本策略
- [fastimg-hotlink-protection.md](./fastimg-hotlink-protection.md)：Referer 白名单、签名 URL、CDN 和防盗链策略
- [fastimg-payment-gateway.md](./fastimg-payment-gateway.md)：支付网关、订单、支付流水、退款、订阅和权益履约
- [fastimg-production-readiness.md](./fastimg-production-readiness.md)：生产开发就绪评估、阶段门禁和 Provider 验证状态
- [fastimg-production-audit.md](./fastimg-production-audit.md)：生产配置、支付幂等、审计脱敏和运维清理门禁
- [fastimg-production-gates.md](./fastimg-production-gates.md)：真实支付、TLS、备份恢复、漏洞扫描、对象存储和压测门禁
- [fastimg-stage-development-plan.md](./fastimg-stage-development-plan.md)：M0-M8 阶段性开发计划和详细功能清单

## 核心原则

```text
Goravel 有的能力不重复实现
标准 CRUD 优先使用 Resource
复杂业务使用普通 Vue 页面和 Goravel Service
前端直接使用 shadcn-vue 官方组件
生成代码与人工代码分离
权限必须由后端最终裁决
```

第一阶段只要求：Goravel v1.18 基础工程、后台壳层、JWT 认证、RBAC、标准 Resource CRUD、OpenAPI、TypeScript Client、PostgreSQL、Redis 和基础审计。插件运行时、AI 生成、全局搜索、复杂 Dashboard、导入导出和数据库反向工程属于后续能力。
