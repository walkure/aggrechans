# aggrechans

Slackの全Public channelにおける人間の発言を一つのチャンネルに集約して流します。

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

### Subscribe events

#### bot events

- `channel_rename` - チャンネルのリネーム
- `user_change` - ユーザ情報の変更

#### user events

- `message.channels` public channelに流れるメッセージ

## Author

walkure at 3pf.jp

## License

MIT
