由于前端页面的显示固定了使用https，但proxypool没有添加https中间件的支持只支持http。推荐使用nginx的reverse proxy支持域名的https。以下以Ubuntu为例说明简单的步骤。

需要准备
- 一个申请好并添加了DNS解析的的域名(freenom可以申请免费域名并解析)
- 一台具有管理员权限的VPS
- 具有使用ssh登录VPS与基本命令行操作的技能

如果以上超纲了，请使用README中的heroku部署。

此教程nginx之后部分适用于没有申请ssl证书的域名。如果自己已经有证书，需要自行配置nginx的ssl。

# 1 部署到VPS

登录到VPS后，下载编译好的版本。

```shell
wget https://github.com/Sansui233/proxypool/releases/download/v0.6.1/proxypool-linux-amd64-v0.6.1.gz # 下载
gzip -d proxypool-linux-amd64-v0.6.1.gz # 解压
mv proxypool-linux-amd64-v0.6.1 proxypool #重命名
chmod 755 proxypool # 赋予执行权限
```

再自行下载配置文件(config.yaml与source.yaml)，放在与proxypool相同目录下。
```shell
wget https://raw.githubusercontent.com/Sansui233/proxypool/master/config/config.yaml
wget https://raw.githubusercontent.com/Sansui233/proxypool/master/config/source.yaml
```

在config.yaml中的`port`字段设置运行的端口，留空为`12580`。`source`字段更改为`./source.yaml`。**所有字段均可自行按需更改**。

> 注意，如果你的环境变量中有PORT字段，会以环境变量优先。  
> 如果端口和已经运行的程序冲突请自行更改端口，或kill占用了端口的进程。

后台运行

```shell
nohup ./proxypool -c config.yaml &
```

检查前端是否正常工作

```shell
curl http://127.0.0.1:12580
```

# 2 配置nginx

下载并安装nginx。

```shell
sudo apt-get install nginx
```

修改nginx的配置中的server部分，设置reverse proxy到proxypool服务的端口。默认配置文件路径可以使用nginx -h查看，通常入口配置文件中会包含了sites-enable文件夹下的配置。具体情况请以自己的机器为准。

```
# vim /etc/nginx/sites-enable/default
server {
	listen 80; # 需要设置为80，稍后certbot验证使用
	server_name proxypoolss.tk; # 你的域名

	location / {
		proxy_pass http://127.0.0.1:12580/; # proxypool服务的地址
	}
}
```

启动nginx

```shell
nginx
```

查看是否正常工作

```shell
curl http://127.0.0.1:80
```

# 3 使用Certbot自动颁发安装证书

如果遇到问题，以及详细步骤请根据[Certbot官网](https://certbot.eff.org/)操作。不同平台有差异。

安装snapd。Ubuntu20.4（以及其他某些版本）已经预装了snapd，可以跳过。详细见[Certbot官网](https://certbot.eff.org/)。

安装certbot

``` shell
sudo snap install --classic certbot # 安装
sudo ln -s /snap/bin/certbot /usr/bin/certbot # 程序加入PATH
```

为nginx安装证书。需要根据引导自行完成

```shell
sudo certbot --nginx
```

重启nginx
```shell
nginx -s reload
```

查看是否正常工作
```shell
curl https://127.0.0.1:443
```

完成，可以使用https访问了。
