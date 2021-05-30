another cloudflare dynamic DNS for lede-project write by golang

## build
```shell
GOOS=linux GOARCH=arm go build -o ddns main.go
```

## supported platform
| platform | support |
| --- | --- |
| arm | YES |
| mips | not sure |

## tips
> your can use `upx` to compress binary size for router with little ram space