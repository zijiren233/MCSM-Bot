# MCSM-Bot

- 一个 **[MCSM](https://github.com/MCSManager/MCSManager)** 与 **[go-cqhttp](https://github.com/Mrs4s/go-cqhttp)** 的附属产物，实现 `MCSManager 管理` 机器人!

- 版本要求：`mcsm-web:9.5.0 & mcsm-daemon:1.6.0` 及更高，并开启实例设置中的 `仿真终端`

- 采用高并发模式，快速高效处理多群组消息 (多群组同时处理建议有较高的网络性能)

### 项目相关

本项目依赖于 **[go-cqhttp](https://github.com/Mrs4s/go-cqhttp)** QQ机器人API，请先安装运行后修改配置文件并登录QQ即可后即可运行本MCSM-Bot。

- 建议把 ``go-cqhttp`` 做为服务启动，或用 screen 运行并不再关闭，不然由于tx风控的原因每次运行都要重新登录扫码！

<br>

-----

<br>

### 开始使用

#### 1.启动QQ_API

下载 **[go-cqhttp](https://github.com/Mrs4s/go-cqhttp)** 运行后选择 2.正向WS通信，启动后 `go-cqhttp` 会生成配置文件，只需要修改 `config.yml` 中：
```
account里面的：
    uin: // 用于机器人的QQ号
    password: // QQ密码
```

修改完成后再次运行 `go-cqhttp` 扫码登录后即可，此时 **QQ API** WS 端口为默认的 8080。

- 登录出错

    如果 go-cqhttp QQ机器人登录不上，可以先在和自己同一个网络环境下的 **windows** 安装 go-cqhttp ，在 windows 下扫码登录成功后会生成 `session.token` 和 `device.json` 两个文件，请复制替换到远程 vps 后登录即可。

<br>

#### 2.启动MCSM-Bot

- 下载运行程序 **[MCSM-Bot](https://github.com/zijiren233/MCSM-Bot/releases)** 
首次运行可执行程序会在当前文件夹生成配置文件 `config.json` ,按照下面的 `config.sample.json` 修改配置后再次运行即可，MCSM-Bot可以在不需要公网的环境下运行。

- 如果 MCSM-Bot 和 CQHTTP 在同一台设备或同一个内网，则都不需要公网，MCSM-Bot 配置文件指定内网地址或者本机环回地址即可。

<br>

- config.example.yaml :

```yaml
mcsmdata:
  - id: 1 # Id 为任意小于256的数,但不可重复!
    url: https://mcsm.domain.com:443 # MCSM面板的地址,包含http(s)://,结尾不要有斜杠/
    remote_uuid: d6a27b0b13ad44ce879b5a56c88b4d34 # 守护进程的GID
    uuid: a8788991a64e4a06b76d539b35db1b16 # 实例的UID
    apikey: vmajkfnvklNSdvkjbnfkdsnv7e0f # 不可为空，用户中心->右上角点蓝色用户名->个人资料->右方生成API密钥
    group_id: # 要管理的QQ群号
      - 383033610
      - 1145141919
    user_allows_commands: # 所有群成员均可运行的命令,填正则表达式
      - ^help$
      - ^list$
      - ^status$
    adminlist: # 群管理员，第一个为主管理员，只有管理员才可以发送命令,管理员列表可以为空，则所有用户都可以发送命令
      - 1670605849
      - 1145141919
  - id: 2
    url: http://mcsm.domain.com:24444
    remote_uuid: 3ec8d0ff584c43bd95598b18949a8bac
    uuid: 76a49c5ef46a41f29b374109d58f994a
    apikey: vmajkfnvklNSdvkjbnfkdsnv7e0f
    group_id:
      - 383033610
    user_allows_commands:
      - ^help$
      - ^status$
    adminlist:
      - 1670605849
      - 1145141919

cqhttp:
url: ws://q-api.pyhdxy.com:8080 # cqhttp 请求地址,末尾不带斜杠!只能使用Ws(s)协议
adminlist: # 可以私聊机器人以访问所有实例,填服务器所有者的QQ号,用于管理所有实例
  - 1670605849
  - 1145141919
```

    <img src="docs\sc\Sample_4.png" />

- 修改完成后运行MCSM-Bot即可。

- 如果启动失败则为配置文件配置错误或 MCSM/CQHTTP 服务连接失败。

<br>

### 参数

```shell
  -dlp
        Disable Log Print
  -log uint
        记录命令日志的级别 0:Debug 1:Info 2:Warning 3:Error 4:Fatal 5:None (default 1)
```

-----

<br>

### 普通命令

- * 代表监听此群的所有实例,但是只有你设置的管理员才能运行

```
括号内的 id 可省略

普通命令就是可在MC控制台直接运行的命令，比如 set time day

run (id/*) list

run (id/*) tps

run (id/*) weather clear

...

控制台内可运行的命令在群内都可以输入！
```

<br>

### 特殊命令

```
括号内的 id 可省略

run (id) help 查看帮助

run (id/*) status 查看服务器运行状态

run (id/*) start 启动服务器

run (id/*) stop 关闭服务器

run (id/*) restart 重启服务器

run (id/*) kill 终止服务器
```

### 效果展示

<img src="docs\sc\Sample_1.png" />

<img src="docs\sc\Sample_2.png" />

<img src="docs\sc\Sample_3.png" />

<img src="docs\sc\Sample_status.png" />