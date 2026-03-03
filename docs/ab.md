AB 压测示例

- 单个生成
  ab -n 1000 -c 50 -p scripts\\payloads\\single.json -T application/json http://127.0.0.1:8080/api/links

- 查询
  ab -n 2000 -c 100 http://127.0.0.1:8080/api/links/<CODE>

- 删除
  ab -n 500 -c 50 -m DELETE http://127.0.0.1:8080/api/links/<CODE>

- 批量生成
  ab -n 200 -c 20 -p scripts\\payloads\\batch.json -T application/json http://127.0.0.1:8080/api/links/batch

注意：
- 请先用“单个生成”得到一个 CODE 再进行查询/删除压测。
- Windows 可通过 WSL 或安装 ApacheBench（httpd）后使用 ab。

