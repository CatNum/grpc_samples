[English](./README.md) | 简体中文
# 生成私有 RSA 秘钥

使用 OpenSSL 工具去生成 RSA 秘钥，我们需要使用 `genrsa` 命令，具体如下：

```shell script
$ openssl genrsa<1> -out server.key<2> 2048<3>
Generating RSA private key, 2048 bit long modulus
....................................................+++
.............................................+++
e is 65537 (0x10001)
```

1. 指定用于创建秘钥的算法。OpenSSl 支持使用不同的算法去创建秘钥，比如 RSA，DSA 和 ECDSA。所有的类型都适合所有的场所。
比如，对 web 服务器秘钥，通常使用 RSA，在我们的情况，我们需要生成 RSA 类型的秘钥。
2. 指定生成秘钥的名字。可以使用任何带有 `.kty` 作为扩展名的名字。
3. 指定秘钥的大小。RSA 秘钥的默认大小是 512 bits，这是不安全的，
因为入侵者可以使用暴力破解您的秘钥，所以我们使用被认为安全的 2048 bit 的 RSA 秘钥，

您可以在这里给秘钥添加密码，如果这样，你在任何时候使用秘钥的时候，都需要密码。在这个例子中，我们不会给秘钥设置密码。
所以我们现在成功的生成了我们的私钥(`server.key`)，我们将在 gRPC 中使用它。让我们生成一个自签名的公共证书来分给我们的客户。
# 生成公钥/证书

一旦我们有了私钥，我们需要创造一个证书。在这个例子中，我们创造一个自签名的证书，换句话说，就是没有证书办法机构（CA）。
通常在生产部署中，您将使用公共证书颁发机构或企业级证书颁发机构来签署您的公共证书。因此，任何信任证书颁发机构的客户都可以进行验证。

在这个例子中，我们的 TLS 服务器是为了我们自己的测试目的，我们可能不想去证书颁发机构（CA）获取公开信任的证书。

让我们执行下面的命令来生成一个自签名的公共证书。证书生成是一个交互式过程，在这个过程中，你会被请求去输入一些将会跟证书合并的信息。
```shell script
$ openssl req -new -x509<1> -sha256<2> -key server.key<3> \
              -out server.crt<4> -days 3650<5>
You are about to be asked to enter information that will be incorporated
into your certificate request.
What you are about to enter is what is called a Distinguished Name or a DN.
There are quite a few fields but you can leave some blank
For some fields, there will be a default value 
If you enter '.', the field will be left blank.
-----
Country Name (2 letter code) []:US
State or Province Name (full name) []:California
Locality Name (eg, city) []:Mountain View
Organization Name (eg, company) []:O’Reilly Media, Inc
Organizational Unit Name (eg, section) []:Publishers
Common Name (eg, fully qualified hostname[]:localhost
Email Address []:webmaster@localhost
```

1. 指定公共证书的格式。X.509 是一个标准格式，用于许多 Internet 协议，包括 TLS/SSL。
2. 指定安全散列算法。
3. 指定我们之前生成的私钥(`server.key`)文件位置。
4. 指定生成证书的名称。可以使用任何以 `.crt` 为扩展名的名字。
5. 指定证书的有效时间为 365 天。

[ 注意 ]
=====
在生成证书时提出的问题中，最重要的问题是“公用名”，它由证书相关的服务器主机、域名或 IP 地址组成。
此名称在验证期间使用，如果主机名与公用名不匹配，则会引发警告。



# 将服务器/客户端秘钥转换为 pem 格式
为了保护 java 应用程序，我们需要提供秘钥库（.pem文件）。我们可以使用以下命令轻松转换服务器和客户端秘钥。
```shell script
$ openssl pkcs8 -topk8 -inform pem -in server.key -outform pem -nocrypt -out server.pem
$ openssl pkcs8 -topk8 -inform pem -in client.key -outform pem -nocrypt -out client.pem
```