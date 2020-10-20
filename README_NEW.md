Clash客户端支持：
- Clash for Windows（需要Clash Core1.3以上）
- ClashX（需要Clash Core1.3以上）
- 不支持ClashXR与ClashR等非原生Clash Core客户端。  

## New

2020-10-21
- 数据库更新改为保留所有节点，且会清扫失效时间大于7天的节点

2020-10-10
- 修复：对空provider添加NULL节点，防止Clash报错
- 数据库更新不再存储所有的节点，只保留当次有效节点

2020-10-09
- 增加本地http运行用的配置文件  
    > clash的本地配置文件位于127.0.0.1:8080/clash/localconfig