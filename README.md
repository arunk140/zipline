# zipline
## Quick TCP (...and UDP) Forwarding using Go Lang

Try 

```sh
go run . -config example.json -vars example.vars.json
```

Output

```
2023/07/13 20:07:57 SSH Server 1 Proxy Running on :2222 Forwarding to 10.144.2.3:22 Type:tcp
2023/07/13 20:07:57 Secure Web Server X Proxy Running on :443 Forwarding to 10.144.2.4:3002 Type:tcp
```

example.json

```json
{
    "silent": false,
    "disable": false,
    "forward": [
        {
            "disable": false,
            "src": ":2222",
            "label": "SSH Server 1",
            "dst": "{{serverip}}:22"
        },
        {
            "type": "https",
            "label": "Secure Web Server X",
            "dst": "10.144.2.4:3002"
        }
    ]
}
```

example.vars.json

```json
{
    "serverip": "10.144.2.3"
}
```

---

Build 

``` sh
go build .
```

Usage 

```sh
❯ ./zipline -h
Usage of ./zipline:
  -config string
        Path to JSON config for Proxies. 
        See example.json for format. (default "proxy.json")
  -vars string
        Path to JSON config for Variables, (Optional)
        File Format { "key1": "value1", "key2": "value2" ...} 
        See example.vars.json for format. 

```
