# MCSM-Bot

- 一个 **[MCSM](https://github.com/MCSManager/MCSManager)** 与 **[go-cqhttp](https://github.com/Mrs4s/go-cqhttp)** 的附属产物，实现我的世界服务器群组机器人!

- 采用高并发模式，快速高效处理多群组消息 (多群组同时处理建议有较高的网络性能)

### 项目相关

本项目依赖于 **[go-cqhttp](https://github.com/Mrs4s/go-cqhttp)** QQ机器人API，请先安装运行后修改配置文件并登录QQ即可后即可运行本MCSM-Bot。

- 建议把 ``go-cqhttp`` 做为服务启动，或用 screen 运行并不再关闭，不然由于tx风控的原因每次运行都要重新登录扫码！

-----

### 开始使用

#### 1.启动QQ_API

下载 **[go-cqhttp](https://github.com/Mrs4s/go-cqhttp)** 运行后选择 0.HTTP通信，启动后 `go-cqhttp` 会生成配置文件，只需要修改 `config.yml` 中：

```

account里面的：
    uin: // 用于机器人的QQ号
    
default-middlewares里面的：
    access-token: // 设置任意长度字符串

```

修改完成后再次运行 `go-cqhttp` 扫码登录后即可，此时 **QQ API** 端口为默认的5700。

- 登录出错

    如果 go-cqhttp QQ机器人登录不上，可以先在和自己同一个网络环境下的 **windows** 安装 go-cqhttp ，在 windows 下扫码登录成功后会生成 `session.token` 和 `device.json` 两个文件，请复制替换到远程 vps 后登录即可。

- 如果你实在登录不上去，可以使用我提供的测试API地址
    - 方法：
        - 先添加 **公共机器人QQ** 为好友 ： 3426898431
        - 然后邀请机器人进入群即可

        ```

        "token": "test",
        "url": "https://q-api.pyhdxy.com:443",
        "qq": "3426898431"

        ```

#### 2.启动MCSM-Bot

- 下载运行程序 **[MCSM-Bot](https://github.com/zijiren233/MCSM-Bot/releases)** 
首次运行可执行程序会在当前文件夹生成配置文件 `config.json` ,按照下面的 `config.sample.json` 修改配置后再次运行即可，MCSM-Bot可以在不需要公网的环境下运行。

- 如果 MCSM-Bot 和 CQHTTP 在同一台设备或同一个内网，则都不需要公网，MCSM-Bot 配置文件指定内网地址或者本机环回地址即可。

- config.sample.json :

    ```

    { // 真正的配置文件为标准的json格式，里面不要有注释！！！
        "mcsmdata": [
            {
                "id": 0, // 按顺序填,此项为监听服务器的序号，从0开始依次增加，用于启动监听时填的要监听哪一个服务器
                "name": "server1", // MCSM里面的实例名，即基本信息里的昵称，实例名不可重复！！！
                "url": "https://mcsm.domain.com:443", // MCSM面板的地址，包含http(s)//，结尾不要有斜杠/
                "remote_uuid": "d6a27b0b13ad44ce879b5a56c88b4d34", // 守护进程的GID
                "uuid": "a8788991a64e4a06b76d539b35db1b16", // 实例的UID
                "apikey": "vmajkfnvklNSdvkjbnfkdsnv7e0f", // 不可为空，用户中心->右上角个人资料->右方生成API密钥
                "group_id": "234532", // 要管理的QQ群号
                "adminlist": [
                    "1145141919", // 群管理员，第一个为主管理员，只有管理员才可以发送命令
                    "1433223" // 管理员列表可以为空，则所有用户都可以发送命令
                ]
            }, // 只有一个实例可以删掉后面的服务器，有多个则自行添加
            {
                "id": 1, // 按顺序填，0，1，2，3 ......
                "name": "server2",
                "url": "http://mcsm.domain.com:24444",
                "remote_uuid": "d6a27b0b13ad44ce879b5ascwfscr323",
                "uuid": "a8788991a6acasfaca76d539b35db1b16",
                "apikey": "6ewc6292daefvlksmdvjadnvjbf",
                "group_id": "234532",
                "adminlist": [
                    "114514", // 不同实例在同一个群也可以有不同的管理员
                    "1919"
                ]
            } // <--最后一个实例配置这里没有逗号！！！
        ],
        "cqhttp": {
            "token": "test", // 默认中间件锚点中的access-token，不可为空
            "url": "http://10.10.10.4:5700", // cqhttp 请求地址，末尾不带斜杠！
            "qq": "3333446431" // 机器人QQ号
        }
    }

    ```

    <img src="docs\sc\Sample_4.png" />

- 修改完成后运行MCSM-Bot即可。

- 如果启动失败则为配置文件配置错误或 MCSM/CQHTTP 服务连接失败。

-----

### 普通命令

```

括号内的 order 可省略，则优先输出第一个监听此群的服务器

普通命令就是可在MC控制台直接运行的命令，比如 set time day

run (order) list

run (order) tps

run (order) weather clear

...

控制台内可运行的命令在群内都可以输入！

```

### 特殊命令

```

括号内的 order 可省略，则优先输出第一个监听此群的服务器

run (order) status 查看服务器运行状态

run (order) start 启动服务器

run (order) stop 关闭服务器

run (order) restart 重启服务器

run (order) kill 终止服务器

```

### 效果展示

<img src="docs\sc\Sample_1.png" />

<img src="docs\sc\Sample_2.png" />

<img src="docs\sc\Sample_3.png" />

<img src="docs\sc\Sample_5.png" />