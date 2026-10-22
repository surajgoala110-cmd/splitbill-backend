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
docker compose up --build -d
```

API は `http://localhost:8080`、ヘルスチェックは `http://localhost:8080/healthz` で利用できます。API の contract、精算ルール、未確認事項は [docs/design-decisions.md](docs/design-decisions.md) に記載しています。

## 動作確認

起動後、まずヘルスチェックを実行します。

```sh
curl -fsS http://localhost:8080/healthz
# {"status":"ok"}
```

次の手順は、グループ作成から端数を含む精算結果の確認までを実行します。`curl` と標準的な Unix ツールだけで実行できるため、`localhost:8080` を対象にした hidden test の前提確認にも利用できます。

```sh
group=$(curl -fsS -X POST http://localhost:8080/api/groups \
	-H 'content-type: application/json' \
	-d '{"name":"Dinner"}')
group_id=$(printf '%s' "$group" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')

a=$(curl -fsS -X POST "http://localhost:8080/api/groups/$group_id/members" \
	-H 'content-type: application/json' -d '{"name":"A"}')
b=$(curl -fsS -X POST "http://localhost:8080/api/groups/$group_id/members" \
	-H 'content-type: application/json' -d '{"name":"B"}')
c=$(curl -fsS -X POST "http://localhost:8080/api/groups/$group_id/members" \
	-H 'content-type: application/json' -d '{"name":"C"}')
a_id=$(printf '%s' "$a" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')
b_id=$(printf '%s' "$b" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')
c_id=$(printf '%s' "$c" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')

curl -fsS -X POST "http://localhost:8080/api/groups/$group_id/expenses" \
	-H 'content-type: application/json' \
	-d "{\"description\":\"Dinner\",\"amount\":100,\"paidBy\":\"$a_id\",\"participants\":[\"$a_id\",\"$b_id\",\"$c_id\"]}"

curl -fsS "http://localhost:8080/api/groups/$group_id/balances"
# [{"memberId":"mem_2","balance":66},{"memberId":"mem_3","balance":-33},{"memberId":"mem_4","balance":-33}]

curl -fsS "http://localhost:8080/api/groups/$group_id/settlements"
# [{"from":"mem_3","to":"mem_2","amount":33},{"from":"mem_4","to":"mem_2","amount":33}]
```

The generated IDs can differ between runs; the important assertions are that the balances sum to zero, the `100` units are split as `34/33/33`, and the settlement amounts total `66`. Stop the local service with:

```sh
docker compose down
```
