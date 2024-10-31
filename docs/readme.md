## swagger 格式

https://github.com/swaggo/swaggo.io/tree/master/declarative_comments_format

## 文档使用步骤

1. 按照swagger要求给接口代码添加声明式注释，相关注释需要main入口处、controller每个接口入口处
2. 使用swag工具扫描代码自动生成API接口文档数据 swag init
3. 在内部http开启文档服务，app.json增加"Docs": "true"，同时server导入 _ "xxx/docs"