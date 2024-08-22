### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询负载均衡下的监听器数量。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/load_balancers/listeners/count

### 输入参数

| 参数名称   | 参数类型       | 必选 | 描述          |
|-----------|--------------|------|--------------|
| bk_biz_id | int64        | 是   | 业务ID        |
| lb_ids    | string array | 是   | 负载均衡ID数组 |

### 调用示例

```json
{
  "lb_ids": [
    "00000001",
    "00000002"
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "details": [
      {
        "lb_id": "00000001",
        "num": 10
      },
      {
        "lb_id": "00000002",
        "num": 2
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型       | 描述    |
|---------|--------------|---------|
| code    | int          | 状态码   |
| message | string       | 请求信息 |
| data    | object array | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述         |
|---------|--------|--------------|
| details | array  | 查询返回的数据 |

#### data.details[n]

| 参数名称  | 参数类型  | 描述       |
|----------|---------|------------|
| lb_id    | string  | 负载均衡ID  |
| num      | int     | 监听器数量   |