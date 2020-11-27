
# port-forward
Go语言开发的端口转发工具  for port data forward

```
开发语言：GO
控制台框架：beego
数据库：sqlite3
```

# 最近更新
```
v1.3.2 发布，服务稳定性已经过长时间的验证，推荐使用该版本！
v1.3.1 发布，增加程序启动自动开启转发
v1.2.9_beta 发布，完善点对点转发的稳定性，支持TCP和UDP协议转发
```

# 最新编译好的版本下载：
> https://gitee.com/tavenli/port-forward/releases/v1.3.2


# 功能介绍

> 支持Web控制台添加端口映射

> 支持对每条端口映射进行开启和关闭控制

> 支持 RestfulAPI 接口，方便被其它系统集成

> 支持每条端口转发的同时，再分发给多个端口，满足某些测试场景

# 已编译好的程序包

[http://git.oschina.net/tavenli/port-forward/releases](http://git.oschina.net/tavenli/port-forward/releases)


# 使用交流群

> 使用问题或个性化需求可加QQ号：17020415  （申请时请备注：端口转发）

# 快速安装说明
1. 下载编译好的程序包，并解压程序包
2. 执行 start.sh （Linux）或 start.bat （Win）命令
3. 打开浏览器，进入控制台，打开 http://127.0.0.1:8000/login
4. 输入用户 admin  密码 123456 进入控制台


# 控制台UI
![登录](http://git.oschina.net/tavenli/port-forward/raw/master/screenshot/Login.png "在这里输入图片标题")


![转发列表](http://git.oschina.net/tavenli/port-forward/raw/master/screenshot/List.png "在这里输入图片标题")


![端口转发配置](http://git.oschina.net/tavenli/port-forward/raw/master/screenshot/edit.png "在这里输入图片标题")


![方便与其它平台集成接口](http://git.oschina.net/tavenli/port-forward/raw/master/screenshot/ApiDoc.png "在这里输入图片标题")

