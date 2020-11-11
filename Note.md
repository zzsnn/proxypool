# Note 

由于个人并未学过go，看代码需要做下笔记。仅为个人参考。

## 处理订阅源：getter类
有关订阅源的package位于pkg/getter。

订阅源的类型为接口Getter，实现Getter至少需要实现Get()和Get2chan()。 
- Get() 返回一个ProxyList
- Get2chan() Send proxy to Channel用于并发抓取

已实现的Getter（以sourceType命名）
- subscribe（该实现比接口Getter多了个url）
- tgchannel
- web_fanqiangdang
- web_fuzz
- web_fuzz_sub

接口Getter与err状态组成一个creator，方便错误处理。
为了方便外部程序辨认creator类型，在init()中初始化一个map，key为sourceType字符串，value为creator。

程序运行时，package app由配置文件读取到source.yaml，由sourceType map到对应的creator类型，同时使用sourceOption(通常是url)初始化一个creator。

所有Getter最后存于package app的Getters中。

## proxy类
节点的接口为interface proxy，由struct Base实现其基类，Vmess等实现多态。

所有字段名依据clash的配置文件标准设计。比如
```
type ShadowsocksR struct {
	Base          // 节点基本信息
	Password      string `yaml:"password" json:"password"`
	Cipher        string `yaml:"cipher" json:"cipher"`
	Protocol      string `yaml:"protocol" json:"protocol"`
	ProtocolParam string `yaml:"protocol-param,omitempty" json:"protocol_param,omitempty"`
	Obfs          string `yaml:"obfs" json:"obfs"`
	ObfsParam     string `yaml:"obfs-param,omitempty" json:"obfs_param,omitempty"`
	Group         string `yaml:"group,omitempty" json:"group,omitempty"`
}
```

Proxylist是proxy数组加上一系列批量处理proxy的方法。

## 抓取
task.go的Crawl.go实现抓取。

1. 并发抓取订阅源，加载历史节点
2. 节点去重，去除Clash不支持的类型，重命名
3. 存储所有节点（包括不可用节点）到database和cache
4. 检测IP可用性  
  尽管已经对IP的有效性测试过，但并不保证节点在客户端上可用，因为可能其他参数有误。
5. 存储可用的节点到cache

## 存储
所有节点存储到cache中。

cache中的key设计有：
- allproxies: 所有节点（包括不可用节点）
- proxies: 可用节点
- clashproxies: clash支持的节点。第一次运行时是把proxies复制过来的。

问题是对于失效的节点也存储，运行时间久了无用的cache会非常多。可以考虑删除对失效节点的存放。

### 使用数据库
远程运行时添加add即可，heroku自己会添加DATABASE_URL环境变量到provision，无需其他配置。

本地运行时安装postgresql，建立相应user和database。

```
dsn := "user=proxypool password=proxypool dbname=proxypool port=5432 sslmode=disable TimeZone=Asia/Shanghai"
```

程序运行时建立会proxies表。每次运行时读出节点，爬虫完成后再存储可用的节点进去。

更新时会Update所有数据库上次可用节点的usable为false（此时useable全是false），然后存储新节点，已有的条目则Update usable。最后再自动清除7天未更新且不可用的节点。

重点在于，失效的条目不能更新。

## Health Check

分为延迟测试与测速（带宽测试）。

延迟测试用于筛选掉无效的节点。

测速的质量取决于带宽，目前来讲十分不靠谱仅供参考。为了测速不占用太多时间尽可能做了并发，单个节点最多可能同时发出24（8*3）个tcp连接。如果看到许多节点速度相同，大概率是带宽限制，实际速度在这之上。我测试时20协程并发卡的是12.8，17.7，25.6，51.2这四个数。

如果想完全准确测速，需要在配置里减少并发量(最好降到1)，或提升服务器带宽，源码里调workload至目前值的两倍以上。本机网络环境的20个go routine的测试速度为100个节点/分钟。heroku似乎更快一些。

测速的实现大概依据speedtestgo项目的思路。使用speedtest，延迟测试3次，同时测试三个服务器。依据热身小测试加大负荷。因此即使无节点并发，想到测出100Mbps本机也至少需要300Mbps的带宽。正确的做法本应该是单个节点无并发，对于proxy端而言更加准确。但这样实在太慢了。后续可能会考虑只测一个服务器（同speedtest官方），但依然有其他协程在竞争。

2020-11-11：现在觉得没必要了，测并发下的速度，对于实际的应用上似乎比全速的带宽更有意义。至于为什么卡那几个数，应该是warmup后对workload的分级决定的，同时也和并发量有关。总之，大部分情况下，卡到阈值了的节点实际速度只多不少吧。

另外目前参数对于极其缓慢的节点也会测速也会失败。因为warmup的size过小，无法完全利用TCP的窗口，导致了高速和低速节点拉不开差距，所以即便是极其低速的节点也会进行加大负荷的测试，最后导致超过时间限制而测不出来。WarmUp的参数可以再调整一下（又要拖慢测速过程了）

对于单个节点的负荷调参是比较难的，并发量小了跑不满节点带宽，大了容易相互挤网速。如果减少并发加大单次请求的体积？测试时间变长。如果加大并发减小单次的体积？通常启动时的传输速度会慢，测出来的速度低于实际值。如果选择1.125M的大小的测试结果将导致没有一个能超过10Mbps。参考项目的参数已经算是目前最优。

## Web界面

为了方便打包，原作者将静态的assets文件模板由zip压缩后存为字符串的形式，如

```
var _assetsHtmlSurgeHtml="[]byte("\x1f\x8b\x...")"
```

以上字节解压后是一个go的HTML模板。解压时，由gzip的reader写入byte.Buffer，再转换为Bytes写入相应文件。

静态文件打包工具见：[这里](https://github.com/go-bindata/go-bindata) 或 [这里](https://github.com/shuLhan/go-bindata) 。请在修改后html文件后执行docs里的shell脚本

根据原作者要求，请勿修改原作者版权信息。

Web数据根据cache更新，而非Get一次更新一次

## 本地测试

需要注意：
- 修改config的domain
- 修改source，注释掉较慢的源

增加了对config-local文件的解析。url为/clash/localconfig

## Github Action Release和源码自行make版本的区别

Release版本在make之前还打包了所有的静态文件，其中包括了一个60M的GeoIP数据库。  
缺点是打包体积相对较大，因为geoip.go经过gobindata打包后的文件是120。且第一次部署运行时会占用较多内存。  
优点是不需要额外自行下载数据库。打包的体积也比自己下载数据库的体积小太多。

自行源码make版本在不修改bindata/GeoIP的情况下，不包含数据库，打包程序较小。缺点是要自已下载数据库（源码都有了这不算是问题）