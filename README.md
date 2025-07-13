# WebpDownloader

指定したURLから数字のみのファイル名を持つwebp画像を自動ダウンロードするGoアプリケーションです。

## 機能

- JavaScriptが使用されているWebサイトに対応（ChromeDP使用）
- IMGタグから数字のみのファイル名を持つwebp画像を検出
- h1タグの内容を基にしたディレクトリ作成
- 重複ファイルの自動リネーム（連番付与）
- HTTPS対応

## インストール

### Go環境が必要です
```bash
go install github.com/kznagamori/go_WebpDownloader@latest
```

### ソースからビルド
```bash
git clone https://github.com/kznagamori/go_WebpDownloader.git
cd go_WebpDownloader
go mod tidy
go build -o go_WebpDownloader
```

## 使用方法

```bash
go_WebpDownloader <URL>
```

### 例
```bash
go_WebpDownloader https://example.com/gallery
```

## 動作条件

- 対象ファイル：webp拡張子
- ファイル名：数字のみ（例：123.webp, 456.webp）
- IMGタグ内の画像
- ダウンロード先：h1タグの内容を基にしたディレクトリ

## 依存関係

- github.com/PuerkitoBio/goquery - HTMLパース
- github.com/chromedp/chromedp - JavaScript対応ブラウザ自動化

## 注意事項

- ChromeDPを使用するため、Chromeブラウザのインストールが必要です
- 大量の画像をダウンロードする際は、対象サイトの利用規約を確認してください
- ネットワーク環境によってはタイムアウトが発生する場合があります

## ライセンス

MIT License
