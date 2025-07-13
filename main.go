package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("使用方法: WebpDownloader <URL>")
		fmt.Println("例: WebpDownloader https://example.com")
		os.Exit(1)
	}

	url := os.Args[1]
	
	// ChromeDPを使用してJavaScriptレンダリング後のHTMLを取得
	html, err := getRenderedHTML(url)
	if err != nil {
		log.Fatalf("HTMLの取得に失敗しました: %v", err)
	}

	// HTMLをパース
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		log.Fatalf("HTMLのパースに失敗しました: %v", err)
	}

	// h1タグからディレクトリ名を取得
	h1Text := doc.Find("h1").First().Text()
	if h1Text == "" {
		log.Fatal("h1タグが見つかりません")
	}
	
	// ディレクトリ名をサニタイズ
	dirName := sanitizeFilename(h1Text)
	
	// ディレクトリを作成
	if err := os.MkdirAll(dirName, 0755); err != nil {
		log.Fatalf("ディレクトリの作成に失敗しました: %v", err)
	}

	fmt.Printf("ダウンロード先ディレクトリ: %s\n", dirName)

	// 数字のみのファイル名を持つwebp画像を検索
	downloadCount := 0
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if !exists {
			return
		}

		// 相対URLを絶対URLに変換
		if !strings.HasPrefix(src, "http") {
			src = resolveURL(url, src)
		}

		// webpファイルかつ数字のみのファイル名かチェック
		if isValidWebpFile(src) {
			fmt.Printf("ダウンロード中: %s\n", src)
			
			filename := getFilenameFromURL(src)
			filePath := filepath.Join(dirName, filename)
			
			// 重複チェックと名前調整
			adjustedPath := getUniqueFilename(filePath)
			
			if err := downloadFile(src, adjustedPath); err != nil {
				fmt.Printf("エラー: %s のダウンロードに失敗しました: %v\n", src, err)
			} else {
				fmt.Printf("完了: %s\n", adjustedPath)
				downloadCount++
			}
		}
	})

	fmt.Printf("\nダウンロード完了: %d個のファイルをダウンロードしました\n", downloadCount)
}

// ChromeDPを使用してJavaScriptレンダリング後のHTMLを取得
func getRenderedHTML(url string) (string, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var html string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second), // JavaScriptの実行を待つ
		chromedp.OuterHTML("html", &html),
	)

	return html, err
}

// webpファイルかつ数字のみのファイル名かチェック
func isValidWebpFile(url string) bool {
	// URLからファイル名を抽出
	parts := strings.Split(url, "/")
	filename := parts[len(parts)-1]
	
	// クエリパラメータを除去
	if idx := strings.Index(filename, "?"); idx != -1 {
		filename = filename[:idx]
	}

	// .webp拡張子をチェック
	if !strings.HasSuffix(strings.ToLower(filename), ".webp") {
		return false
	}

	// ファイル名（拡張子を除く）が数字のみかチェック
	basename := strings.TrimSuffix(filename, filepath.Ext(filename))
	matched, _ := regexp.MatchString(`^\d+$`, basename)
	
	return matched
}

// URLからファイル名を抽出
func getFilenameFromURL(url string) string {
	parts := strings.Split(url, "/")
	filename := parts[len(parts)-1]
	
	// クエリパラメータを除去
	if idx := strings.Index(filename, "?"); idx != -1 {
		filename = filename[:idx]
	}
	
	return filename
}

// 相対URLを絶対URLに変換
func resolveURL(baseURL, relativeURL string) string {
	if strings.HasPrefix(relativeURL, "//") {
		// プロトコル相対URL
		if strings.HasPrefix(baseURL, "https:") {
			return "https:" + relativeURL
		}
		return "http:" + relativeURL
	}
	
	if strings.HasPrefix(relativeURL, "/") {
		// ドメイン相対URL
		parts := strings.Split(baseURL, "/")
		if len(parts) >= 3 {
			return parts[0] + "//" + parts[2] + relativeURL
		}
	}
	
	// 相対URL
	baseParts := strings.Split(baseURL, "/")
	if len(baseParts) > 0 {
		baseParts = baseParts[:len(baseParts)-1]
		return strings.Join(baseParts, "/") + "/" + relativeURL
	}
	
	return relativeURL
}

// ファイル名をサニタイズ
func sanitizeFilename(filename string) string {
	// 使用できない文字を置換
	reg := regexp.MustCompile(`[<>:"/\\|?*]`)
	return reg.ReplaceAllString(filename, "_")
}

// 重複しないファイル名を生成
func getUniqueFilename(filePath string) string {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return filePath
	}

	dir := filepath.Dir(filePath)
	filename := filepath.Base(filePath)
	ext := filepath.Ext(filename)
	basename := strings.TrimSuffix(filename, ext)

	counter := 1
	for {
		newFilename := fmt.Sprintf("%s_%d%s", basename, counter, ext)
		newPath := filepath.Join(dir, newFilename)
		
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
		counter++
	}
}

// ファイルをダウンロード
func downloadFile(url, filePath string) error {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTPエラー: %s", resp.Status)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}