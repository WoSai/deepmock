# DeepMock

[![Go Report Card](https://goreportcard.com/badge/github.com/qastub/ultron)](https://goreportcard.com/report/github.com/qastub/deepmock)
[![build-state](https://travis-ci.org/qastub/deepmock.svg?branch=master)](https://travis-ci.org/qastub/deepmock)
[![codecov](https://codecov.io/gh/qastub/deepmock/branch/master/graph/badge.svg)](https://codecov.io/gh/qastub/deepmock)

### 下载安装

建议使用docker

```bash
docker run --name deepmock -p 16600:16600 -v `pwd`/log:/app/log -d qastub/deepmock
```

### 快速上手

**创建Mock规则:**

```bash
curl -X POST http://127.0.0.1:16600/api/v1/rule \
  -d '{
    "request": {
        "path": "/whoami",
        "method": "get"
    },
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
- Response可以通过[Go Template](https://golang.org/pkg/text/template/)实现，因此可以支持以下特性
    * 可以使用逻辑控制，如: `if`，`range`
    * 可以使用内置函数
    * 可以自定义函数
- 规则中的`Variable`、`Weight`以及请求中的`Header`、`Query`、`Form`、`Json`同样参与Response模板的渲染

### 接口列表：

#### 创建规则: `POST /api/v1/rule`

请求报文样例

```json
{
    "request": {
        "path": "/(.*)",
        "method": "get"
    },
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
    "request": {
        "path": "/baidu_logo",
        "method": "get"
    },
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
            "request": {
                "path": "/(.*)",
                "method": "get"
            },
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
            "request": {
                "path": "/whoami",
                "method": "get"
            },
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
        "request": {
            "path": "/(.*)",
            "method": "get"
        },
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
        "request": {
            "path": "/whoami",
            "method": "get"
        },
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

### Benchmark

#### 静态response - `is_template: false`

```json
{
	"request": {
		"path": "/echo",
		"method": "get"
	},
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

### 动态Response - `is_template: true`

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
Document Length:        44 bytes

Concurrency Level:      100
Time taken for tests:   26.472 seconds
Complete requests:      1000000
Failed requests:        500213
   (Connect: 0, Receive: 0, Length: 500213, Exceptions: 0)
Keep-Alive requests:    1000000
Total transferred:      210499787 bytes
HTML transferred:       43499787 bytes
Requests per second:    37776.35 [#/sec] (mean)
Time per request:       2.647 [ms] (mean)
Time per request:       0.026 [ms] (mean, across all concurrent requests)
Transfer rate:          7765.54 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    0   0.0      0       3
Processing:     0    3   6.8      3     637
Waiting:        0    3   6.8      3     637
Total:          0    3   6.8      3     637

Percentage of the requests served within a certain time (ms)
  50%      3
  66%      3
  75%      3
  80%      4
  90%      4
  95%      5
  98%      6
  99%      7
 100%    637 (longest request)

```

