package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	_ "github.com/go-sql-driver/mysql"
)

type config struct {
	ChromedpUrl    string
	ScrapeInterval int
}

func main() {
	var conf config
	confFile, err := os.ReadFile("config.toml")
	if err != nil {
		panic(err)
	}
	if err = toml.Unmarshal(confFile, &conf); err != nil {
		panic(err)
	}
	// 初始化Chromedp上下文
	ctx, cancel := chromedp.NewRemoteAllocator(context.Background(), conf.ChromedpUrl)
	defer cancel()
	//ctx, cancel = chromedp.NewExecAllocator(ctx,
	//	append(chromedp.DefaultExecAllocatorOptions[:],
	//		//	chromedp.Flag("headless", false), // 设置为有头模式
	//		chromedp.Flag("enable-automation", false),
	//		chromedp.Flag("disable-blink-features", "AutomationControlled"), //禁用 blink 特征，防检测关键
	//		chromedp.UserAgent(`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36`),
	//	)...,
	//)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	for {
		// 爬取数据
		articles, err := scrapeArticles(ctx)
		if err != nil {
			fmt.Printf("%s 数据爬取失败，错误信息: %s\n", time.Now().Format("01-02 15:04:05"), err.Error())
		}
		fmt.Printf("%s [量子位]数据爬取成功！本次数据量：%d 条\n", time.Now().Format("01-02 15:04:05"), len(articles))

		// 插入数据到数据库
		count, err := insertDataIntoDB(articles)
		if err != nil {
			fmt.Printf("%s [量子位]数据插入失败，错误信息: %s\n", time.Now().Format("01-02 15:04:05"), err.Error())
		}
		fmt.Printf("%s [量子位]数据插入成功！本次插入条数：%d 条\n", time.Now().Format("01-02 15:04:05"), count)
		time.Sleep(time.Minute * time.Duration(conf.ScrapeInterval))
	}

}

type Article struct {
	Title   string
	Content string
	PubTime int32
}

func scrapeArticles(ctx context.Context) ([]Article, error) {
	var articles []Article

	// 访问文章列表页
	if err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.qbitai.com/"),
	); err != nil {
		return nil, err
	}

	// 获取所有文章链接
	var links []string
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`Array.from(document.querySelectorAll('body > div.main.index_page > div.content > div > * > div > a')).map(a => a.href)`, &links),
	); err != nil {
		return nil, err
	}

	for _, link := range links {
		var title, content, date, timeStr string

		// 访问文章链接并提取数据
		if err := chromedp.Run(ctx,
			chromedp.Navigate(link),
			chromedp.Sleep(time.Millisecond*time.Duration(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(100)*10+2000)),
			chromedp.Text("body > div.main > div.content > div.article h1", &title),
			// 获取所有p标签的文本，即正文
			chromedp.Evaluate(`Array.from(document.querySelectorAll('body > div.main > div.content > div.article > p')).map(p => p.textContent).join('\n')`, &content),
			chromedp.Text("body > div.main > div.content > div.article > div.article_info > span.date", &date),
			chromedp.Text("body > div.main > div.content > div.article > div.article_info > span.date + span", &timeStr),
		); err != nil {
			return nil, err
		}

		pubTime, err := parseTime(date, timeStr)
		if err != nil {
			return nil, err
		}
		if len([]rune(content)) > 1000 {
			content = string([]rune(content)[:1000]) // 最多保存一千字，足够满足前端展示需求
		}
		articles = append(articles, Article{
			Title:   title,
			Content: content,
			PubTime: pubTime,
		})
	}

	return articles, nil
}

// 解析文章割裂的日期和时间到秒时间戳
func parseTime(dateStr, timeStr string) (int32, error) {
	// 日期格式为 2024-05-25，时间格式为 21:41:09
	layout := "2006-01-02 15:04:05"
	timeStr = strings.TrimSpace(timeStr)
	dateTimeStr := fmt.Sprintf("%s %s", dateStr, timeStr)
	t, err := time.ParseInLocation(layout, dateTimeStr, time.Local)
	if err != nil {
		return 0, err
	}
	return int32(t.Unix()), nil
}

func insertDataIntoDB(articles []Article) (int, error) {
	// 连接数据库
	dsn := "root:wulien123@tcp(101.34.126.169:3306)/krillin_ai?charset=utf8mb4&parseTime=True&loc=Local&readTimeout=1s&timeout=1s&writeTimeout=3s"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	count := 0
	for _, article := range articles {
		// 检查标题是否已存在
		var exists bool
		query := "SELECT EXISTS(SELECT 1 FROM trends_from_crawler WHERE title = ?)"
		err := db.QueryRow(query, article.Title).Scan(&exists)
		if err != nil {
			return 0, err
		}

		// 如果标题不存在，则插入数据
		if !exists {
			query = "INSERT INTO trends_from_crawler (title, content, platform_id, pub_time) VALUES (?, ?, ?, ?)"
			_, err = db.Exec(query, article.Title, article.Content, 1, article.PubTime)
			if err != nil {
				return 0, err
			}
			count++
		}
	}

	return count, nil
}
