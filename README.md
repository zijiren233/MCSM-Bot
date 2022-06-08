# MCSM-Bot

一个MCSM与GO-CQHTTP的附属产物，实现我的世界服务器群组机器人!

### 项目相关

本项目依赖于 **[go-cqhttp](https://github.com/Mrs4s/go-cqhttp)** QQ机器人API，请先安装运行后修改配置文件并登录QQ即可后即可运行本MCSM-Bot。

- 建议把 ``go-cqhttp`` 做为服务启动，或用 screen 运行并不再关闭，不然由于tx风控的原因每次运行都要重新登录扫码！

### 登录出错

如果 go-cqhttp QQ机器人登录不上，可以先在和自己同一个网络环境下登录后，go-cqhttp 会生成 `session.token`

和 `device.json ` ，请复制到远程vps后登录即可。

或者用远程vps搭建节点使手机和vps在同一个网络环境再登录。

-----

### 开始使用

- 启动QQ_API

    下载运行 `go-cqhttp`后会生成配置文件，只需要修改`config.yml`中：

    ```
    account:
        uin: // 用于机器人的QQ号
        password: // 用于机器人的QQ密码
        
    default-middlewares: &default
        access-token: // 设置任意长度字符串
    ```

    修改完成后再次运行`go-cqhttp`完成登录后即可，此时API端口为默认的5700。

- 启动MCSM-Bot

    下载运行程序 **[MCSM-Bot](https://github.com/zijiren233/MCSM-Bot/releases)** 
    首次运行可执行程序会在当前文件夹生成配置文件 `config.json` 
    修改配置后再次运行即可，MCSM-Bot可以在不需要公网的环境下运行。

    如果 MCSM-Bot 和 CQHTTP 在同一台设备或同一个内网，则都不需要公网，MCSM-Bot 配置文件指定内网地址即可。


- config.sample.json :
    ```
    { // 真正的配置文件为标准的json格式，里面不要有注释！！！
    "mcsmdata": [
        {
            "order": 0, // 按顺序填
            "sendtype": "QQ", // 暂时只有QQ
            "name": "LYC_01", // MCSM里面的实例名，即基本信息里的昵称
            "url": "https://mcsm.domain.com:443", // MCSM面板的地址，包含http(s)//，结尾不要有斜杠/
            "remote_uuid": "d6a27b0b13ad44ce879b5a56c88b4d34", // 守护进程的GID
            "uuid": "a8788991a64e4a06b76d539b35db1b16", // 实例的UID
            "apikey": "vmajkfnvklNSdvkjbnfkdsnv7e0f", // 不可为空，用户中心->右上角个人资料->右方生成API密钥
            "group_id": "234532", // 要管理的QQ群号
            "adminlist": [
                "1145141919", // 群管理员，第一个为主管理员，只有管理员才可以发送命令
                "1145141919" // 管理员列表可以为空，则所有用户都可以发送命令
            ]
        }, // 只有一个实例则可以删掉后面的这个order，有多个则自行添加
        {
            "order": 1,
            "sendtype": "TG",
            "name": "server2",
            "url": "http://mcsm.domain.com:24444",
            "remote_uuid": "d6a27b0b13ad44ce879b5ascwfscr323",
            "uuid": "a8788991a6acasfaca76d539b35db1b16",
            "apikey": "6ewc6292daefvlksmdvjadnvjbf",
            "group_id": "234532",
            "adminlist": [
                "1145141919",
                "1145141919"
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

修改完成后运行MCSM-Bot即可。

如果启动失败则为配置文件配置错误。

-----

### 效果展示

<img src="docs\sc\Sample_1.png" />

<img src="docs\sc\Sample_2.png" />

<img src="docs\sc\Sample_3.png" />