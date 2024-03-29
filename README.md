# DeepMock

[![Go Report Card](https://goreportcard.com/badge/github.com/WoSai/deepmock)](https://goreportcard.com/report/github.com/WoSai/deepmock) 
![master](https://github.com/WoSai/deepmock/workflows/master/badge.svg?branch=master) 
[![codecov](https://codecov.io/gh/WoSai/deepmock/branch/master/graph/badge.svg)](https://codecov.io/gh/WoSai/deepmock) 
[![GoDoc](https://godoc.org/github.com/WoSai/deepmock?status.svg)](https://godoc.org/github.com/WoSai/deepmock)


### 下载安装

建议使用docker

```bash
docker run --name deepmock -p 16600:16600 wosai/deepmock
```

### 快速上手

**创建Mock规则:**

```bash
curl -X POST http://127.0.0.1:16600/api/v1/rule \
  -d '{
    "path": "/whoami",
    "method": "get",
    "responses": [
        {
            "is_default": true,
            "response": {
                "header": {
                    "Content-Type": "application/json"
                },
                "body": "{\"im\": \"deepmock\"}"
            }
        }
    ]
}'
```

### DeepMock的特性

- 可以以正则表达式声明Mock接口的Path，以便支持RESTFul风格的请求路径
- 支持设定规则级别的变量(`Variable`)，用于在Response中返回
- 支持设定规则级别的随机值(`Weight`)，并配以权重，权重越高返回概率越高
- 单个规则支持多Response模板，并通过筛选器`filter`来命中相应模板
- 筛选器支持QueryString、HTTP Header、Body
- 筛选器支持四种模板：
    * `always_true`: 必定筛选成功
    * `exact`: 精确筛选
    * `keyword`: 关键字筛选
    * `regular`: 正则表达式筛选
- Response中的body和header均可以通过[Go Template](https://golang.org/pkg/text/template/)实现，因此可以支持以下特性
    * 可以使用逻辑控制，如: `if`，`range`
    * 可以使用内置函数
    * 可以自定义函数
- 规则中的`Variable`、`Weight`以及请求中的`Header`、`Query`、`Form`、`Json`同样参与Response模板的渲染
- Response.body中使用template渲染时，需设置`is_template: true`
- Response.header可采用Patch形式，使用template渲染部分Header字段，需设置`render_template: true`以及template字符串`header_template`

### 接口列表：

#### 创建规则: `POST /api/v1/rule`

请求报文样例

```json
{
    "path": "/(.*)",
    "method": "get",
    "variable": {
        "value": "123",
        "name":"jack",
        "version": 123
    },
    "weight": {
        "code": {
            "CREATED": 2,
            "CLOSED": 1
        }
    },
    "responses": [
        {
            "is_default": true,
            "filter":{
            	"header": {"Content-Type":"text/xml", "mode":"exact"}
            },
            "response": {
                "header": {
                    "Content-Type": "application/json"
                },
                "status_code": 200,
                "body": "{\"hello\": \"world\"}"
            }
        },
        {
        	"filter": {
        		"query": {
        			"country": "china",
        			"mode":"exact"
        		}
        	},
        	"response": {
        		"is_template": true,
        		"header": {
        			"Content-Type": "application/json"
        		},
        		"body": "{\"hello\":\"{{.Variable.name}}\", \"country\":\"{{.Query.country}}\",\"body\":\"{{.Form.nickname}}\"}"
        	}
        }
    ]
}
```

DeepMock支持返回二进制报文，只需要对二进制内容进行base64编码后传入即可，如：

```json
{
    "path": "/baidu_logo",
    "method": "get",
    "responses": [
        {
            "is_default": true,
            "response": {
                "header": {
                    "Content-Type": "image/png"
                },
                "base64encoded_body": "0KGgoAAAANSUhEUgAAAh...g9Qs4AAAAASUVORK5CYII="
            }
        }
    ]
}
```

![](https://my-storage.oss-cn-shanghai.aliyuncs.com/picgo/20190831183004.png)

### 获取规则详情： `GET /api/v1/rule/<rule_id>`

```bash
curl http://127.0.0.1:16600/api/v1/rule/bba079deaa2b97037694a89386616d88
```

### 完整更新规则: `PUT /api/v1/rule`

报文与创建接口一致，**规则ID必须存在才允许更新**

### 部分更新规则: `PATCH /api/v1/rule`

请求报文样例：

```json
{
    "id": "bba079deaa2b97037694a89386616d88",
    "variable": {
        "value": "456"
    },
    "weights": {
        "code": {
            "CLOSED": 0
        }
    },
    "responses": [
        {
            "default": true,
            "filter": {
                "header": {
                    "mode": "exact",
                    "apiCode": "xxx"
                },
                "body": {
                    "mode": "keyword",
                    "keyword": "createStore"
                },
                "query": {
                    "mode": "regular"
                }
            }
        }
    ]
}
```

**如果在该接口中传入`.response`，将会清空原有的response regulation**

### 根据ID删除规则: `DELETE /api/v1/rule`

```json
{
  "id": "bba079deaa2b97037694a89386616d88"
}
```

### 导出所有规则 `GET /api/v1/rules`

响应报文如下：

```json
{
    "code": 200,
    "data": [
        {
            "id": "bba079deaa2b97037694a89386616d88",
            "path": "/(.*)",
            "method": "get",
            "responses": [
                {
                    "is_default": true,
                    "response": {
                        "header": {
                            "Content-Type": "application/json"
                        },
                        "body": "{\"name\": \"deepmock\"}"
                    }
                }
            ]
        },
        {
            "id": "ccf2e319d7d51ff3a73b1c704d77b0c1",
            "path": "/whoami",
             "method": "get",
            "responses": [
                {
                    "is_default": true,
                    "response": {
                        "header": {
                            "Content-Type": "application/json"
                        },
                        "body": "{\"im\": \"deepmock\"}"
                    }
                }
            ]
        }
    ]
}
```

### 导入规则 `POST /api/v1/rules`

**注意调用该接口会清空原有规则**

```json
[
    {
        "id": "bba079deaa2b97037694a89386616d88",
        "path": "/(.*)",
        "method": "get",
        "responses": [
            {
                "is_default": true,
                "response": {
                    "header": {
                        "Content-Type": "application/json"
                    },
                    "body": "{\"name\": \"deepmock\"}"
                }
            }
        ]
    },
    {
        "id": "ccf2e319d7d51ff3a73b1c704d77b0c1",
        "path": "/whoami",
        "method": "get",
        "responses": [
            {
                "is_default": true,
                "response": {
                    "header": {
                        "Content-Type": "application/json"
                    },
                    "body": "{\"im\": \"deepmock\"}"
                }
            }
        ]
    }
]
```

### 过滤器Filter设置规则

#### Header Filter

精确模式

```json
{
    "filter": {
        "header": {
            "mode": "exact",
            "Authorization": "balabala",
            "Content-Type": "application/json"
        }
    }
}
```

关键字模式

```json
{
    "filter": {
        "header": {
            "mode": "keyword",
            "Content-Type": "json"
        }
    }
}
```

正则匹配模式

```json
{
    "filter": {
        "header": {
            "mode": "regular",
            "Authorization": "[0-9]+"
        }
    }
}
```

#### Query Filter

精确模式

```json
{
    "filter": {
        "query": {
            "mode": "exact",
            "code": "balabala",
            "version": "1.0"
        }
    }
}
```

关键字模式

```json
{
    "filter": {
        "query": {
            "mode": "keyword",
            "version": "1"
        }
    }
}
```

正则匹配模式

```json
{
    "filter": {
        "query": {
            "mode": "regular",
            "version": "[0-9.]+"
        }
    }
}
```

#### Body Filter

**暂时不支持精确匹配模式**

关键字模式

```json
{
    "filter": {
        "body": {
            "mode": "keyword",
            "keyword": "store"  // 必须使用该key值
        }
    }
}
```

正则匹配模式

```json
{
    "filter": {
        "query": {
            "mode": "regular",
            "regular": "[0-9.]+"  // 必须使用该key值
        }
    }
}
```

### Response模板内置函数

| 内置函数 | 参数 |使用方法 |说明 |
| :---: | ---- | ---- | --- |
|`uuid` | 无 | `{{ uuid }}`|返回一个uuid字符串|
|`date`| `layout` | `{{date "layout"}}` | 按指定的格式返回当前日期，[参考链接](https://golang.google.cn/pkg/time/) |
|`timestamp` | `precision` | `{{timestamp ms}}` | 按指定的精度返回unix时间戳：mcs,ms,sec|
|`plus`| `v`, `i` | `{{plus v i}}` | 将v的值增加i，实现简单的计算，支持string\int\float类型|
|`rand_string`| `n` | `{{rand_string n}}`| 生成长度为n的随机字符串 |
|`html_unescaped`|无|`{{.Variable.str `&#x7c;`html_unescaped}}`| 防止template渲染时，将部分字符进行html编码，如"&" --> "\&amp;" |
 
#### Response.Header使用Template渲染示例
```json
{
  "path": "/redirect/baidu",
  "method": "get",
  "variable": {
    "app_id": "app_id",
    "code": "123456"
  },
  "responses": [
    {
      "is_default": true,
      "response": {
        "is_template": false,
        "render_header": true,
        "header": {
          "Content-Type": "text/html",
          "User-Agent": "Wechat",
          "x-env-flag": "test-111",
          "rand-string": "I'm a {{rand_string}}",
          "timestamp-sec": "2021-01-01",
          "uuid": "I'm a {{uuid}}",
          "location": "https://www.bing.com"
        },
        "header_template": "{\"location\": \"{{.Query.redirect_uri | html_unescaped}}&state={{.Query.state}}&app_id={{.Variable.app_id}}&auth_code={{.Variable.code}}\",\"rand-string\":\"{{rand_string 20}}\",\"uuid\":\"{{uuid}}\", \"timestamp-sec\":\"{{timestamp \"sec\"}}\"}",
        "status_code": 302
      }
    }
  ]
}
```
```
curl -I --location --request GET 'http://localhost:16600/redirect/baidu?redirect_uri=https%3A%2F%2Fwww.baidu.com%3Fref%3Dhello&appid=appid&state=true'

HTTP/1.1 302 Found
Server: DeepMock Service
Date: Fri, 03 Dec 2021 03:49:55 GMT
Content-Type: text/html
Content-Length: 0
Location: https://www.baidu.com?ref=hello&state=true&app_id=app_id&auth_code=123456
Rand-String: YHYEglVGzYB8FIIKbify
Timestamp-Sec: 1638503396
Uuid: 1cb6c914-04d7-40fe-908f-73d202bb63ea
X-Env-Flag: test-111
User-Agent: Wechat
```

### Benchmark

#### 静态response - `is_template: false`

```json
{
	"path": "/echo",
	"method": "get",
	"responses": [
		{
			"is_default": true,
			"response": {
				"body": "hello deepmock"
			}
		}
	]
}
```

```bash
ab -c 100 -n 1000000 -k http://127.0.0.1:16600/echo

Server Software:        DeepMock
Server Hostname:        127.0.0.1
Server Port:            16600

Document Path:          /echo
Document Length:        14 bytes

Concurrency Level:      100
Time taken for tests:   14.863 seconds
Complete requests:      1000000
Failed requests:        0
Keep-Alive requests:    1000000
Total transferred:      181000000 bytes
HTML transferred:       14000000 bytes
Requests per second:    67280.35 [#/sec] (mean)
Time per request:       1.486 [ms] (mean)
Time per request:       0.015 [ms] (mean, across all concurrent requests)
Transfer rate:          11892.33 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    0   0.0      0       2
Processing:     0    1   1.3      1      10
Waiting:        0    1   1.3      1      10
Total:          0    1   1.3      1      10

Percentage of the requests served within a certain time (ms)
  50%      1
  66%      2
  75%      3
  80%      3
  90%      3
  95%      4
  98%      4
  99%      5
 100%     10 (longest request)

```

#### 动态Response - `is_template: true`

创建如下规则:

```json
{
	"request": {
		"path": "/render",
		"method": "get"
	},
	"variable": {
		"name": "mike"
	},
	"weight": {
		"return_code": {
			"SUCCESS": 10,
			"FAILED": 10
		}	
	},
	"responses": [
		{
			"is_default": true,
			"filter": {
				"query": {
					"mode": "regular",
					"age": "[0-9]+"
				}
			},
			"response": {
				"is_template": true,
				"body": "{{.Weight.return_code}}: my name is {{.Variable.name}}, i am {{.Query.age}} years old."
			}
		}
		]
}
```

```bash
ab -c 100 -n 1000000 -k http://127.0.0.1:16600/render?age=12

Server Software:        DeepMock
Server Hostname:        127.0.0.1
Server Port:            16600

Document Path:          /render?age=12
Document Length:        43 bytes

Concurrency Level:      100
Time taken for tests:   24.930 seconds
Complete requests:      1000000
Failed requests:        499830
   (Connect: 0, Receive: 0, Length: 499830, Exceptions: 0)
Keep-Alive requests:    1000000
Total transferred:      210499830 bytes
HTML transferred:       43499830 bytes
Requests per second:    40112.24 [#/sec] (mean)
Time per request:       2.493 [ms] (mean)
Time per request:       0.025 [ms] (mean, across all concurrent requests)
Transfer rate:          8245.72 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    0   0.0      0       2
Processing:     0    2   1.6      3      14
Waiting:        0    2   1.6      3      14
Total:          0    2   1.6      3      14

Percentage of the requests served within a certain time (ms)
  50%      3
  66%      3
  75%      3
  80%      4
  90%      4
  95%      5
  98%      5
  99%      6
 100%     14 (longest request)
```

