# spanneranime

## Usage

### Manual Run

スペースキーを押して手動でアニメーションを開始します。

```bash
go run ./cmd/main.go
```

### JOIN1

最もシンプルなUser TableとOrder TableをJOINするアニメーションです。

以下のコマンドで実行します。

```bash
go run ./cmd/main.go JOIN1
```

### JOIN2

User TableとOrder Tableが2台ある場合のJOINです。Order TableはUserIDで並び替えてはいないので、効率はあまりよくありません。

以下のコマンドで実行します。

```bash
go run ./cmd/main.go JOIN2
```
