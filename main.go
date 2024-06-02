package main

import (
	"database/sql"
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type config struct {
	ChromedpUrl    string
	ScrapeInterval int
	Timeout        int
}

var db *sql.DB

func init() {
	// 连接数据库
	dsn := "root:wulien123@tcp(101.34.126.169:3306)/krillin_ai?charset=utf8mb4&parseTime=True&loc=Local&readTimeout=1s&timeout=1s&writeTimeout=3s"
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
}

type Crawler struct {
	InstanceUrl string
	Timeout     time.Duration
}

func main() {
	defer db.Close()
	var conf config
	confFile, err := os.ReadFile("config.toml")
	if err != nil {
		panic(err)
	}
	if err = toml.Unmarshal(confFile, &conf); err != nil {
		panic(err)
	}
	cr := Crawler{
		InstanceUrl: conf.ChromedpUrl,
		Timeout:     time.Duration(conf.Timeout) * time.Minute,
	}

	for {
		// 爬取量子位数据
		liangziweiArticles, err := cr.scrapeLiangziweiArticles()
		if err != nil {
			fmt.Printf("%s [量子位]数据爬取失败，错误信息: %s\n", time.Now().Format("01-02 15:04:05"), err.Error())
			time.Sleep(time.Minute * time.Duration(conf.ScrapeInterval))
			continue
		}
		fmt.Printf("%s [量子位]数据爬取成功！本次数据量：%d 条\n", time.Now().Format("01-02 15:04:05"), len(liangziweiArticles))
		// 爬取36氪数据
		krArticles, err := cr.scrape36KrArticles()
		if err != nil {
			fmt.Printf("%s [36氪]数据爬取失败，错误信息: %s\n", time.Now().Format("01-02 15:04:05"), err.Error())
			time.Sleep(time.Minute * time.Duration(conf.ScrapeInterval))
			continue
		}
		fmt.Printf("%s [36氪]数据爬取成功！本次数据量：%d 条\n", time.Now().Format("01-02 15:04:05"), len(krArticles))

		// 插入动态数据到数据库
		articles := make([]Article, 0, len(liangziweiArticles)+len(krArticles))
		articles = append(articles, liangziweiArticles...)
		articles = append(articles, krArticles...)
		count, err := insertTrendsData(articles)
		if err != nil {
			fmt.Printf("%s 动态数据插入失败，错误信息: %s\n", time.Now().Format("01-02 15:04:05"), err.Error())
			time.Sleep(time.Minute * time.Duration(conf.ScrapeInterval))
			continue
		}
		fmt.Printf("%s 动态数据插入成功！本次插入条数：%d 条\n", time.Now().Format("01-02 15:04:05"), count)

		// 爬取张小珺数据
		zhangXiaoJun, err := cr.scrapeZhangXiaoJun()
		if err != nil {
			fmt.Printf("%s [张小珺]数据爬取失败，错误信息: %s\n", time.Now().Format("01-02 15:04:05"), err.Error())
			time.Sleep(time.Minute * time.Duration(conf.ScrapeInterval))
			continue
		}
		fmt.Printf("%s [张小珺]数据爬取成功！本次数据量：%d 条\n", time.Now().Format("01-02 15:04:05"), len(zhangXiaoJun))
		personTracks := make([]personTrack, 0, len(zhangXiaoJun))
		personTracks = append(personTracks, zhangXiaoJun...)

		// 插入人物追踪数据
		count, err = insertPersonTracksData(personTracks)
		if err != nil {
			fmt.Printf("%s 人物追踪数据插入失败，错误信息: %s\n", time.Now().Format("01-02 15:04:05"), err.Error())
			time.Sleep(time.Minute * time.Duration(conf.ScrapeInterval))
			continue
		}
		fmt.Printf("%s 人物追踪数据插入成功！本次插入条数：%d 条\n", time.Now().Format("01-02 15:04:05"), count)
		time.Sleep(time.Minute * time.Duration(conf.ScrapeInterval)) // 轮次间隔
	}
}

type Article struct {
	Title      string
	Content    string
	PubTime    int32
	Link       string
	PlatformId int
}

type personTrack struct {
	PersonId       int
	PersonIdInside string
	Content        string
	ImageInfo      string
	VideoInfo      string
	Link           string
	PubTime        int
}

func insertTrendsData(articles []Article) (int, error) {
	count := 0
	for _, article := range articles {
		if article.Title == "" || article.Content == "" || article.PubTime == 0 {
			continue
		}
		// 检查标题是否已存在
		var exists bool
		query := "SELECT EXISTS(SELECT 1 FROM trends_from_crawler WHERE title = ?)"
		err := db.QueryRow(query, article.Title).Scan(&exists)
		if err != nil {
			return 0, err
		}
		if !exists {
			query = "INSERT INTO trends_from_crawler (title, content, platform_id, pub_time,link,type,create_time) VALUES (?, ?, ?, ?,?,?,?)"
			_, err = db.Exec(query, article.Title, article.Content, article.PlatformId, article.PubTime, article.Link, 1, time.Now().Unix())
			if err != nil {
				return 0, err
			}
			count++
		}
	}

	return count, nil
}

func insertPersonTracksData(tracks []personTrack) (int, error) {
	count := 0
	for _, v := range tracks {
		if v.Content == "" || v.ImageInfo == "" || v.VideoInfo == "" || v.Link == "" || v.PubTime == 0 {
			continue
		}
		// 检查标题是否已存在
		var exists bool
		query := "SELECT EXISTS(SELECT 1 FROM person_track WHERE content = ?)"
		err := db.QueryRow(query, v.Content).Scan(&exists)
		if err != nil {
			return 0, err
		}

		// 如果标题不存在，则插入数据
		if !exists {
			query = "INSERT INTO person_track(person_id,person_id_inside, content, image_info,video_info,link, pub_time,create_time) VALUES (?, ?, ?, ?,?,?,?,?)"
			_, err = db.Exec(query, v.PersonId, v.PersonIdInside, v.Content, v.ImageInfo, v.VideoInfo, v.Link, v.PubTime, time.Now().Unix())
			if err != nil {
				return 0, err
			}
			count++
		}
	}

	return count, nil
}
