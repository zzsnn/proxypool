Clash客户端支持：
- Clash for Windows（需要Clash Core1.3以上）
- ClashX（需要Clash Core1.3以上）
- 不支持ClashXR与ClashR等非原生Clash Core客户端。

TODO
- [ ] set http context time(dev)

## New

2020-10-30
- 减少启动时的内存占用（使用release版本第一次运行时除外）

2020-10-26
- 单独分离出healthcheck模块
- 分离出用于本地检测proxypool可用性的部分，见[proxypoolCheck](https://github.com/Sansui233/proxypoolCheck)项目

2020-10-24
- Vmess动态格式解析，对链接的字段类型进行强制转换，提高一点点抓取数量。

2020-10-23
- 修复数据库未连接时的err提示
- 忽略vmess的Unmarshal时的ps类型错误

2020-10-21
- 数据库更新改为保留数据库已有节点与当次有效节点，且清扫失效时间大于7天的节点
- Manually sync to original source v0.3.10

2020-10-10
- 修复：对空provider添加NULL节点，防止Clash报错
- 数据库更新不再存储所有的节点，只保留当次有效节点

2020-10-09
- 增加本地http运行用的配置文件  
    > clash的本地配置文件位于127.0.0.1:8080/clash/localconfig