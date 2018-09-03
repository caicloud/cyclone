# Cargo 问题排查路径

pull/push 镜像失败，请按如下步骤排查：

- 仓库域名是否正确解析，可以从错误信息里判断
- 是否 docker login
- 在本地用相同账号 docker login 后能否 pull/push，（确保本地域名正确解析，先 ping 一下）
- 如果是权限问题，检查 Cargo 配置的 token service 是否正确
- 如果不是用的 admin 账号，检查权限分配里是否分配了相应访问的权限（访客、操作员、管理员）
- 确保项目在 Cargo-Admin 和 Harbor 中均存在（如果是 admin 账号，Cargo-Admin 中可以不存在）
- 查看 cargo-token 的日志或找陈德（chende@caicloud.io）
- 找陈德（chende@caicloud.io）

进阶步骤

- 确保域名解析没问题
- 确保 Cargo 配置的 token service 没问题
- 查看 cargo-token 的 log 发现问题
- 如果是上传镜像、构建镜像，还需要查看 cargo-admin 的 log
- 解决问题

# 常见问题

## 域名无法解析

### 本地场景
```bash
$ docker pull nonexist.caicloudprivatetest.com/library/redis:latest
Error response from daemon: Get https://nonexist.caicloudprivatetest.com/v2/: Service Unavailable
```

**解决办法**

编辑 /etc/hosts 文件，添加域名解析，如：

```vim
192.168.21.14 nonexist.caicloudprivatetest.com
```

### 流水线场景

```bash
Stage: Push image status: start
The push refers to a repository [devops.caicloudprivatetest.com/devops_github-token-test/test]
Stage: Push image status: fail with error: Fail to push image devops.caicloudprivatetest.com/devops_github-token-test/test:zj-v2 as Get https://devops.caicloudprivatetest.com/v1/_ping: dial tcp: lookup devops.caicloudprivatetest.com on 10.254.0.100:53: no such host
```

**解决办法**

在 POD 中域名无法解析，需要通过修改 `kube-dns` 来配置域名。修改后60秒后生效。

```bash
$ kubectl edit cm kube-dns -n kube-system
```

```yaml
apiVersion: v1
data:
  myhosts: |-
    192.168.21.106 cargo.caicloudprivatetest.com
    192.168.21.17 devops.caicloudprivatetest.com
    192.168.21.16 test.caicloudprivatetest.com
kind: ConfigMap
metadata:
  creationTimestamp: 2018-06-13T06:05:34Z
  labels:
    kubernetes.io/cluster-service: "true"
  name: kube-dns
  namespace: kube-system
```

## 使用错误的域名访问

该问题的最直接表现形式就是，docker login 失败。

假设部署应用的时候用的是 foo.caicloudprivatetest.com 域名，而如果你通过配置 hosts 用 bar.caicloudprivatetest.com 去访问，将会出现以下问题：

```bash
$ docker login bar.caicloudprivatetest.com -u admin -p Pwd123456
Error response from daemon: Get https://bar.caicloudprivatetest.com/v2/: Get https://foo.caicloudprivatetest.com/service/token?account=admin&client_id=docker&offline_token=true&service=harbor-registry: Service Unavailable
```

这个问题会影响大部分仓库的操作，属于比较严重的问题。

**解决办法**

【 方法一 】

使用正确的域名去访问，在上述例子中，是 `foo.caicloudprivatetest.com` 而不是 `bar.caicloudprivatetest.com`。

【 方法二 】

最简单的办法是配置正确的域名重新部署。

【 方法三 】
最快速的办法是，修改配置文件 `/compass/cargo-ansible/cargo/config/registry/config.yml` 中的域名：

```yaml
...
realm: https://<domain>/service/token
...
```

以及 `/compass/cargo-ansible/cargo/config/nginx/nginx.conf` 中的域名：

```bash
...
check_http_send "GET /registryproxy/v2 HTTP/1.0\r\nX-Cargo: <domain>\r\n\r\n";
...
```

之后会提供一个脚本 `change-domain.sh` 来完成域名更新。

修改完之后通过 `restart.sh` 重启 Cargo。


## Token Service 配置错误

Token service 是配置文件 `env-single.yml` 或 `env-ha.yml` 中的 `cargo_token_service` 配置的。有效的配置应该是集群中 Master 节点的 VIP，通过它能够访问到 cargo-token 这个 NodePort 类型的 service

```bash
$ kubectl get svc | grep cargo-token
cargo-token                                                    NodePort    10.254.254.7     <none>        8081:30099/TCP   22d
```

任何无效的配置，包括 127.0.0.1 都会导致 Cargo 使用内置的 token service。

### 使用 Cargo 内置的 Token Service

有两种情况下会导致使用 Cargo 内置的 token service

- Cargo 配置的 token service 是 127.0.0.1
- Cargo 配置的 token service 无效

使用内置 token service 时，Cargo 只支持 admin 账户进行操作，并且拥有所有权限。而平台的其他用户将无法 pull/push 镜像，包括流水线。

### 使用了其他集群里的 Token Service

举例来说，假设 Cargo devops.caicloudprivatetest.com 集成进了 129.10 环境里使用，而它配置的 token service 却是另外的集群，比如 129.30。那么一般情况下，只有 admin 账户能正常工作，而其他平台用户包括流水线都无法正常工作。

**解决办法**

确认 token service 的配置，最准确的方法是直接登录到 Cargo 所在的机器，查看 tengine 这个容器里的 /etc/hosts 文件。

```bash
docker exec -it tengine cat /etc/hosts
127.0.0.1	localhost
::1	localhost ip6-localhost ip6-loopback
fe00::0	ip6-localnet
ff00::0	ip6-mcastprefix
ff02::1	ip6-allnodes
ff02::2	ip6-allrouters
192.168.129.20	cargo-admin
172.19.0.10	a78fb61e1288
```

这里 `192.168.129.20	cargo-admin` 就是 token service 的解析。注意对于老版的 Cargo，这个名字可能不一样，例如 `cargo-admin.caicloudprivatetest.com`。

如果该项配置有错需要修改，请直接编辑 `/compass/cargo-ansible/cargo/config/docker-compose-single.yml` 文件，修改 `- "cargo-admin:192.168.129.20"` 中的 IP 。

```yaml
...
proxy:
    image: cargo.caicloudprivatetest.com/release/tengine:2.1.0
    container_name: tengine
    restart: always
    volumes:
      - /compass/cargo-ansible/cargo/config/nginx:/etc/nginx:z
    networks:
      - cargo
    extra_hosts:
      - "cargo-admin:192.168.129.20"
    ports:
      - 443:443
...
```

对于新版本的 Cargo，也可以通过 `/compass/cargo-ansible/cargo` 下的 `change-token-servce.sh` 脚本来修改配置。

修改完配置后，需要重启 Cargo，

```bash
$ cd /compass/cargo-ansible/cargo
$ ./restart.sh
```

## 镜像同步一直同步中

如果镜像同步的同步记录一直处于同步中无法结束，可以直接查看 Harbor 中的同步任务。

在复制管理中找到对应的同步策略和同步任务，一般情况下可以发现任务处在 `retrying` 状态， 查看任务的日志确认失败重试的原因。

常见的原因是域名无法解析。例如我们要将镜像从 cargo.caicloudprivatetest.com 同步到 devops.caicloudprivatetest.com，但是 cargo.caicloudprivatetest.com 这个 Cargo 中的 `jobservice` 容器内无法解析目的地址 `devops.caicloudprivatetest.com`。要确认这一点，进入到 cargo.caicloudprivatetest.com 所在的机器。

```bash
docker exec -it tengine ping devops.caicloudprivatetest.com
ping: unknown host devops.caicloudprivatetest.com
```

```bash
$ docker exec -it harbor-jobservice cat /etc/hosts
127.0.0.1	localhost
::1	localhost ip6-localhost ip6-loopback
fe00::0	ip6-localnet
ff00::0	ip6-mcastprefix
ff02::1	ip6-allnodes
ff02::2	ip6-allrouters
172.19.0.9	466823d901c3
```

确认 tengine 中无法解析域名 devops.caicloudprivatetest.com，并且 `harbor-jobservice` 的 `/etc/hosts` 也没有配置改域名。（之所以在 tengine 里 ping 是因为 `harbor-jobservice` 中没有 `ping` 命令）

**解决办法**

编辑 Cargo 所在机器的 `/etc/hosts`，添加域名解析。同时编辑 `harbor-jobservice` 容器的 `/etc/hosts`，添加域名解析。

```bash
$ docker exec -it harbor-jobservice sh
sh-4.3# vi /etc/hosts
```

修改完配置后，在 Harbor 中将那些 retrying 状态的任务终止，只需点击复制管理中的【停止任务】按钮。然后在 Cargo-Admin 中再次触发同步。

## 镜像同步同步失败

镜像同步如果失败，或者出现有时成功有时失败的情况，可以直接去 Harbor 上查看同步任务失败的 log。

常见的原因可能是仓库的 cargo token 配置错误，导致 auth 错误（401）：

```bash
[ERROR] [transfer.go:395]: an error occurred while checking existence of blob sha256:22c2dd5ee85dc01136051800684b0bf30016a3082f97093c806152bf43d4e089 of auto2/busybox:latest on https://single-tenant.caicloudprivatetest.com: 401 
```

## 权限问题

```bash
$ docker pull cargo.caicloudprivatetest.com/devops_test-cargo/node:8.9-alpine
Error response from daemon: pull access denied for cargo.caicloudprivatetest.com/devops_test-cargo/node, repository does not exist or may require 'docker login'
```

请通过 docker login 登陆后再 pull 私有项目中的镜像。

如果登陆后依旧出现上述问题，请检查使用的用户是否具有该项目的权限。

没权限情况下，pull/push 会报如下错误：

```bash
$ docker pull cargo.caicloudprivatetest.com/devops_test-cargo/node:8.9-alpine
Error response from daemon: pull access denied for cargo.caicloudprivatetest.com/devops_test-cargo/node, repository does not exist or may require 'docker login'

$ docker push cargo.caicloudprivatetest.com/devops_test-cargo/node:8.9-alpine                                                                                    
The push refers to repository [cargo.caicloudprivatetest.com/devops_test-cargo/node]
27bfa37e00a5: Preparing
be7f582bc675: Preparing
2aebd096e0e2: Preparing
denied: requested access to the resource is denied
```