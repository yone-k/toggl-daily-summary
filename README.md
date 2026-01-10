# toggl-daily-summary

Toggl Track の時間エントリを日付単位で集計し、Markdown でサマリを出力する
CLI です。単日と期間指定に対応します。

## 必要要件

- Go 1.25.5+（ソースからビルドする場合）
- Toggl Track の API トークンと Workspace ID

## インストール

### ビルド済みバイナリ

GitHub Releases から該当アーカイブをダウンロードして展開してください。

macOS Gatekeeper の注意:
ダウンロードしたバイナリが macOS にブロックされた場合は、自己責任で
quarantine 属性を削除できます。

```bash
xattr -d com.apple.quarantine /path/to/toggl-daily-summary
```

この操作は信頼できるバイナリに対してのみ行ってください。
実行の判断と結果については利用者の責任となります。

### ソースからビルド

```bash
go build ./cmd/toggl-daily-summary
```

## 設定

デフォルト設定ファイル:

```
~/.config/toggl-daily-summary/config.json
```

設定例（`config.example.json` 参照）:

```json
{
  "api_token": "YOUR_TOGGL_API_TOKEN",
  "workspace_id": "1234567",
  "base_url": "https://api.track.toggl.com/api/v9"
}
```

環境変数の上書き:

- `TOGGL_API_TOKEN`
- `TOGGL_WORKSPACE_ID`
- `TOGGL_BASE_URL`

## 使い方

```bash
toggl-daily-summary --date 2026-1-10
```

```bash
toggl-daily-summary --from 2026-1-1 --to 2026-1-7 --daily
```

```bash
toggl-daily-summary --date 2026-1-10 --format detail --out summary.md
```

主なフラグ:

- `--date` 対象日（YYYY-M-D。未指定ならローカルの今日）
- `--from` 開始日（YYYY-M-D）
- `--to` 終了日（YYYY-M-D）
- `--daily` 期間指定時に日別で分割
- `--out` 出力先ファイル（未指定なら stdout）
- `--config` 設定ファイル（デフォルト: `~/.config/toggl-daily-summary/config.json`）
- `--workspace` Workspace ID（config/env を上書き）
- `--format` 出力形式（`default` / `detail`）

## 開発

```bash
go run ./cmd/toggl-daily-summary
```

```bash
go test ./...
```
