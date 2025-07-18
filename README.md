简单读取写入yaml文件，由AI生成。
想用来给n8n用于网络配置。

webui端口: 7788


api:

read: GET
http://localhost:7788/api/yaml

write: POST
http://localhost:7788/api/yaml


推送tag后会自动构建
```sh
git tag v0.0.1
git push --tags
```