# DeepMock

[![Go Report Card](https://goreportcard.com/badge/github.com/qastub/ultron)](https://goreportcard.com/report/github.com/qastub/deepmock)
[![build-state](https://travis-ci.org/qastub/deepmock.svg?branch=master)](https://travis-ci.org/qastub/deepmock)
[![codecov](https://codecov.io/gh/qastub/deepmock/branch/master/graph/badge.svg)](https://codecov.io/gh/qastub/deepmock)

### 下载安装

建议使用docker

```bash
docker run -name deepmock -p 16600:16600 -v `pwd`/log:/app/log
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
- 支持设定规则级别的变量(`Context`)，用于在Response中返回
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
- 规则中的`Context`、`Weight`以及请求中的`Header`、`Query`、`Form`、`Json`同样参与Response模板的渲染

### 接口列表：

#### 创建规则: `POST /api/v1/rule`

请求报文样例

```json
{
    "request": {
        "path": "/(.*)",
        "method": "get"
    },
    "context": {
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
        		"body": "{\"hello\":\"{{.Context.name}}\", \"country\":\"{{.Query.country}}\",\"body\":\"{{.Form.nickname}}\"}"
        	}
        }
    ]
}
```

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
    "context": {
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