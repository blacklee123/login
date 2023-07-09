# login

## OAuth登陆流程
[GitHub OAuth 第三方登录示例教程-阮一峰](https://www.ruanyifeng.com/blog/2019/04/github-oauth.html)  
[GitHub官方说明文档](https://docs.github.com/en/developers/apps/building-oauth-apps/authorizing-oauth-apps)

## 密钥生成
```
openssl genrsa -out private_key.pem 1024
openssl rsa -in private_key.pem -pubout -out public_key.pem
```

## Usage
启动前需要设置两个github app相关的环境变量
```
LOGIN_CLIENTID=你的应用id
LOGIN_SECRETS=你的应用secrets
```
github app callback url设置为 yourhost:yourhost/web/login/callback  
启动后访问：yourhost:yourport 即可触发登录流程
## Usage with docker
`docker run -d --name login -p 7777:7777 --env LOGIN_CLIENTID=YOUR_LOGIN_CLIENTID --env LOGIN_SECRETS=YOUR_LOGIN_SECRETS blacklee123/login-github`

## Deploy with nginx

Recommended configuration, assume your login listening on `127.0.0.1:7777`

```nginx
server {
    listen 443 ssl;
    server_name qaq.com;

    location = /web {
    	return 302 /web/;
    }
    location /web/ {
       proxy_set_header X-Real-IP $remote_addr;
       proxy_set_header  Host    $host;
       proxy_set_header  X-Forwarded-For $proxy_add_x_forwarded_for;
       proxy_set_header  X-Forwarded-Proto $scheme; 
       proxy_pass http://127.0.0.1:7777;
    }
}
```

## 目标
平台以及各子服务间需要保持登陆状态的共享，即从用户从平台任何子站点登陆或者登出，其登陆状态都是共享给整个平台的。例如：
- 用户从子站点A登陆，进入子站点B不需要重新登陆
- 用户从子站点A登出，对于子站点B来说用户也是登出

所以，子站点和主站点需要统一用户标记，这个用户标记是一个存放在浏览器中的一个http-only 的cookie。特性如下：
- cookie名称是jwt
- 是一个JWT Token
- 采用RSA加密
- JWT 相关资料 https://jwt.io/

## 方案概述
1. 所有子站点的登陆都跳转到qaq平台的统一登陆url，附带子站点的回调地址
2. 浏览器redirect 到qaq平台统一登陆url的时候，qaq平台引导用户完成OpenID登陆请求
3. 登陆成功后，qaq平台生成JWT Token
    1. 将这个token以cookie的形式set到请求的域名下
    2. 让浏览器redirect到子站点的回调地址，同时附带参数jwt
    3. 同时设置cookie email以及fullname，方便前端获取登陆用户的email地址和姓名
4. 子站点需要做如下事情
    1. 验证jwt cookie（==同时也是判断用户是否登陆的标志==），验证方式见下文
    2. 如果有需要设定子站点自己的用户标识符（有效期需要和平台的jwt cookie保持一致）
5. 用户登出
    1. 子站点先删除自己定义的用户标识符
    2. 向主站点发出登出请求
   
## 登录相关功能：

### 主要流程
1. 前端发起登录请求 /web/login?next=callbackUrl
2. 主站点服务端收到登录请求
3. 服务端让前端redirect到OpenID的登录页面
4. 用户填信息，点击登录,OpenID服务器验证通过后，让前端redirect到主站点/web/login/callback?xxxxx 参数中包含了很多信息
5. qaq主站点收到OpenID跳转回的请求后，获得用户信息
    1. 生成jwt token
    2. set jwt cookie
    3. set fullname cookie
    4. set email cookie
    5. 让前端redirect到callbackUrl

7. 总的跳转
```
/web/login?next=callbackUrl -> openid -> /web/login/callback -> callbackUrl
```


## 登出
调用登出接口`/web/logout?next=urlA`
删除jwt cookie即可
