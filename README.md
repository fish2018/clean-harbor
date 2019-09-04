# clean-harbor
清理harbor镜像，只保留最新的n个。使用go和python各实现了一版，python版中有详尽注释及物理清理方法，golang版中不再赘述。

## 克隆代码
```
git clone https://github.com/fish2018/clean-harbor.git
# python版
cd clean-harbor/python/
# golang版
cd clean-harbor/golang/
```

## python版使用方法：
### 修改配置
修改脚本harborgc.py,根据自己情况设置harbor地址、用户名、密码和要保留的最近的镜像数量
```
harbor_domain="harbor.test.com",
                        username="username",
                        password="password",
                        num=10)
```

### 安装依赖
```
pip install requests
```

### 执行程序
```
python3 harborgc.py
```

## golang版使用方法

## 修改配置
修改config.yaml文件，根据自己情况设置harbor地址、用户名、密码和要保留的最近的镜像数量
```
num: 10 # 需要保留最新的tag数
harbor:
  url: "https://harbor.test.com"
  username: "username"
  password: "password"
```

### 安装依赖
```
export GOPROXY=https://goproxy.cn
go build
```

# 执行程序
```
./harbor-clean
```
