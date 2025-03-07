# dify-apps-exporter

あなたの Dify アプリを一括でエクスポートします。

## 準備

以下の環境変数を設定してください。

```sh
DIFY_CONSOLE_API=https://{Dify installed host}/console/api
DIFY_EMAIL={your email}
DIFY_PASSWORD={dify console password}
```

### 実行バイナリの準備

利用環境のアーキテクチャにあったバイナリを[リリースページ](https://github.com/kkazuo/dify-apps-exporter/releases)からダウンロードしてください。

またはソースコードをダウンロードして

    go run .

でも実行できます。

## 実行

    dify-apps-exporter

実行したディレクトリに apps.zip というファイルが作成され、すべての Dify App 定義 YAML が格納されています。

## 制限

* コミュニティ版でしかテストしていません。
