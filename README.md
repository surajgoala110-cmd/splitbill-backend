# 割り勘（グループ精算）アプリ — Backend 実装

この workspace には Backend のスターターコードはありません。ここに **Backend API を自由な技術スタックで実装** してください。

## 前提

- API の endpoint・リクエスト/レスポンスの形は **既存フロントエンドのコードが正** です。仕様書はこの workspace にはありません。
- 残高・端数・精算（誰が誰にいくら送るか）の業務ルールについても、正確な仕様は含まれていません。フロントの mock 実装は正しくありません。
- contract と業務ルールの確認方法は、出題画面の問題文を参照してください。

## 必須要件

- `docker compose up` で起動できること
- `http://localhost:8080` で API にアクセスできること
- 使用する言語 / フレームワーク / DB は自由

## 提出物

1. 実装コードの Pull Request
2. `docs/design-decisions.md` — 対話とフロントコードから確定させた API contract・精算ルール・採用した設計と根拠
3. 起動手順を記した `README.md`

ディレクトリ構成・ツール・ライブラリは自由に追加してかまいません。

## 起動手順

Docker がインストールされた環境で、リポジトリのルートから実行します。

```sh
docker compose up --build
```

API は `http://localhost:8080`、ヘルスチェックは `http://localhost:8080/healthz` で利用できます。API の contract、精算ルール、未確認事項は [docs/design-decisions.md](docs/design-decisions.md) に記載しています。
