# aggrechans

Slackの全Public channelにおける人間の発言を一つのチャンネルに集約して流したりします。

## Appの権限設定

- アプリを生成して、適当に設定します。
- socket modeを有効にして、eventを有効化します。
- スコープは以下の設定をしてください。
- private channelに発言させるには、private channelでintegration設定で当該アプリを入れる必要があります。

### scopes

#### Bot scope

- `users:read` ユーザIDとユーザ名の変換
- `channels:read` チャンネルIDとチャンネル名の変換
- `chat:write.customize` 名前を変更してpostする権限
  - `chat:write` 親権限
- `team:read` チームURIのドメイン取得

#### User scope

- `channels:history` イベントを受信
- eventsをuser eventsにした場合
  - `channels:read` - チャンネル状態の変化イベント
  - `users:read` - ユーザ情報の変更イベント

### Subscribe events

#### bot(or user) events

- `channel_rename` - チャンネルのリネーム
- `channel_created` - チャンネルの作成
- `channel_unarchive` - チャンネルのアーカイブ解除
- `user_change` - ユーザ情報の変更

#### user events

- `message.channels` public channelに流れるメッセージ

## Message dispatch rules

### 全メッセージを一箇所に集約

環境変数 `AGGREGATE_CHANNEL_ID` にチャンネルIDを指定して起動すると、受信した人間による全発言を投入します。

巨大Workspaceなどでは[chat.postMessage](https://api.slack.com/methods/chat.postMessage)の[API rate limit](https://api.slack.com/docs/rate-limits#tier_t5)を突き抜けるかもしれません。

### チャンネル名で集約先を分ける

環境変数`DISPATCH_CHANNEL`にJSON形式で集約ルールを書くことが出来ます。`prefix`と`suffix`の2つがあり、`prefix`を先に評価します。

```json
[{
  "prefix": "times_",
  "cid": "CIDTIMES"
},{
  "suffix": "_zatsu",
  "cid": "CIDZATSU"
},{
  "suffix": "_foobar",
  "cid": "CIDFOOBAR"
}]
```

## Heroku 

WebhookでEvent API受け取る場合はHerokuでも動きます。

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy)

## Author

walkure at 3pf.jp

## License

MIT
