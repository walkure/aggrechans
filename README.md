# aggrechans

Slackの全Public channelにおける人間の発言を一つのチャンネルに集約して流したりします。

## Appの権限設定

- アプリを生成して、適当に設定します。
- (必要に応じてsocket modeを有効にして)eventを有効化します。
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

## Redis cache

Slackが送ってくるイベントではチャンネルとユーザ名が内部UIDで表記されているため、適当にlookupする必要があります。
オンメモリでキャッシュしていますが、なんかの弾みでプロセスが再起動される毎にAPI叩くのを避けるためにRedisに保存することができます。

`REDIS_TLS_URL`(TLSでの接続URL)、`REDIS_URL`(生TCPでの接続URL)、`REDIS_HOST`(host:port形式)のどれか環境変数で設定すると、Redisをオンメモリキャッシュの裏として使うようになります。

現状Redisに書き込むkeyは特にTTLを設定していないので、適当に`allkeys-lru`などのeviction policyを設定してください。

## Heroku

WebhookでEvent API受け取る場合はHerokuでも動きます。

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy)

## Render

WebhookでEvent API受け取る場合はRenderでも動きます。

[![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy)

Renderではリポジトリからビルドせず、GHCRに公開しているwebhook modeのイメージ`ghcr.io/walkure/aggrechans:latest-webhook`を使います(`render.yaml`で設定済み)。

新しいイメージは自動ではデプロイされません。リリース後に、Renderの管理画面からデプロイしてください。

> 以前に作ったサービスはリポジトリからビルドする設定のままで、そのままだとsocket modeのイメージがビルドされて起動に失敗します。Renderの管理画面でイメージから動かす設定(`ghcr.io/walkure/aggrechans:latest-webhook`)に切り替えてください。

## Docker

dockerコンテナを作るようにしたので、`ghcr.io/walkure/aggrechans:latest`などで取ってくることが出来ます。[Package](https://github.com/walkure/aggrechans/pkgs/container/aggrechans)を参照してください。

socket modeとwebhook modeでイメージが分かれています。どちらも`linux/amd64`、`linux/arm64`、`linux/arm/v6`、`linux/arm/v7`に対応しています。

| mode | タグ |
| --- | --- |
| socket mode | `latest`, `1.1.2` など |
| webhook mode | `latest-webhook`, `1.1.2-webhook` など(末尾に`-webhook`) |

### Dockerでの起動

`SLACK_BOT_TOKEN`はどちらのmodeでも必要です。

socket modeの場合は`SLACK_APP_TOKEN`も必要です。`REDIS_HOST`の存在は任意で、`SLACK_SIGNING_SECRET`は不要です。

`docker run -e SLACK_BOT_TOKEN=(BOT TOKEN) -e REDIS_HOST=localhost:6379 -e AGGREGATE_CHANNEL_ID=(CHANNEL_ID) -e SLACK_APP_TOKEN=(APP TOKEN) ghcr.io/walkure/aggrechans:latest`

webhook modeの場合は`REDIS_HOST`と`SLACK_SIGNING_SECRET`(App CredentialsのSigning Secretにある値)も必要です。待ち受けポートは`PORT`(既定値8080)で変えられます。

`docker run -p 8080:8080 -e SLACK_BOT_TOKEN=(BOT TOKEN) -e SLACK_SIGNING_SECRET=(SIGNING SECRET) -e REDIS_HOST=localhost:6379 -e AGGREGATE_CHANNEL_ID=(CHANNEL_ID) ghcr.io/walkure/aggrechans:latest-webhook`

### イメージのビルド

`--target`でmodeを選びます(指定しなければ`socket`)。

`docker build --target webhook .`

> 以前のイメージは`SLACK_APP_TOKEN`の有無でmodeを自動で切り替えていましたが、`latest`はsocket mode専用になりました。webhook modeで使っている場合は`latest-webhook`などに切り替えてください。

## Author

walkure at 3pf.jp

## License

MIT
