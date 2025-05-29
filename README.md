# MinIO Demo

## 二进制安装

| os      | url                                          |
| ------- | -------------------------------------------- |
| linux   | https://min.io/docs/minio/linux/index.html   |
| macos   | https://min.io/docs/minio/macos/index.html   |
| windows | https://min.io/docs/minio/windows/index.html |

## docker 安装

docker-compose.yaml

```
services:
  minio:
    image: minio/minio
    container_name: minio
    ports:
      - "9000:9000" # S3 API 端口
      - "9001:9001" # 控制台端口
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    volumes:
      - minio_data:/data
    command: server /data --console-address ":9001"

volumes:
  minio_data:
```

运行

```
docker compose up -d
```

## 登录控制台

```
http://localhost:9001
```

## 命令行模式

```
# 设置一个服务别名（连接信息）
mc alias set local-admin http://127.0.0.1:9000 minioadmin minioadmin

# 删除一个别名
mc alias remove local

# 查看所有别名
mc alias list

# 创建一个桶，桶名称必须符合 S3 规范：小写字母、数字和连字符组成，不能大写。 mb = make bucket
# 一个用户创建桶后，默认只有该用户有访问权限
mc mb local-admin/my-bucket

# 删除一个桶
mc rm --recursive --force local-admin/my-bucket

# 列出别名为 local 的 MinIO 实例里的所有桶
mc ls local-admin

# 上传文件到某个桶
mc cp girl.png local-admin/my-bucket

# 列出某个桶的所有文件 (根目录下所有)
mc ls local-admin/my-bucket

# 列出某个桶的所有文件 (包含子目录)
mc ls --recursive local-admin/my-bucket
mc ls -r local-admin/my-bucket

# 重命名文件 (复制 + 删除)
mc cp local-admin/my-bucket/girl.png local-admin/my-bucket/beauty.png
mc rm local-admin/my-bucket/girl.png

# 下载桶中某个文件到当前目录
mc cp local-admin/my-bucket/beauty.png .

# 查看桶策略（权限）
mc anonymous get local-admin/my-bucket

# 设置桶策略 - 允许匿名下载（可读）
mc anonymous set download lolocal-admincal/my-bucket

# 设置桶策略 - 允许匿名上传（可写）
mc anonymous set upload local-admin/my-bucket

# 设置桶策略 - 允许匿名读写（下载+上传）
mc anonymous set public local-admin/my-bucket

# 设置桶策略 - 禁止匿名访问（默认）
mc anonymous set private local-admin/my-bucket

# 针对 private 桶的某个文件生成临时下载链接，临时公开（预签名 URL）
mc share download --expire=60m local-admin/my-bucket/beauty.png

# 用管理员账号 minioadmin 连接 MinIO，并添加了一个新用户
mc admin user add local-admin guobin guobin123

# 给 guobin 分配 consoleAdmin 权限 (这是 MinIO 内置的管理员策略，赋予用户所有管理员权限)
mc admin policy detach local-admin readwrite --user guobin
mc admin policy attach local-admin consoleAdmin --user guobin

# 查看用户列表 (不会列出root用户)
mc admin user list local-admin

# 让 guobin 用户创建其他用户
mc alias set local-guobin http://localhost:9000 guobin guobin123
mc admin user add local-guobin binguo binguo123

# 给 binguo 分配 readwrite 权限 (不能执行管理操作，比如用户管理、策略管理、系统配置等)
mc admin policy attach local-guobin readwrite --user binguo # guobin 用户可以给 binguo 分配权限
or
mc admin policy attach local-admin readwrite --user binguo # root 用户也可以给 binguo 分配权限

# 删除用户
mc admin user remove local-guobin binguo # guobin 用户可以删除 bingguo
or
mc admin user remove local-admin binguo  # root 用户也可以删除 binguo

# 查看用户权限
mc admin user info local-admin binguo
mc admin user info local-admin guobin

# 查看内置策略文件
mc admin policy info local-admin readwrite
mc admin policy info local-admin consoleAdmin
mc admin policy info local-admin readonly
mc admin policy info local-admin writeonly
```
