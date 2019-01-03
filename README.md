# rock

Scissor Rock Paper !!! Let's go

## Mixin Payment

### Generate Payment URL

```
mixin://pay?amount=0.01&asset=c94ac88f-4671-3976-b60a-09064f1811e8&memo=gqNOdW0yok9wojw9&recipient=3bf07a88-e5cb-4508-a817-b0fad3f5e473&trace=672604b4-50fc-4174-a777-7caa4f8f2647
```

> amount, asset is the pay out amount and asset id of the coin;
> trace is an UUID string; recipient is the mixin id of rock engine.

### Create Arena

**Memo**

```javascript
{
    "E" : 2,    // 2 hours
    "M" : "1.2" // max bet
}
```

### Challenge

**Memo**

```javascript
{
    "A" : "Ahsgdhsa", // arena id
    "G" : "012" // gestures
}
```

| num | gesture |
| --- | ------- |
| 0   | Scissor |
| 1   | Rock    |
| 2   | Paper   |

### Encoding Memo

Msgpack encode first and then base64 encode.

```javascript
// example in Go
action := map[string]interface{}{
    "E" : 2,
    "M" : "100",
}
data,_ := msgpack.Marshal(action)
memo := base64.StdEncoding.EncodeToString(data)
```

## API

### Get Config

```
GET /api/config
```

**Response**

```javascript
{
  "client_id": "3bf07a88-e5cb-4508-a817-b0fad3f5e473" // engine's mixin id
}
```

### Get New Arenas

```
GET /api/arenas/new
```

**Parameters**

| key      | option | desc               |
| -------- | ------ | ------------------ |
| asset_id | true   | asset id           |
| cursor   | true   | cursor             |
| limit    | true   | max 50, default 20 |

**Response**

```javascript
{
  "arenas": [
    {
      "created_at": "2018-12-21T05:11:21Z",
      "expired_at": "2018-12-21T06:11:21Z",
      "user_id": "8017d200-7870-4b82-b53f-74bae1d2dad7",
      "asset_id": "965e5c6e-434c-3fa9-b780-c50f43cd955c",
      "amount": "100",
      "balance": "100",
      "max_bet": "1",
      "min_gesture": 1,
      "max_gesture": 4,
      "id": "504Md91dXYZW",
      "asset": {
        "asset_id": "965e5c6e-434c-3fa9-b780-c50f43cd955c",
        "chain_id": "43d61dcd-e413-450d-80b8-101d5e903357",
        "name": "Chui Niu Bi",
        "symbol": "CNB",
        "icon_url": "https://images.mixin.one/0sQY63dDMkWTURkJVjowWY6Le4ICjAFuu3ANVyZA4uI3UdkbuOT5fjJUT82ArNYmZvVcxDXyNjxoOv0TAYbQTNKS=s128"
      }
    }
  ],
  "pagination": {
    "has_next": false,
    "next_cursor": ""
  }
}
```

### Get Trending Arenas

```
GET /api/arenas/top
```

**Response**

```javascript
[
  {
    created_at: "2018-12-21T05:11:21Z",
    expired_at: "2018-12-21T06:11:21Z",
    user_id: "8017d200-7870-4b82-b53f-74bae1d2dad7",
    asset_id: "965e5c6e-434c-3fa9-b780-c50f43cd955c",
    amount: "100",
    balance: "100",
    max_bet: "1",
    min_gesture: 1,
    max_gesture: 4,
    id: "504Md91dXYZW",
    asset: {
      asset_id: "965e5c6e-434c-3fa9-b780-c50f43cd955c",
      chain_id: "43d61dcd-e413-450d-80b8-101d5e903357",
      name: "Chui Niu Bi",
      symbol: "CNB",
      icon_url:
        "https://images.mixin.one/0sQY63dDMkWTURkJVjowWY6Le4ICjAFuu3ANVyZA4uI3UdkbuOT5fjJUT82ArNYmZvVcxDXyNjxoOv0TAYbQTNKS=s128"
    }
  }
];
```

### Get Arena Detail

```
GET /api/arena/:id & /api/arena/:trace_id
```

**Response**

```javascript
{
  "created_at": "2018-12-21T05:11:21Z",
  "expired_at": "2018-12-21T06:11:21Z",
  "user_id": "8017d200-7870-4b82-b53f-74bae1d2dad7",
  "asset_id": "965e5c6e-434c-3fa9-b780-c50f43cd955c",
  "amount": "100",
  "balance": "100",
  "max_bet": "1",
  "min_gesture": 1,
  "max_gesture": 4,
  "id": "504Md91dXYZW",
  "asset": {
    "asset_id": "965e5c6e-434c-3fa9-b780-c50f43cd955c",
    "chain_id": "43d61dcd-e413-450d-80b8-101d5e903357",
    "name": "Chui Niu Bi",
    "symbol": "CNB",
    "icon_url": "https://images.mixin.one/0sQY63dDMkWTURkJVjowWY6Le4ICjAFuu3ANVyZA4uI3UdkbuOT5fjJUT82ArNYmZvVcxDXyNjxoOv0TAYbQTNKS=s128"
  }
}
```

### Get Arena Records

```
GET /api/arena/:id/records
```

**Parameters**

| key    | option | desc               |
| ------ | ------ | ------------------ |
| cursor | true   | cursor             |
| limit  | true   | max 50, default 20 |

**Response**

```javascript
{
  "pagination": {
    "has_next": false,
    "next_cursor": ""
  },
  "records": [
    {
      "id": 4,
      "created_at": "2018-12-20T13:14:14Z",
      "snapshot_id": "54b09af9-4fe9-4434-b1ea-4608d481fb32",
      "trace_id": "6c24b895-9098-49ec-8b4d-1b01f64f62a9",
      "asset_id": "965e5c6e-434c-3fa9-b780-c50f43cd955c",
      "amount": "1",
      "result": 2,
      "err": 0,
      "gestures": "1",
      "defend_gesture": "1",
      "reward": "1"
    },
    {
      "id": 3,
      "created_at": "2018-12-20T13:13:02Z",
      "snapshot_id": "60f15bf1-3823-4c8d-b4e0-ff82d61850ba",
      "trace_id": "9720f08f-8c19-47d4-b7a5-d78e8c2b8a75",
      "asset_id": "965e5c6e-434c-3fa9-b780-c50f43cd955c",
      "amount": "1",
      "result": 3,
      "err": 0,
      "gestures": "0",
      "defend_gesture": "2",
      "reward": "1.96"
    },
    {
      "id": 2,
      "created_at": "2018-12-20T13:11:51Z",
      "snapshot_id": "ad829794-6bbe-4b36-8dd9-8d58f8075c7a",
      "trace_id": "30c66666-5277-4167-b3e7-d7087d0ab779",
      "asset_id": "965e5c6e-434c-3fa9-b780-c50f43cd955c",
      "amount": "1",
      "result": 1,
      "err": 0,
      "gestures": "11",
      "defend_gesture": "02",
      "reward": "0"
    },
    {
      "id": 1,
      "created_at": "2018-12-20T13:10:37Z",
      "snapshot_id": "7a4c939a-c5ea-4684-958d-7408191a641a",
      "trace_id": "f5e52ec9-2799-4f43-906d-2e235f16fd60",
      "asset_id": "965e5c6e-434c-3fa9-b780-c50f43cd955c",
      "amount": "1",
      "result": 3,
      "err": 0,
      "gestures": "1",
      "defend_gesture": "0",
      "reward": "1.96"
    }
  ]
}
```

### Get Record Detail

```
GET /api/record/:trace_id
```

**Response**

```javascript
{
    "id": 1,
    "created_at": "2018-12-20T13:10:37Z",
    "snapshot_id": "7a4c939a-c5ea-4684-958d-7408191a641a",
    "trace_id": "f5e52ec9-2799-4f43-906d-2e235f16fd60",
    "asset_id": "965e5c6e-434c-3fa9-b780-c50f43cd955c",
    "amount": "1",
    "result": 3,
    "err": 0,
    "gestures": "1",
    "defend_gesture": "0",
    "reward": "1.96"
}
```

**Record Result**

| code | msg     |
| ---- | ------- |
| 0    | Invalid |
| 1    | Lose    |
| 2    | Draw    |
| 3    | Win     |

**Record Error**

| code | msg                                                  |
| ---- | ---------------------------------------------------- |
| 1    | invalid arena id                                     |
| 2    | asset not match                                      |
| 3    | arena closed                                         |
| 4    | beyond max bet limit                                 |
| 5    | number of gestures out of range                      |
| 6    | arena's balance is not enough to pay possible reawrd |
