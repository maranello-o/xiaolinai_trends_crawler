package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// 动态-量子位
func (cr Crawler) scrapeLiangziweiArticles(ctx context.Context) ([]Article, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, cr.Timeout)
	defer cancel()
	var articles []Article

	fmt.Printf("%s [量子位]初始化Chromedp上下文成功\n", time.Now().Format("01-02 15:04:05"))
	// 访问文章列表页
	if err := chromedp.Run(timeoutCtx,
		chromedp.Navigate("https://www.qbitai.com/"),
	); err != nil {
		return nil, err
	}
	fmt.Printf("%s [量子位]访问文章列表页成功\n", time.Now().Format("01-02 15:04:05"))

	// 获取所有文章链接
	var links []string
	if err := chromedp.Run(timeoutCtx,
		chromedp.Sleep(time.Millisecond*time.Duration(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(100)*10+2000)),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('body > div.main.index_page > div.content > div > * > div > a')).map(a => a.href)`, &links),
	); err != nil {
		return nil, err
	}
	fmt.Printf("%s [量子位]获取当前页所有文章链接成功,条数：%d 条\n", time.Now().Format("01-02 15:04:05"), len(links))

	for _, link := range links {
		var title, content, date, timeStr string

		// 访问文章链接并提取数据
		if err := chromedp.Run(timeoutCtx,
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

		pubTime, err := parseSeparateTime(date, timeStr)
		if err != nil {
			return nil, err
		}
		if len([]rune(content)) > 500 {
			content = string([]rune(content)[:500]) // 最多保存五百字，足够满足前端展示需求
		}
		articles = append(articles, Article{
			Title:      title,
			Content:    content,
			PubTime:    pubTime,
			Link:       link,
			PlatformId: 1,
		})
		fmt.Printf("%s [量子位]本篇文章数据爬取成功！文章标题：%s\n", time.Now().Format("01-02 15:04:05"), title)
	}

	return articles, nil
}

// 动态-36氪
func (cr Crawler) scrape36KrArticles(ctx context.Context) ([]Article, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, cr.Timeout)
	defer cancel()
	var articles []Article

	fmt.Printf("%s [36氪]初始化Chromedp上下文成功\n", time.Now().Format("01-02 15:04:05"))
	// 访问文章列表页
	if err := chromedp.Run(timeoutCtx,
		chromedp.Navigate("https://36kr.com/search/articles/AI?sort=date"),
	); err != nil {
		return nil, err
	}
	fmt.Printf("%s [36氪]访问文章列表页成功\n", time.Now().Format("01-02 15:04:05"))

	// 获取所有文章链接
	var links []string
	if err := chromedp.Run(timeoutCtx,
		chromedp.Sleep(time.Millisecond*time.Duration(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(100)*10+2000)),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('#app > div > div.kr-layout-main.clearfloat > div.main-right > div > div > div.kr-search-result-list > div.kr-loading-more > ul > li > div > div > div.kr-shadow-content > div.article-item-pic-wrapper > a')).map(a => a.href)`, &links),
	); err != nil {
		return nil, err
	}
	fmt.Printf("%s [36氪]获取当前页所有文章链接成功,条数：%d 条\n", time.Now().Format("01-02 15:04:05"), len(links))

	for _, link := range links {
		var title, content, timeStr string

		// 访问文章链接并提取数据
		if err := chromedp.Run(timeoutCtx,
			chromedp.Navigate(link),
			chromedp.Sleep(time.Millisecond*time.Duration(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(100)*10+2000)),
			chromedp.Text("#app > div > div.box-kr-article-new-y > div > div.kr-layout-main.clearfloat > div.main-right > div > div > div > div.article-detail-wrapper-box > div > div.article-left-container > div.article-content > div > div > div:nth-child(1) > div > h1", &title),
			// 获取所有p标签的文本，即正文
			chromedp.Evaluate(`Array.from(document.querySelectorAll('#app > div > div.box-kr-article-new-y > div > div.kr-layout-main.clearfloat > div.main-right > div > div > div > div.article-detail-wrapper-box > div > div.article-left-container > div.article-content > div > div > div.common-width.margin-bottom-20 > div')).map(p => p.textContent).join('\n')`, &content),
			chromedp.Text("#app > div > div.box-kr-article-new-y > div > div.kr-layout-main.clearfloat > div.main-right > div > div > div > div.article-detail-wrapper-box > div > div.article-left-container > div.article-content > div > div > div:nth-child(1) > div > div.article-title-icon.common-width.margin-bottom-40 > span", &timeStr),
		); err != nil {
			return nil, err
		}
		timeStr = strings.TrimPrefix(timeStr, "·") // 去除前面的特殊字符
		pubTime, err := parseTime(timeStr)
		if err != nil {
			return nil, err
		}
		if len([]rune(content)) > 500 {
			content = string([]rune(content)[:500]) // 最多保存五百字，足够满足前端展示需求
		}
		articles = append(articles, Article{
			Title:      title,
			Content:    content,
			PubTime:    pubTime,
			Link:       link,
			PlatformId: 2,
		})
		fmt.Printf("%s [36氪]本篇文章数据爬取成功！文章标题：%s\n", time.Now().Format("01-02 15:04:05"), title)
	}

	return articles, nil
}

// 人物追踪-张小珺
func (cr Crawler) scrapeZhangXiaoJun(ctx context.Context) ([]personTrack, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, cr.Timeout)
	defer cancel()

	fmt.Printf("%s [张小珺]初始化Chromedp上下文成功\n", time.Now().Format("01-02 15:04:05"))
	// 访问文章列表页
	if err := chromedp.Run(timeoutCtx,
		chromedp.Navigate("https://www.xiaoyuzhoufm.com/podcast/626b46ea9cbbf0451cf5a962"),
	); err != nil {
		return nil, err
	}
	fmt.Printf("%s [张小珺]访问动态列表页成功\n", time.Now().Format("01-02 15:04:05"))
	var nodes []*cdp.Node
	if err := chromedp.Run(timeoutCtx,
		chromedp.Sleep(time.Millisecond*time.Duration(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(100)*10+2000)),
		chromedp.Nodes("#__next > div.jsx-753695559.jsx-677813307 > main > main > div.jsx-7bbe0f84186f1998.tabs-container > ul > *", &nodes),
	); err != nil {
		return nil, err
	}
	tracks := make([]personTrack, len(nodes))
	for k, v := range nodes {
		var content, timeStr, link string
		err := chromedp.Run(timeoutCtx,
			chromedp.Text(v.FullXPath()+"/a/div[contains(@class,\"info\")]/div[contains(@class,\"title\")]", &content),
			chromedp.AttributeValue(v.FullXPath()+"/a/div[contains(@class,\"info\")]/div[contains(@class,\"footer\")]/div[1]/time", "datetime", &timeStr, nil),
			chromedp.AttributeValue(v.FullXPath()+"/a", "href", &link, nil),
		)
		if err != nil {
			return nil, err
		}
		// 处理时间
		t, err := time.ParseInLocation(time.RFC3339, timeStr, time.Local)
		if err != nil {
			fmt.Println("解析时间出错:", err)
			return nil, err
		}
		link = "https://www.xiaoyuzhoufm.com" + link
		content = strings.Split(content, ".")[1]
		tracks[k] = personTrack{
			PersonId:       1,
			PersonIdInside: "",
			Content:        content,
			ImageInfo:      "[]",
			VideoInfo:      "[]",
			Link:           link,
			PubTime:        int(t.Unix()),
		}
		fmt.Printf("%s [张小珺]获取当前动态成功,内容%v，时间：%d\n", time.Now().Format("01-02 15:04:05"), content, t.Unix())
	}

	return tracks, nil
}

func (cr Crawler) scrapeFuSheng() ([]personTrack, error) {
	// 访问文章列表页
	resp, err := http.Get("https://www.douyin.com/aweme/v1/web/aweme/post/?device_platform=webapp&aid=6383&channel=channel_pc_web&sec_user_id=MS4wLjABAAAAAtRQ2UenO2AJ4l0XcBQLek2Tu8Cm2tVm_ZrbF13SI8M&max_cursor=0&locate_query=false&show_live_replay_strategy=1&need_time_list=1&time_list_query=0&whale_cut_token=&cut_version=1&count=18&publish_video_strategy_type=2&update_version_code=170400&pc_client_type=1&version_code=290100&version_name=29.1.0&cookie_enabled=true&screen_width=1707&screen_height=1067&browser_language=zh-CN&browser_platform=Win32&browser_name=Chrome&browser_version=126.0.0.0&browser_online=true&engine_name=Blink&engine_version=126.0.0.0&os_name=Windows&os_version=10&cpu_core_num=16&device_memory=8&platform=PC&downlink=10&effective_type=4g&round_trip_time=50&webid=7352702470689539618&msToken=0flT8cSw1FVJfgHzyTqafvKhfj5PVGv9CHD8lU6eI6HRmhAeIzLe5uV-DrxVd25xUfAJRdJ-siw5bthNMi9vxhGjHKze9j-X52jHp6pAyxDhoZ5LvxNcdg09THw-cpMa&a_bogus=YyWhQmhDDkdkvd6g54QLfY3q6Vl3YDRe0trEMD2fAx3Gm639HMPH9exoEvUvc1RjNs%2FDIeYjy4hjT3BMxQCbA3vIH8WKUIc2QfSkKl5Q5xSSs1XyeykgrUkx4XsAtMa0sv1liQ8kww%2FSSYmmWnAJ5kIlO62-zo0%2F9Wu%3D&verifyFp=verify_lwizw3tf_mb5hS8lF_TDxT_4kdp_9HYY_cdOJNT9gfoj9&fp=verify_lwizw3tf_mb5hS8lF_TDxT_4kdp_9HYY_cdOJNT9gfoj9")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	resByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var res douyinGetVideoListResp
	if err = json.Unmarshal(resByte, &res); err != nil {
		return nil, err
	}
	tracks := make([]personTrack, len(res.AwemeList))
	for k, v := range res.AwemeList {
		tracks[k] = personTrack{
			PersonId:       3,
			PersonIdInside: "78697509",
			Content:        v.Desc,
			ImageInfo:      "[]",
			VideoInfo:      "[{\"title\":\"" + v.Desc + "\",\"url\":\"" + v.Video.Cover.UrlList[0] + "\",}]",
			Link:           "https://www.douyin.com/video/" + v.AwemeId,
			PubTime:        v.CreateTime,
		}
	}
	return tracks, nil
}
