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
	url := "https://www.douyin.com/aweme/v1/web/aweme/post/?device_platform=webapp&aid=6383&channel=channel_pc_web&sec_user_id=MS4wLjABAAAAAtRQ2UenO2AJ4l0XcBQLek2Tu8Cm2tVm_ZrbF13SI8M&max_cursor=0&locate_query=false&show_live_replay_strategy=1&need_time_list=1&time_list_query=0&whale_cut_token=&cut_version=1&count=18&publish_video_strategy_type=2&update_version_code=170400&pc_client_type=1&version_code=290100&version_name=29.1.0&cookie_enabled=true&screen_width=1707&screen_height=1067&browser_language=zh-CN&browser_platform=Win32&browser_name=Chrome&browser_version=126.0.0.0&browser_online=true&engine_name=Blink&engine_version=126.0.0.0&os_name=Windows&os_version=10&cpu_core_num=16&device_memory=8&platform=PC&downlink=10&effective_type=4g&round_trip_time=50&webid=7352702470689539618&msToken=0flT8cSw1FVJfgHzyTqafvKhfj5PVGv9CHD8lU6eI6HRmhAeIzLe5uV-DrxVd25xUfAJRdJ-siw5bthNMi9vxhGjHKze9j-X52jHp6pAyxDhoZ5LvxNcdg09THw-cpMa&a_bogus=YyWhQmhDDkdkvd6g54QLfY3q6Vl3YDRe0trEMD2fAx3Gm639HMPH9exoEvUvc1RjNs%2FDIeYjy4hjT3BMxQCbA3vIH8WKUIc2QfSkKl5Q5xSSs1XyeykgrUkx4XsAtMa0sv1liQ8kww%2FSSYmmWnAJ5kIlO62-zo0%2F9Wu%3D&verifyFp=verify_lwizw3tf_mb5hS8lF_TDxT_4kdp_9HYY_cdOJNT9gfoj9&fp=verify_lwizw3tf_mb5hS8lF_TDxT_4kdp_9HYY_cdOJNT9gfoj9"

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Referer", "https://www.douyin.com/user/MS4wLjABAAAAAtRQ2UenO2AJ4l0XcBQLek2Tu8Cm2tVm_ZrbF13SI8M")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	req.Header.Add("Cookie", "csrf_session_id=0630bbd62e9f88a806b21be3b95f71b6; passport_assist_user=Cj0A5mjsHGzMMRItWq-dI4pq-MnoPzsTZqZWZJfqAqKGS0Pgbv5dPuiomMa2abRyLVGBhfFQPq2TtIGcauhxGkoKPKafyL4bcNaeFdx2A1_FKrBLj3xGGAzUG5HQsS_HCbAVCaaoKxNBn6tSe1DH3K9zmRclk5Y71-H4JT8uCBC3mcgNGImv1lQgASIBA9IOYio%3D; ttwid=1%7CauShNcQ1PHqvepvJhYeH7OaXpvIWWD6Wkf2ejBxvzCQ%7C1711934465%7Cce1f3d03ca6a01753e7fe9856d6cf1d272301e48f00ef6b7d9b5e26528a0b008; bd_ticket_guard_client_web_domain=2; sid_guard=a345e4979658d065956adf2c6156d6d3%7C1711934468%7C5184000%7CFri%2C+31-May-2024+01%3A21%3A08+GMT; douyin.com; device_web_cpu_core=16; device_web_memory_size=8; architecture=amd64; LOGIN_STATUS=1; store-region=cn-js; store-region-src=uid; xg_device_score=7.798873949579832; odin_tt=6585e6c1c7ef3961c6c5844b2dda4055e5726c4ba862355f5801dde11fa1d36ac67963a58a6e6bcecb8dd009ae60826b; passport_fe_beating_status=false; SEARCH_RESULT_LIST_TYPE=%22single%22; s_v_web_id=verify_lwizw3tf_mb5hS8lF_TDxT_4kdp_9HYY_cdOJNT9gfoj9; passport_csrf_token=8a8595a7ba40296fc7677c52765ac95b; passport_csrf_token_default=8a8595a7ba40296fc7677c52765ac95b; dy_swidth=1707; dy_sheight=1067; FORCE_LOGIN=%7B%22videoConsumedRemainSeconds%22%3A180%7D; download_guide=%223%2F20240613%2F0%22; pwa2=%220%7C0%7C3%7C0%22; volume_info=%7B%22isUserMute%22%3Afalse%2C%22isMute%22%3Afalse%2C%22volume%22%3A0.956%7D; strategyABtestKey=%221718342776.951%22; xgplayer_device_id=6840998882; xgplayer_user_id=863629687864; bd_ticket_guard_client_data=eyJiZC10aWNrZXQtZ3VhcmQtdmVyc2lvbiI6MiwiYmQtdGlja2V0LWd1YXJkLWl0ZXJhdGlvbi12ZXJzaW9uIjoxLCJiZC10aWNrZXQtZ3VhcmQtcmVlLXB1YmxpYy1rZXkiOiJCTHNacURhNkRocy96Z2dwZTRqemdPNkQzN3QvMjY4R05uOFlIQnp0K2U5NXovUGJFdk11TW9YYkxOZmlUYk15a0tOaG1wVytNQi9IZXF3V1FuZlVsRXM9IiwiYmQtdGlja2V0LWd1YXJkLXdlYi12ZXJzaW9uIjoxfQ%3D%3D; msToken=5foJAadsmkOTekHJn9TOK5uR_p0f7YytooyVfX7trICffJpuhuE91AnbklaYHsw8gfIdmuDnNOuPqlFLn9lhcl2a2FBm7VtZGqHN2mjqhv4BUhrAtey70N3vYMZIvnLw; __ac_nonce=0666c1bdb0086b462330c; __ac_signature=_02B4Z6wo00f01RIXv9wAAIDAzm4gatozirESN7tAACLsHd2wt.IeUp90yIjd-uO6pjJGykdS6fcmStoWyBNdlw10yP8VAj3n3wgqwPboH3OaRLGhA0kswY0lbdLtLN3vIwo1AcIWoKZ0PLeA06; IsDouyinActive=true; home_can_add_dy_2_desktop=%220%22; stream_recommend_feed_params=%22%7B%5C%22cookie_enabled%5C%22%3Atrue%2C%5C%22screen_width%5C%22%3A1707%2C%5C%22screen_height%5C%22%3A1067%2C%5C%22browser_online%5C%22%3Atrue%2C%5C%22cpu_core_num%5C%22%3A16%2C%5C%22device_memory%5C%22%3A8%2C%5C%22downlink%5C%22%3A10%2C%5C%22effective_type%5C%22%3A%5C%224g%5C%22%2C%5C%22round_trip_time%5C%22%3A50%7D%22")

	resp, err := client.Do(req)
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
		videoInfo, _ := json.Marshal(VideoInfo{
			{
				Title: v.Desc,
				Url:   v.Video.Cover.UrlList[0],
			},
		})
		tracks[k] = personTrack{
			PersonId:       3,
			PersonIdInside: "78697509",
			Content:        v.Desc,
			ImageInfo:      "[]",
			VideoInfo:      string(videoInfo),
			Link:           "https://www.douyin.com/video/" + v.AwemeId,
			PubTime:        v.CreateTime,
		}
	}
	return tracks, nil
}

func (cr Crawler) scrapeLiKaiFu() ([]personTrack, error) {
	// 访问文章列表页
	url := "https://www.douyin.com/aweme/v1/web/aweme/post/?device_platform=webapp&aid=6383&channel=channel_pc_web&sec_user_id=MS4wLjABAAAAOb9-QrTSGKHvcQqVkoggNs4-xuI4CtKZ5I12uf2ZSdSg2cRfxdrNsrxZpSV6rIYE&max_cursor=0&locate_query=false&show_live_replay_strategy=1&need_time_list=1&time_list_query=0&whale_cut_token=&cut_version=1&count=18&publish_video_strategy_type=2&update_version_code=170400&pc_client_type=1&version_code=290100&version_name=29.1.0&cookie_enabled=true&screen_width=1707&screen_height=1067&browser_language=zh-CN&browser_platform=Win32&browser_name=Chrome&browser_version=126.0.0.0&browser_online=true&engine_name=Blink&engine_version=126.0.0.0&os_name=Windows&os_version=10&cpu_core_num=16&device_memory=8&platform=PC&downlink=10&effective_type=4g&round_trip_time=50&webid=7352702470689539618&msToken=9QVAxKsFP7V0wz8qPKwKo19oodLQN6VWvV5N0S4rW4hg67BGmrmhSmqBvFmn4S1j9Z-lW4cxw9BQNIXo8iRNA4_EXUpl9xUg0ofiYFW07XrhJVFNKtsKPg-f2sKb9CA%3D&a_bogus=dvWZBQwkDkdpvfSD54xLfY3q6Ra3YQr00trEMD2fKn3WZg39HMY%2F9exLxCXv7mSjNs%2FDIeEjy4hbY3xhrQcGM1wf9Skw%2F2CZm6T0t-P2so0j53inCgWME0hN-vW3SFqQ-wNAEOsQy75cFbt0W9QamhK4bfebY7Y6i6trnf%3D%3D&verifyFp=verify_lwizw3tf_mb5hS8lF_TDxT_4kdp_9HYY_cdOJNT9gfoj9&fp=verify_lwizw3tf_mb5hS8lF_TDxT_4kdp_9HYY_cdOJNT9gfoj9"
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Referer", "https://www.douyin.com/user/MS4wLjABAAAAAtRQ2UenO2AJ4l0XcBQLek2Tu8Cm2tVm_ZrbF13SI8M")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	req.Header.Add("Cookie", "csrf_session_id=0630bbd62e9f88a806b21be3b95f71b6; passport_assist_user=Cj0A5mjsHGzMMRItWq-dI4pq-MnoPzsTZqZWZJfqAqKGS0Pgbv5dPuiomMa2abRyLVGBhfFQPq2TtIGcauhxGkoKPKafyL4bcNaeFdx2A1_FKrBLj3xGGAzUG5HQsS_HCbAVCaaoKxNBn6tSe1DH3K9zmRclk5Y71-H4JT8uCBC3mcgNGImv1lQgASIBA9IOYio%3D; ttwid=1%7CauShNcQ1PHqvepvJhYeH7OaXpvIWWD6Wkf2ejBxvzCQ%7C1711934465%7Cce1f3d03ca6a01753e7fe9856d6cf1d272301e48f00ef6b7d9b5e26528a0b008; bd_ticket_guard_client_web_domain=2; sid_guard=a345e4979658d065956adf2c6156d6d3%7C1711934468%7C5184000%7CFri%2C+31-May-2024+01%3A21%3A08+GMT; douyin.com; device_web_cpu_core=16; device_web_memory_size=8; architecture=amd64; LOGIN_STATUS=1; store-region=cn-js; store-region-src=uid; xg_device_score=7.798873949579832; odin_tt=6585e6c1c7ef3961c6c5844b2dda4055e5726c4ba862355f5801dde11fa1d36ac67963a58a6e6bcecb8dd009ae60826b; passport_fe_beating_status=false; SEARCH_RESULT_LIST_TYPE=%22single%22; s_v_web_id=verify_lwizw3tf_mb5hS8lF_TDxT_4kdp_9HYY_cdOJNT9gfoj9; passport_csrf_token=8a8595a7ba40296fc7677c52765ac95b; passport_csrf_token_default=8a8595a7ba40296fc7677c52765ac95b; dy_swidth=1707; dy_sheight=1067; FORCE_LOGIN=%7B%22videoConsumedRemainSeconds%22%3A180%7D; download_guide=%223%2F20240613%2F0%22; pwa2=%220%7C0%7C3%7C0%22; volume_info=%7B%22isUserMute%22%3Afalse%2C%22isMute%22%3Afalse%2C%22volume%22%3A0.956%7D; strategyABtestKey=%221718342776.951%22; xgplayer_device_id=6840998882; xgplayer_user_id=863629687864; bd_ticket_guard_client_data=eyJiZC10aWNrZXQtZ3VhcmQtdmVyc2lvbiI6MiwiYmQtdGlja2V0LWd1YXJkLWl0ZXJhdGlvbi12ZXJzaW9uIjoxLCJiZC10aWNrZXQtZ3VhcmQtcmVlLXB1YmxpYy1rZXkiOiJCTHNacURhNkRocy96Z2dwZTRqemdPNkQzN3QvMjY4R05uOFlIQnp0K2U5NXovUGJFdk11TW9YYkxOZmlUYk15a0tOaG1wVytNQi9IZXF3V1FuZlVsRXM9IiwiYmQtdGlja2V0LWd1YXJkLXdlYi12ZXJzaW9uIjoxfQ%3D%3D; msToken=5foJAadsmkOTekHJn9TOK5uR_p0f7YytooyVfX7trICffJpuhuE91AnbklaYHsw8gfIdmuDnNOuPqlFLn9lhcl2a2FBm7VtZGqHN2mjqhv4BUhrAtey70N3vYMZIvnLw; __ac_nonce=0666c1bdb0086b462330c; __ac_signature=_02B4Z6wo00f01RIXv9wAAIDAzm4gatozirESN7tAACLsHd2wt.IeUp90yIjd-uO6pjJGykdS6fcmStoWyBNdlw10yP8VAj3n3wgqwPboH3OaRLGhA0kswY0lbdLtLN3vIwo1AcIWoKZ0PLeA06; IsDouyinActive=true; home_can_add_dy_2_desktop=%220%22; stream_recommend_feed_params=%22%7B%5C%22cookie_enabled%5C%22%3Atrue%2C%5C%22screen_width%5C%22%3A1707%2C%5C%22screen_height%5C%22%3A1067%2C%5C%22browser_online%5C%22%3Atrue%2C%5C%22cpu_core_num%5C%22%3A16%2C%5C%22device_memory%5C%22%3A8%2C%5C%22downlink%5C%22%3A10%2C%5C%22effective_type%5C%22%3A%5C%224g%5C%22%2C%5C%22round_trip_time%5C%22%3A50%7D%22")

	resp, err := client.Do(req)
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
		videoInfo, _ := json.Marshal(VideoInfo{
			{
				Title: v.Desc,
				Url:   v.Video.Cover.UrlList[0],
			},
		})
		tracks[k] = personTrack{
			PersonId:       4,
			PersonIdInside: "83250291547",
			Content:        v.Desc,
			ImageInfo:      "[]",
			VideoInfo:      string(videoInfo),
			Link:           "https://www.douyin.com/video/" + v.AwemeId,
			PubTime:        v.CreateTime,
		}
	}
	return tracks, nil
}

func (cr Crawler) scrapeZhouHongYi() ([]personTrack, error) {
	// 访问文章列表页
	url := "https://www.douyin.com/aweme/v1/web/aweme/post/?device_platform=webapp&aid=6383&channel=channel_pc_web&sec_user_id=MS4wLjABAAAAJ3T5moYwIGWeicRl5wBdfosV7R_dCmIbcmAIVZ_3iLK3aLLrOq9pWQDaZBfU0kpQ&max_cursor=0&locate_query=false&show_live_replay_strategy=1&need_time_list=1&time_list_query=0&whale_cut_token=&cut_version=1&count=18&publish_video_strategy_type=2&update_version_code=170400&pc_client_type=1&version_code=290100&version_name=29.1.0&cookie_enabled=true&screen_width=1707&screen_height=1067&browser_language=zh-CN&browser_platform=Win32&browser_name=Chrome&browser_version=126.0.0.0&browser_online=true&engine_name=Blink&engine_version=126.0.0.0&os_name=Windows&os_version=10&cpu_core_num=16&device_memory=8&platform=PC&downlink=6.5&effective_type=4g&round_trip_time=150&webid=7352702470689539618&verifyFp=verify_lwizw3tf_mb5hS8lF_TDxT_4kdp_9HYY_cdOJNT9gfoj9&fp=verify_lwizw3tf_mb5hS8lF_TDxT_4kdp_9HYY_cdOJNT9gfoj9&msToken=i21OeOtWQzDzU9IYf3bA2-QHDTK8crFl-LhjjfDHJlx2yD-16NM-ZkRqiGW2IRYKono9YbDmjNtrpSOGnBNNS0kwonR4kix8kpo3sutegwjw7xxLIkQYam1p_wfsTUo%3D&a_bogus=Ej80%2F5LDdDIkXDyf5AdLfY3q6vP3YQJM0trEMD2fsn3WZ639HMYF9exLS9Uv75yjNs%2FDIeEjy4hbY3cZrQcGM1wf9Skw%2F2CZm6T0t-P2so0j53inCgWME0hN-vW3SFqQ-wNAEOsQy75cFRw0W9QamhK4bfebY7Y6i6trjE%3D%3D"
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Referer", "https://www.douyin.com/user/MS4wLjABAAAAJ3T5moYwIGWeicRl5wBdfosV7R_dCmIbcmAIVZ_3iLK3aLLrOq9pWQDaZBfU0kpQ")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	req.Header.Add("Cookie", "csrf_session_id=0630bbd62e9f88a806b21be3b95f71b6; passport_assist_user=Cj0A5mjsHGzMMRItWq-dI4pq-MnoPzsTZqZWZJfqAqKGS0Pgbv5dPuiomMa2abRyLVGBhfFQPq2TtIGcauhxGkoKPKafyL4bcNaeFdx2A1_FKrBLj3xGGAzUG5HQsS_HCbAVCaaoKxNBn6tSe1DH3K9zmRclk5Y71-H4JT8uCBC3mcgNGImv1lQgASIBA9IOYio%3D; ttwid=1%7CauShNcQ1PHqvepvJhYeH7OaXpvIWWD6Wkf2ejBxvzCQ%7C1711934465%7Cce1f3d03ca6a01753e7fe9856d6cf1d272301e48f00ef6b7d9b5e26528a0b008; bd_ticket_guard_client_web_domain=2; sid_guard=a345e4979658d065956adf2c6156d6d3%7C1711934468%7C5184000%7CFri%2C+31-May-2024+01%3A21%3A08+GMT; douyin.com; device_web_cpu_core=16; device_web_memory_size=8; architecture=amd64; LOGIN_STATUS=1; store-region=cn-js; store-region-src=uid; xg_device_score=7.798873949579832; odin_tt=6585e6c1c7ef3961c6c5844b2dda4055e5726c4ba862355f5801dde11fa1d36ac67963a58a6e6bcecb8dd009ae60826b; passport_fe_beating_status=false; SEARCH_RESULT_LIST_TYPE=%22single%22; s_v_web_id=verify_lwizw3tf_mb5hS8lF_TDxT_4kdp_9HYY_cdOJNT9gfoj9; passport_csrf_token=8a8595a7ba40296fc7677c52765ac95b; passport_csrf_token_default=8a8595a7ba40296fc7677c52765ac95b; dy_swidth=1707; dy_sheight=1067; FORCE_LOGIN=%7B%22videoConsumedRemainSeconds%22%3A180%7D; download_guide=%223%2F20240613%2F0%22; pwa2=%220%7C0%7C3%7C0%22; volume_info=%7B%22isUserMute%22%3Afalse%2C%22isMute%22%3Afalse%2C%22volume%22%3A0.956%7D; strategyABtestKey=%221718342776.951%22; xgplayer_device_id=6840998882; xgplayer_user_id=863629687864; bd_ticket_guard_client_data=eyJiZC10aWNrZXQtZ3VhcmQtdmVyc2lvbiI6MiwiYmQtdGlja2V0LWd1YXJkLWl0ZXJhdGlvbi12ZXJzaW9uIjoxLCJiZC10aWNrZXQtZ3VhcmQtcmVlLXB1YmxpYy1rZXkiOiJCTHNacURhNkRocy96Z2dwZTRqemdPNkQzN3QvMjY4R05uOFlIQnp0K2U5NXovUGJFdk11TW9YYkxOZmlUYk15a0tOaG1wVytNQi9IZXF3V1FuZlVsRXM9IiwiYmQtdGlja2V0LWd1YXJkLXdlYi12ZXJzaW9uIjoxfQ%3D%3D; msToken=5foJAadsmkOTekHJn9TOK5uR_p0f7YytooyVfX7trICffJpuhuE91AnbklaYHsw8gfIdmuDnNOuPqlFLn9lhcl2a2FBm7VtZGqHN2mjqhv4BUhrAtey70N3vYMZIvnLw; __ac_nonce=0666c1bdb0086b462330c; __ac_signature=_02B4Z6wo00f01RIXv9wAAIDAzm4gatozirESN7tAACLsHd2wt.IeUp90yIjd-uO6pjJGykdS6fcmStoWyBNdlw10yP8VAj3n3wgqwPboH3OaRLGhA0kswY0lbdLtLN3vIwo1AcIWoKZ0PLeA06; IsDouyinActive=true; home_can_add_dy_2_desktop=%220%22; stream_recommend_feed_params=%22%7B%5C%22cookie_enabled%5C%22%3Atrue%2C%5C%22screen_width%5C%22%3A1707%2C%5C%22screen_height%5C%22%3A1067%2C%5C%22browser_online%5C%22%3Atrue%2C%5C%22cpu_core_num%5C%22%3A16%2C%5C%22device_memory%5C%22%3A8%2C%5C%22downlink%5C%22%3A10%2C%5C%22effective_type%5C%22%3A%5C%224g%5C%22%2C%5C%22round_trip_time%5C%22%3A50%7D%22")

	resp, err := client.Do(req)
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
		videoInfo, _ := json.Marshal(VideoInfo{
			{
				Title: v.Desc,
				Url:   v.Video.Cover.UrlList[0],
			},
		})
		tracks[k] = personTrack{
			PersonId:       5,
			PersonIdInside: "83250291547",
			Content:        v.Desc,
			ImageInfo:      "[]",
			VideoInfo:      string(videoInfo),
			Link:           "https://www.douyin.com/video/" + v.AwemeId,
			PubTime:        v.CreateTime,
		}
	}
	return tracks, nil
}

func (cr Crawler) scrapeALiYun() ([]personTrack, error) {
	// 访问文章列表页
	url := "https://www.douyin.com/aweme/v1/web/aweme/post/?device_platform=webapp&aid=6383&channel=channel_pc_web&sec_user_id=MS4wLjABAAAA8g9JrmDHhbY-w9tevgXrEbMGqUNNlmIaNTg7tYJTjA8&max_cursor=0&locate_query=false&show_live_replay_strategy=1&need_time_list=1&time_list_query=0&whale_cut_token=&cut_version=1&count=18&publish_video_strategy_type=2&update_version_code=170400&pc_client_type=1&version_code=290100&version_name=29.1.0&cookie_enabled=true&screen_width=1707&screen_height=1067&browser_language=zh-CN&browser_platform=Win32&browser_name=Chrome&browser_version=126.0.0.0&browser_online=true&engine_name=Blink&engine_version=126.0.0.0&os_name=Windows&os_version=10&cpu_core_num=16&device_memory=8&platform=PC&downlink=10&effective_type=4g&round_trip_time=50&webid=7352702470689539618&verifyFp=verify_lwizw3tf_mb5hS8lF_TDxT_4kdp_9HYY_cdOJNT9gfoj9&fp=verify_lwizw3tf_mb5hS8lF_TDxT_4kdp_9HYY_cdOJNT9gfoj9&msToken=6tqq4WmvIrgvsgksVyYiWTbQvpxNDs2SOsmuIXRQ2_a8Op3adU1y4kdNUgUa_ck5QlTA12ixPlEYPoO1JCGMMAsTMkhFc2RcLjZVpbsWPrpERrnyGeJR1GkN19ghDMw5&a_bogus=DjWZ%2F5hkdEgkvVWg5AdLfY3q6v-3YD400trEMD2f0V3GNL39HMTR9exoS9Xvvz6jNs%2FDIeEjy4hbY3cZrQcGM1wf9Skw%2F2CZm6T0t-P2so0j53inCgWME0hN-vW3SFqQ-wNAEOsQy75cFRw0W9QamhK4bfebY7Y6i6tr1E%3D%3D"
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Referer", "https://www.douyin.com/user/MS4wLjABAAAA8g9JrmDHhbY-w9tevgXrEbMGqUNNlmIaNTg7tYJTjA8")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	req.Header.Add("Cookie", "csrf_session_id=0630bbd62e9f88a806b21be3b95f71b6; passport_assist_user=Cj0A5mjsHGzMMRItWq-dI4pq-MnoPzsTZqZWZJfqAqKGS0Pgbv5dPuiomMa2abRyLVGBhfFQPq2TtIGcauhxGkoKPKafyL4bcNaeFdx2A1_FKrBLj3xGGAzUG5HQsS_HCbAVCaaoKxNBn6tSe1DH3K9zmRclk5Y71-H4JT8uCBC3mcgNGImv1lQgASIBA9IOYio%3D; ttwid=1%7CauShNcQ1PHqvepvJhYeH7OaXpvIWWD6Wkf2ejBxvzCQ%7C1711934465%7Cce1f3d03ca6a01753e7fe9856d6cf1d272301e48f00ef6b7d9b5e26528a0b008; bd_ticket_guard_client_web_domain=2; sid_guard=a345e4979658d065956adf2c6156d6d3%7C1711934468%7C5184000%7CFri%2C+31-May-2024+01%3A21%3A08+GMT; douyin.com; device_web_cpu_core=16; device_web_memory_size=8; architecture=amd64; LOGIN_STATUS=1; store-region=cn-js; store-region-src=uid; xg_device_score=7.798873949579832; odin_tt=6585e6c1c7ef3961c6c5844b2dda4055e5726c4ba862355f5801dde11fa1d36ac67963a58a6e6bcecb8dd009ae60826b; passport_fe_beating_status=false; SEARCH_RESULT_LIST_TYPE=%22single%22; s_v_web_id=verify_lwizw3tf_mb5hS8lF_TDxT_4kdp_9HYY_cdOJNT9gfoj9; passport_csrf_token=8a8595a7ba40296fc7677c52765ac95b; passport_csrf_token_default=8a8595a7ba40296fc7677c52765ac95b; dy_swidth=1707; dy_sheight=1067; FORCE_LOGIN=%7B%22videoConsumedRemainSeconds%22%3A180%7D; download_guide=%223%2F20240613%2F0%22; pwa2=%220%7C0%7C3%7C0%22; volume_info=%7B%22isUserMute%22%3Afalse%2C%22isMute%22%3Afalse%2C%22volume%22%3A0.956%7D; strategyABtestKey=%221718342776.951%22; xgplayer_device_id=6840998882; xgplayer_user_id=863629687864; bd_ticket_guard_client_data=eyJiZC10aWNrZXQtZ3VhcmQtdmVyc2lvbiI6MiwiYmQtdGlja2V0LWd1YXJkLWl0ZXJhdGlvbi12ZXJzaW9uIjoxLCJiZC10aWNrZXQtZ3VhcmQtcmVlLXB1YmxpYy1rZXkiOiJCTHNacURhNkRocy96Z2dwZTRqemdPNkQzN3QvMjY4R05uOFlIQnp0K2U5NXovUGJFdk11TW9YYkxOZmlUYk15a0tOaG1wVytNQi9IZXF3V1FuZlVsRXM9IiwiYmQtdGlja2V0LWd1YXJkLXdlYi12ZXJzaW9uIjoxfQ%3D%3D; msToken=5foJAadsmkOTekHJn9TOK5uR_p0f7YytooyVfX7trICffJpuhuE91AnbklaYHsw8gfIdmuDnNOuPqlFLn9lhcl2a2FBm7VtZGqHN2mjqhv4BUhrAtey70N3vYMZIvnLw; __ac_nonce=0666c1bdb0086b462330c; __ac_signature=_02B4Z6wo00f01RIXv9wAAIDAzm4gatozirESN7tAACLsHd2wt.IeUp90yIjd-uO6pjJGykdS6fcmStoWyBNdlw10yP8VAj3n3wgqwPboH3OaRLGhA0kswY0lbdLtLN3vIwo1AcIWoKZ0PLeA06; IsDouyinActive=true; home_can_add_dy_2_desktop=%220%22; stream_recommend_feed_params=%22%7B%5C%22cookie_enabled%5C%22%3Atrue%2C%5C%22screen_width%5C%22%3A1707%2C%5C%22screen_height%5C%22%3A1067%2C%5C%22browser_online%5C%22%3Atrue%2C%5C%22cpu_core_num%5C%22%3A16%2C%5C%22device_memory%5C%22%3A8%2C%5C%22downlink%5C%22%3A10%2C%5C%22effective_type%5C%22%3A%5C%224g%5C%22%2C%5C%22round_trip_time%5C%22%3A50%7D%22")

	resp, err := client.Do(req)
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
		videoInfo, _ := json.Marshal(VideoInfo{
			{
				Title: v.Desc,
				Url:   v.Video.Cover.UrlList[0],
			},
		})
		tracks[k] = personTrack{
			PersonId:       6,
			PersonIdInside: "83250291547",
			Content:        v.Desc,
			ImageInfo:      "[]",
			VideoInfo:      string(videoInfo),
			Link:           "https://www.douyin.com/video/" + v.AwemeId,
			PubTime:        v.CreateTime,
		}
	}
	return tracks, nil
}

func (cr Crawler) scrapeAppleGuanFang() ([]personTrack, error) {
	// 访问文章列表页
	url := "https://www.douyin.com/aweme/v1/web/aweme/post/?device_platform=webapp&aid=6383&channel=channel_pc_web&sec_user_id=MS4wLjABAAAAsFK8qDk2oXQ4Wnjt_HkPczoab32qhAK5z3Gn8pWIzDp_toc1YhMef4eaDYbPLDrR&max_cursor=0&locate_query=false&show_live_replay_strategy=1&need_time_list=1&time_list_query=0&whale_cut_token=&cut_version=1&count=18&publish_video_strategy_type=2&update_version_code=170400&pc_client_type=1&version_code=290100&version_name=29.1.0&cookie_enabled=true&screen_width=1707&screen_height=1067&browser_language=zh-CN&browser_platform=Win32&browser_name=Chrome&browser_version=126.0.0.0&browser_online=true&engine_name=Blink&engine_version=126.0.0.0&os_name=Windows&os_version=10&cpu_core_num=16&device_memory=8&platform=PC&downlink=10&effective_type=4g&round_trip_time=50&webid=7352702470689539618&verifyFp=verify_lwizw3tf_mb5hS8lF_TDxT_4kdp_9HYY_cdOJNT9gfoj9&fp=verify_lwizw3tf_mb5hS8lF_TDxT_4kdp_9HYY_cdOJNT9gfoj9&msToken=MdCGZ2folWa0bEMi8EKASf1mSW74v3yv8gVXl_IBbJDRqmP5q1epVBUDAbnSFWhZ88q5c85191gmXuTczI5lHAy1Yk-H6WGJKuS_6Ehvqh1sjBBWNlFEriFLf5FsHvmV&a_bogus=Df80MQ8hdiVkvd6D5AdLfY3q6vH3YQjr0trEMD2fAV3W7639HMOC9exLSJzvZ3SjNs%2FDIeEjy4hbY3cZrQcGM1wf9Skw%2F2CZm6T0t-P2so0j53inCgWME0hN-vW3SFqQ-wNAEOsQy75cFRw0W9QamhK4bfebY7Y6i6troD%3D%3D"
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Referer", "https://www.douyin.com/user/MS4wLjABAAAAsFK8qDk2oXQ4Wnjt_HkPczoab32qhAK5z3Gn8pWIzDp_toc1YhMef4eaDYbPLDrR")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	req.Header.Add("Cookie", "csrf_session_id=0630bbd62e9f88a806b21be3b95f71b6; passport_assist_user=Cj0A5mjsHGzMMRItWq-dI4pq-MnoPzsTZqZWZJfqAqKGS0Pgbv5dPuiomMa2abRyLVGBhfFQPq2TtIGcauhxGkoKPKafyL4bcNaeFdx2A1_FKrBLj3xGGAzUG5HQsS_HCbAVCaaoKxNBn6tSe1DH3K9zmRclk5Y71-H4JT8uCBC3mcgNGImv1lQgASIBA9IOYio%3D; ttwid=1%7CauShNcQ1PHqvepvJhYeH7OaXpvIWWD6Wkf2ejBxvzCQ%7C1711934465%7Cce1f3d03ca6a01753e7fe9856d6cf1d272301e48f00ef6b7d9b5e26528a0b008; bd_ticket_guard_client_web_domain=2; sid_guard=a345e4979658d065956adf2c6156d6d3%7C1711934468%7C5184000%7CFri%2C+31-May-2024+01%3A21%3A08+GMT; douyin.com; device_web_cpu_core=16; device_web_memory_size=8; architecture=amd64; LOGIN_STATUS=1; store-region=cn-js; store-region-src=uid; xg_device_score=7.798873949579832; odin_tt=6585e6c1c7ef3961c6c5844b2dda4055e5726c4ba862355f5801dde11fa1d36ac67963a58a6e6bcecb8dd009ae60826b; passport_fe_beating_status=false; SEARCH_RESULT_LIST_TYPE=%22single%22; s_v_web_id=verify_lwizw3tf_mb5hS8lF_TDxT_4kdp_9HYY_cdOJNT9gfoj9; passport_csrf_token=8a8595a7ba40296fc7677c52765ac95b; passport_csrf_token_default=8a8595a7ba40296fc7677c52765ac95b; dy_swidth=1707; dy_sheight=1067; FORCE_LOGIN=%7B%22videoConsumedRemainSeconds%22%3A180%7D; download_guide=%223%2F20240613%2F0%22; pwa2=%220%7C0%7C3%7C0%22; volume_info=%7B%22isUserMute%22%3Afalse%2C%22isMute%22%3Afalse%2C%22volume%22%3A0.956%7D; strategyABtestKey=%221718342776.951%22; xgplayer_device_id=6840998882; xgplayer_user_id=863629687864; bd_ticket_guard_client_data=eyJiZC10aWNrZXQtZ3VhcmQtdmVyc2lvbiI6MiwiYmQtdGlja2V0LWd1YXJkLWl0ZXJhdGlvbi12ZXJzaW9uIjoxLCJiZC10aWNrZXQtZ3VhcmQtcmVlLXB1YmxpYy1rZXkiOiJCTHNacURhNkRocy96Z2dwZTRqemdPNkQzN3QvMjY4R05uOFlIQnp0K2U5NXovUGJFdk11TW9YYkxOZmlUYk15a0tOaG1wVytNQi9IZXF3V1FuZlVsRXM9IiwiYmQtdGlja2V0LWd1YXJkLXdlYi12ZXJzaW9uIjoxfQ%3D%3D; msToken=5foJAadsmkOTekHJn9TOK5uR_p0f7YytooyVfX7trICffJpuhuE91AnbklaYHsw8gfIdmuDnNOuPqlFLn9lhcl2a2FBm7VtZGqHN2mjqhv4BUhrAtey70N3vYMZIvnLw; __ac_nonce=0666c1bdb0086b462330c; __ac_signature=_02B4Z6wo00f01RIXv9wAAIDAzm4gatozirESN7tAACLsHd2wt.IeUp90yIjd-uO6pjJGykdS6fcmStoWyBNdlw10yP8VAj3n3wgqwPboH3OaRLGhA0kswY0lbdLtLN3vIwo1AcIWoKZ0PLeA06; IsDouyinActive=true; home_can_add_dy_2_desktop=%220%22; stream_recommend_feed_params=%22%7B%5C%22cookie_enabled%5C%22%3Atrue%2C%5C%22screen_width%5C%22%3A1707%2C%5C%22screen_height%5C%22%3A1067%2C%5C%22browser_online%5C%22%3Atrue%2C%5C%22cpu_core_num%5C%22%3A16%2C%5C%22device_memory%5C%22%3A8%2C%5C%22downlink%5C%22%3A10%2C%5C%22effective_type%5C%22%3A%5C%224g%5C%22%2C%5C%22round_trip_time%5C%22%3A50%7D%22")

	resp, err := client.Do(req)
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
		videoInfo, _ := json.Marshal(VideoInfo{
			{
				Title: v.Desc,
				Url:   v.Video.Cover.UrlList[0],
			},
		})
		tracks[k] = personTrack{
			PersonId:       7,
			PersonIdInside: "83250291547",
			Content:        v.Desc,
			ImageInfo:      "[]",
			VideoInfo:      string(videoInfo),
			Link:           "https://www.douyin.com/video/" + v.AwemeId,
			PubTime:        v.CreateTime,
		}
	}
	return tracks, nil
}

func (cr Crawler) scrapeDingDing() ([]personTrack, error) {
	// 访问文章列表页
	url := "https://www.douyin.com/aweme/v1/web/aweme/post/?device_platform=webapp&aid=6383&channel=channel_pc_web&sec_user_id=MS4wLjABAAAAdnwM98hf2G2ivOlk9sc0VIAywdu2hLry1LCwGrm9X-w&max_cursor=0&locate_query=false&show_live_replay_strategy=1&need_time_list=1&time_list_query=0&whale_cut_token=&cut_version=1&count=18&publish_video_strategy_type=2&update_version_code=170400&pc_client_type=1&version_code=290100&version_name=29.1.0&cookie_enabled=true&screen_width=1707&screen_height=1067&browser_language=zh-CN&browser_platform=Win32&browser_name=Chrome&browser_version=126.0.0.0&browser_online=true&engine_name=Blink&engine_version=126.0.0.0&os_name=Windows&os_version=10&cpu_core_num=16&device_memory=8&platform=PC&downlink=10&effective_type=4g&round_trip_time=50&webid=7352702470689539618&verifyFp=verify_lwizw3tf_mb5hS8lF_TDxT_4kdp_9HYY_cdOJNT9gfoj9&fp=verify_lwizw3tf_mb5hS8lF_TDxT_4kdp_9HYY_cdOJNT9gfoj9&msToken=ubZGkZ6ax46dI3yNDkseZvQKxJhoQgFM5FLflVYpSI7grCUhRZ_B3aslNak4efH1DayjgwCqwMH-o70CWscKBGU9PfFrBQXzGO4apIjvTUxOEC39YiXVmpnjFgaLDuXC&a_bogus=dJmwBmzvdkIB6fSh5AdLfY3q6IH3YQjr0trEMD2fon3W5g39HMOH9exLSJzvYpRjNs%2FDIeEjy4hbY3cZrQcGM1wf9Skw%2F2CZm6T0t-P2so0j53inCgWME0hN-vW3SFqQ-wNAEOsQy75cFRw0W9QamhK4bfebY7Y6i6trdE%3D%3D"
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Referer", "https://www.douyin.com/user/MS4wLjABAAAAdnwM98hf2G2ivOlk9sc0VIAywdu2hLry1LCwGrm9X-w")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	req.Header.Add("Cookie", "csrf_session_id=0630bbd62e9f88a806b21be3b95f71b6; passport_assist_user=Cj0A5mjsHGzMMRItWq-dI4pq-MnoPzsTZqZWZJfqAqKGS0Pgbv5dPuiomMa2abRyLVGBhfFQPq2TtIGcauhxGkoKPKafyL4bcNaeFdx2A1_FKrBLj3xGGAzUG5HQsS_HCbAVCaaoKxNBn6tSe1DH3K9zmRclk5Y71-H4JT8uCBC3mcgNGImv1lQgASIBA9IOYio%3D; ttwid=1%7CauShNcQ1PHqvepvJhYeH7OaXpvIWWD6Wkf2ejBxvzCQ%7C1711934465%7Cce1f3d03ca6a01753e7fe9856d6cf1d272301e48f00ef6b7d9b5e26528a0b008; bd_ticket_guard_client_web_domain=2; sid_guard=a345e4979658d065956adf2c6156d6d3%7C1711934468%7C5184000%7CFri%2C+31-May-2024+01%3A21%3A08+GMT; douyin.com; device_web_cpu_core=16; device_web_memory_size=8; architecture=amd64; LOGIN_STATUS=1; store-region=cn-js; store-region-src=uid; xg_device_score=7.798873949579832; odin_tt=6585e6c1c7ef3961c6c5844b2dda4055e5726c4ba862355f5801dde11fa1d36ac67963a58a6e6bcecb8dd009ae60826b; passport_fe_beating_status=false; SEARCH_RESULT_LIST_TYPE=%22single%22; s_v_web_id=verify_lwizw3tf_mb5hS8lF_TDxT_4kdp_9HYY_cdOJNT9gfoj9; passport_csrf_token=8a8595a7ba40296fc7677c52765ac95b; passport_csrf_token_default=8a8595a7ba40296fc7677c52765ac95b; dy_swidth=1707; dy_sheight=1067; FORCE_LOGIN=%7B%22videoConsumedRemainSeconds%22%3A180%7D; download_guide=%223%2F20240613%2F0%22; pwa2=%220%7C0%7C3%7C0%22; volume_info=%7B%22isUserMute%22%3Afalse%2C%22isMute%22%3Afalse%2C%22volume%22%3A0.956%7D; strategyABtestKey=%221718342776.951%22; xgplayer_device_id=6840998882; xgplayer_user_id=863629687864; bd_ticket_guard_client_data=eyJiZC10aWNrZXQtZ3VhcmQtdmVyc2lvbiI6MiwiYmQtdGlja2V0LWd1YXJkLWl0ZXJhdGlvbi12ZXJzaW9uIjoxLCJiZC10aWNrZXQtZ3VhcmQtcmVlLXB1YmxpYy1rZXkiOiJCTHNacURhNkRocy96Z2dwZTRqemdPNkQzN3QvMjY4R05uOFlIQnp0K2U5NXovUGJFdk11TW9YYkxOZmlUYk15a0tOaG1wVytNQi9IZXF3V1FuZlVsRXM9IiwiYmQtdGlja2V0LWd1YXJkLXdlYi12ZXJzaW9uIjoxfQ%3D%3D; msToken=5foJAadsmkOTekHJn9TOK5uR_p0f7YytooyVfX7trICffJpuhuE91AnbklaYHsw8gfIdmuDnNOuPqlFLn9lhcl2a2FBm7VtZGqHN2mjqhv4BUhrAtey70N3vYMZIvnLw; __ac_nonce=0666c1bdb0086b462330c; __ac_signature=_02B4Z6wo00f01RIXv9wAAIDAzm4gatozirESN7tAACLsHd2wt.IeUp90yIjd-uO6pjJGykdS6fcmStoWyBNdlw10yP8VAj3n3wgqwPboH3OaRLGhA0kswY0lbdLtLN3vIwo1AcIWoKZ0PLeA06; IsDouyinActive=true; home_can_add_dy_2_desktop=%220%22; stream_recommend_feed_params=%22%7B%5C%22cookie_enabled%5C%22%3Atrue%2C%5C%22screen_width%5C%22%3A1707%2C%5C%22screen_height%5C%22%3A1067%2C%5C%22browser_online%5C%22%3Atrue%2C%5C%22cpu_core_num%5C%22%3A16%2C%5C%22device_memory%5C%22%3A8%2C%5C%22downlink%5C%22%3A10%2C%5C%22effective_type%5C%22%3A%5C%224g%5C%22%2C%5C%22round_trip_time%5C%22%3A50%7D%22")

	resp, err := client.Do(req)
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
		videoInfo, _ := json.Marshal(VideoInfo{
			{
				Title: v.Desc,
				Url:   v.Video.Cover.UrlList[0],
			},
		})
		tracks[k] = personTrack{
			PersonId:       8,
			PersonIdInside: "83250291547",
			Content:        v.Desc,
			ImageInfo:      "[]",
			VideoInfo:      string(videoInfo),
			Link:           "https://www.douyin.com/video/" + v.AwemeId,
			PubTime:        v.CreateTime,
		}
	}
	return tracks, nil
}

func (cr Crawler) scrapeDuJiaJianJi() ([]personTrack, error) {
	// 访问文章列表页
	url := "https://www.douyin.com/aweme/v1/web/aweme/post/?device_platform=webapp&aid=6383&channel=channel_pc_web&sec_user_id=MS4wLjABAAAAp2jMqtHQRKBfj1dAn8LdvFCxsJnBreZ1a8NU3zmpyDjf8rIyIdr1KhGsOTyA5oZA&max_cursor=0&locate_query=false&show_live_replay_strategy=1&need_time_list=1&time_list_query=0&whale_cut_token=&cut_version=1&count=18&publish_video_strategy_type=2&update_version_code=170400&pc_client_type=1&version_code=290100&version_name=29.1.0&cookie_enabled=true&screen_width=1707&screen_height=1067&browser_language=zh-CN&browser_platform=Win32&browser_name=Chrome&browser_version=126.0.0.0&browser_online=true&engine_name=Blink&engine_version=126.0.0.0&os_name=Windows&os_version=10&cpu_core_num=16&device_memory=8&platform=PC&downlink=10&effective_type=4g&round_trip_time=50&webid=7352702470689539618&verifyFp=verify_lwizw3tf_mb5hS8lF_TDxT_4kdp_9HYY_cdOJNT9gfoj9&fp=verify_lwizw3tf_mb5hS8lF_TDxT_4kdp_9HYY_cdOJNT9gfoj9&msToken=9F1gtbNmcDoyreSeAFce5l26jXwA_FNeEEw6uy5AwV--R9SgHKJwhB-1FT2xRrTuWdztVm97Mf2_rsDeVdJFOuf8m6tqRyMm5R3rdfuAH7Ix7ez_BCC-hsOxayVqnKeH&a_bogus=Qv8MQ5hvdiVTDDW65AdLfY3q6-q3YDR70trEMD2fkd3G4639HMOa9exoSJ7vP3LjNs%2FDIeEjy4hbY3cZrQcGM1wf9Skw%2F2CZm6T0t-P2so0j53inCgWME0hN-vW3SFqQ-wNAEOsQy75cFRw0W9QamhK4bfebY7Y6i6trrj%3D%3D"
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Referer", "https://www.douyin.com/user/MS4wLjABAAAAp2jMqtHQRKBfj1dAn8LdvFCxsJnBreZ1a8NU3zmpyDjf8rIyIdr1KhGsOTyA5oZA")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	req.Header.Add("Cookie", "csrf_session_id=0630bbd62e9f88a806b21be3b95f71b6; passport_assist_user=Cj0A5mjsHGzMMRItWq-dI4pq-MnoPzsTZqZWZJfqAqKGS0Pgbv5dPuiomMa2abRyLVGBhfFQPq2TtIGcauhxGkoKPKafyL4bcNaeFdx2A1_FKrBLj3xGGAzUG5HQsS_HCbAVCaaoKxNBn6tSe1DH3K9zmRclk5Y71-H4JT8uCBC3mcgNGImv1lQgASIBA9IOYio%3D; ttwid=1%7CauShNcQ1PHqvepvJhYeH7OaXpvIWWD6Wkf2ejBxvzCQ%7C1711934465%7Cce1f3d03ca6a01753e7fe9856d6cf1d272301e48f00ef6b7d9b5e26528a0b008; bd_ticket_guard_client_web_domain=2; sid_guard=a345e4979658d065956adf2c6156d6d3%7C1711934468%7C5184000%7CFri%2C+31-May-2024+01%3A21%3A08+GMT; douyin.com; device_web_cpu_core=16; device_web_memory_size=8; architecture=amd64; LOGIN_STATUS=1; store-region=cn-js; store-region-src=uid; xg_device_score=7.798873949579832; odin_tt=6585e6c1c7ef3961c6c5844b2dda4055e5726c4ba862355f5801dde11fa1d36ac67963a58a6e6bcecb8dd009ae60826b; passport_fe_beating_status=false; SEARCH_RESULT_LIST_TYPE=%22single%22; s_v_web_id=verify_lwizw3tf_mb5hS8lF_TDxT_4kdp_9HYY_cdOJNT9gfoj9; passport_csrf_token=8a8595a7ba40296fc7677c52765ac95b; passport_csrf_token_default=8a8595a7ba40296fc7677c52765ac95b; dy_swidth=1707; dy_sheight=1067; FORCE_LOGIN=%7B%22videoConsumedRemainSeconds%22%3A180%7D; download_guide=%223%2F20240613%2F0%22; pwa2=%220%7C0%7C3%7C0%22; volume_info=%7B%22isUserMute%22%3Afalse%2C%22isMute%22%3Afalse%2C%22volume%22%3A0.956%7D; strategyABtestKey=%221718342776.951%22; xgplayer_device_id=6840998882; xgplayer_user_id=863629687864; bd_ticket_guard_client_data=eyJiZC10aWNrZXQtZ3VhcmQtdmVyc2lvbiI6MiwiYmQtdGlja2V0LWd1YXJkLWl0ZXJhdGlvbi12ZXJzaW9uIjoxLCJiZC10aWNrZXQtZ3VhcmQtcmVlLXB1YmxpYy1rZXkiOiJCTHNacURhNkRocy96Z2dwZTRqemdPNkQzN3QvMjY4R05uOFlIQnp0K2U5NXovUGJFdk11TW9YYkxOZmlUYk15a0tOaG1wVytNQi9IZXF3V1FuZlVsRXM9IiwiYmQtdGlja2V0LWd1YXJkLXdlYi12ZXJzaW9uIjoxfQ%3D%3D; msToken=5foJAadsmkOTekHJn9TOK5uR_p0f7YytooyVfX7trICffJpuhuE91AnbklaYHsw8gfIdmuDnNOuPqlFLn9lhcl2a2FBm7VtZGqHN2mjqhv4BUhrAtey70N3vYMZIvnLw; __ac_nonce=0666c1bdb0086b462330c; __ac_signature=_02B4Z6wo00f01RIXv9wAAIDAzm4gatozirESN7tAACLsHd2wt.IeUp90yIjd-uO6pjJGykdS6fcmStoWyBNdlw10yP8VAj3n3wgqwPboH3OaRLGhA0kswY0lbdLtLN3vIwo1AcIWoKZ0PLeA06; IsDouyinActive=true; home_can_add_dy_2_desktop=%220%22; stream_recommend_feed_params=%22%7B%5C%22cookie_enabled%5C%22%3Atrue%2C%5C%22screen_width%5C%22%3A1707%2C%5C%22screen_height%5C%22%3A1067%2C%5C%22browser_online%5C%22%3Atrue%2C%5C%22cpu_core_num%5C%22%3A16%2C%5C%22device_memory%5C%22%3A8%2C%5C%22downlink%5C%22%3A10%2C%5C%22effective_type%5C%22%3A%5C%224g%5C%22%2C%5C%22round_trip_time%5C%22%3A50%7D%22")

	resp, err := client.Do(req)
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
		videoInfo, _ := json.Marshal(VideoInfo{
			{
				Title: v.Desc,
				Url:   v.Video.Cover.UrlList[0],
			},
		})
		tracks[k] = personTrack{
			PersonId:       9,
			PersonIdInside: "83250291547",
			Content:        v.Desc,
			ImageInfo:      "[]",
			VideoInfo:      string(videoInfo),
			Link:           "https://www.douyin.com/video/" + v.AwemeId,
			PubTime:        v.CreateTime,
		}
	}
	return tracks, nil
}

func (cr Crawler) scrapeDuiYou() ([]personTrack, error) {
	// 访问文章列表页
	url := "https://www.douyin.com/aweme/v1/web/aweme/post/?device_platform=webapp&aid=6383&channel=channel_pc_web&sec_user_id=MS4wLjABAAAAMYOzfiN_0IemUIGvq3LVTkGe6zRFCI_uQwLGQf2GQeIDh0GeqxRDb2I0lf3MPuIN&max_cursor=0&locate_query=false&show_live_replay_strategy=1&need_time_list=1&time_list_query=0&whale_cut_token=&cut_version=1&count=18&publish_video_strategy_type=2&update_version_code=170400&pc_client_type=1&version_code=290100&version_name=29.1.0&cookie_enabled=true&screen_width=1707&screen_height=1067&browser_language=zh-CN&browser_platform=Win32&browser_name=Chrome&browser_version=126.0.0.0&browser_online=true&engine_name=Blink&engine_version=126.0.0.0&os_name=Windows&os_version=10&cpu_core_num=16&device_memory=8&platform=PC&downlink=10&effective_type=4g&round_trip_time=50&webid=7352702470689539618&msToken=B714gwB0IFqR6IVDXKcqpqUcFa5KR_yMf0e9z1CrhFx6RUAbLeoS4rL7NbAYZvrYniJW2gKkmxCGW4ip2-wm-Lvp3inJx93gHDPQ_waj-JWguGTDsECUdh5su_ETyvH-&a_bogus=mJmM%2FQ06DiDNDfyh5AdLfY3q6133YQje0trEMD2fOV3Wr639HMP89exLSJUvk6LjNs%2FDIeEjy4hbY3cZrQcGM1wf9Skw%2F2CZm6T0t-P2so0j53inCgWME0hN-vW3SFqQ-wNAEOsQy75cFRw0W9QamhK4bfebY7Y6i6trsf%3D%3D&verifyFp=verify_lwizw3tf_mb5hS8lF_TDxT_4kdp_9HYY_cdOJNT9gfoj9&fp=verify_lwizw3tf_mb5hS8lF_TDxT_4kdp_9HYY_cdOJNT9gfoj9"
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Referer", "https://www.douyin.com/user/MS4wLjABAAAAMYOzfiN_0IemUIGvq3LVTkGe6zRFCI_uQwLGQf2GQeIDh0GeqxRDb2I0lf3MPuIN")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	req.Header.Add("Cookie", "csrf_session_id=0630bbd62e9f88a806b21be3b95f71b6; passport_assist_user=Cj0A5mjsHGzMMRItWq-dI4pq-MnoPzsTZqZWZJfqAqKGS0Pgbv5dPuiomMa2abRyLVGBhfFQPq2TtIGcauhxGkoKPKafyL4bcNaeFdx2A1_FKrBLj3xGGAzUG5HQsS_HCbAVCaaoKxNBn6tSe1DH3K9zmRclk5Y71-H4JT8uCBC3mcgNGImv1lQgASIBA9IOYio%3D; ttwid=1%7CauShNcQ1PHqvepvJhYeH7OaXpvIWWD6Wkf2ejBxvzCQ%7C1711934465%7Cce1f3d03ca6a01753e7fe9856d6cf1d272301e48f00ef6b7d9b5e26528a0b008; bd_ticket_guard_client_web_domain=2; sid_guard=a345e4979658d065956adf2c6156d6d3%7C1711934468%7C5184000%7CFri%2C+31-May-2024+01%3A21%3A08+GMT; douyin.com; device_web_cpu_core=16; device_web_memory_size=8; architecture=amd64; LOGIN_STATUS=1; store-region=cn-js; store-region-src=uid; xg_device_score=7.798873949579832; odin_tt=6585e6c1c7ef3961c6c5844b2dda4055e5726c4ba862355f5801dde11fa1d36ac67963a58a6e6bcecb8dd009ae60826b; passport_fe_beating_status=false; SEARCH_RESULT_LIST_TYPE=%22single%22; s_v_web_id=verify_lwizw3tf_mb5hS8lF_TDxT_4kdp_9HYY_cdOJNT9gfoj9; passport_csrf_token=8a8595a7ba40296fc7677c52765ac95b; passport_csrf_token_default=8a8595a7ba40296fc7677c52765ac95b; dy_swidth=1707; dy_sheight=1067; FORCE_LOGIN=%7B%22videoConsumedRemainSeconds%22%3A180%7D; download_guide=%223%2F20240613%2F0%22; pwa2=%220%7C0%7C3%7C0%22; volume_info=%7B%22isUserMute%22%3Afalse%2C%22isMute%22%3Afalse%2C%22volume%22%3A0.956%7D; strategyABtestKey=%221718342776.951%22; xgplayer_device_id=6840998882; xgplayer_user_id=863629687864; bd_ticket_guard_client_data=eyJiZC10aWNrZXQtZ3VhcmQtdmVyc2lvbiI6MiwiYmQtdGlja2V0LWd1YXJkLWl0ZXJhdGlvbi12ZXJzaW9uIjoxLCJiZC10aWNrZXQtZ3VhcmQtcmVlLXB1YmxpYy1rZXkiOiJCTHNacURhNkRocy96Z2dwZTRqemdPNkQzN3QvMjY4R05uOFlIQnp0K2U5NXovUGJFdk11TW9YYkxOZmlUYk15a0tOaG1wVytNQi9IZXF3V1FuZlVsRXM9IiwiYmQtdGlja2V0LWd1YXJkLXdlYi12ZXJzaW9uIjoxfQ%3D%3D; msToken=5foJAadsmkOTekHJn9TOK5uR_p0f7YytooyVfX7trICffJpuhuE91AnbklaYHsw8gfIdmuDnNOuPqlFLn9lhcl2a2FBm7VtZGqHN2mjqhv4BUhrAtey70N3vYMZIvnLw; __ac_nonce=0666c1bdb0086b462330c; __ac_signature=_02B4Z6wo00f01RIXv9wAAIDAzm4gatozirESN7tAACLsHd2wt.IeUp90yIjd-uO6pjJGykdS6fcmStoWyBNdlw10yP8VAj3n3wgqwPboH3OaRLGhA0kswY0lbdLtLN3vIwo1AcIWoKZ0PLeA06; IsDouyinActive=true; home_can_add_dy_2_desktop=%220%22; stream_recommend_feed_params=%22%7B%5C%22cookie_enabled%5C%22%3Atrue%2C%5C%22screen_width%5C%22%3A1707%2C%5C%22screen_height%5C%22%3A1067%2C%5C%22browser_online%5C%22%3Atrue%2C%5C%22cpu_core_num%5C%22%3A16%2C%5C%22device_memory%5C%22%3A8%2C%5C%22downlink%5C%22%3A10%2C%5C%22effective_type%5C%22%3A%5C%224g%5C%22%2C%5C%22round_trip_time%5C%22%3A50%7D%22")

	resp, err := client.Do(req)
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
		videoInfo, _ := json.Marshal(VideoInfo{
			{
				Title: v.Desc,
				Url:   v.Video.Cover.UrlList[0],
			},
		})
		tracks[k] = personTrack{
			PersonId:       10,
			PersonIdInside: "83250291547",
			Content:        v.Desc,
			ImageInfo:      "[]",
			VideoInfo:      string(videoInfo),
			Link:           "https://www.douyin.com/video/" + v.AwemeId,
			PubTime:        v.CreateTime,
		}
	}
	return tracks, nil
}

func (cr Crawler) scrapeKouZi() ([]personTrack, error) {
	// 访问文章列表页
	url := "https://www.douyin.com/aweme/v1/web/aweme/post/?device_platform=webapp&aid=6383&channel=channel_pc_web&sec_user_id=MS4wLjABAAAAPpV84oc56lBGWOoK5AIfNRCkjn9X2SjAGpxJfLU12IKWPeAizXjKxvGuiN3E8dBA&max_cursor=0&locate_query=false&show_live_replay_strategy=1&need_time_list=1&time_list_query=0&whale_cut_token=&cut_version=1&count=18&publish_video_strategy_type=2&update_version_code=170400&pc_client_type=1&version_code=290100&version_name=29.1.0&cookie_enabled=true&screen_width=1707&screen_height=1067&browser_language=zh-CN&browser_platform=Win32&browser_name=Chrome&browser_version=126.0.0.0&browser_online=true&engine_name=Blink&engine_version=126.0.0.0&os_name=Windows&os_version=10&cpu_core_num=16&device_memory=8&platform=PC&downlink=10&effective_type=4g&round_trip_time=50&webid=7352702470689539618&verifyFp=verify_lwizw3tf_mb5hS8lF_TDxT_4kdp_9HYY_cdOJNT9gfoj9&fp=verify_lwizw3tf_mb5hS8lF_TDxT_4kdp_9HYY_cdOJNT9gfoj9&msToken=kmP_0Okz0P-R_R58ltAO59aOxecZVGV0EuNc_p0IvTdFnGMkPU15RbYXMgKqckvK64wP5CMnYDkbCAks3eIXOZYjmHQ2-dCK_1QZk1NeGn1Sl5jNHpWxSdfjj1r2XogN&a_bogus=Q6W0MV0DDkfshDW65AdLfY3q6lP3YDRt0trEMD2flx3G5L39HMY39exoSJXvYTgjNs%2FDIeEjy4hbY3cZrQcGM1wf9Skw%2F2CZm6T0t-P2so0j53inCgWME0hN-vW3SFqQ-wNAEOsQy75cFRw0W9QamhK4bfebY7Y6i6trtE%3D%3D"
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Referer", "https://www.douyin.com/user/MS4wLjABAAAAPpV84oc56lBGWOoK5AIfNRCkjn9X2SjAGpxJfLU12IKWPeAizXjKxvGuiN3E8dBA")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	req.Header.Add("Cookie", "csrf_session_id=0630bbd62e9f88a806b21be3b95f71b6; passport_assist_user=Cj0A5mjsHGzMMRItWq-dI4pq-MnoPzsTZqZWZJfqAqKGS0Pgbv5dPuiomMa2abRyLVGBhfFQPq2TtIGcauhxGkoKPKafyL4bcNaeFdx2A1_FKrBLj3xGGAzUG5HQsS_HCbAVCaaoKxNBn6tSe1DH3K9zmRclk5Y71-H4JT8uCBC3mcgNGImv1lQgASIBA9IOYio%3D; ttwid=1%7CauShNcQ1PHqvepvJhYeH7OaXpvIWWD6Wkf2ejBxvzCQ%7C1711934465%7Cce1f3d03ca6a01753e7fe9856d6cf1d272301e48f00ef6b7d9b5e26528a0b008; bd_ticket_guard_client_web_domain=2; sid_guard=a345e4979658d065956adf2c6156d6d3%7C1711934468%7C5184000%7CFri%2C+31-May-2024+01%3A21%3A08+GMT; douyin.com; device_web_cpu_core=16; device_web_memory_size=8; architecture=amd64; LOGIN_STATUS=1; store-region=cn-js; store-region-src=uid; xg_device_score=7.798873949579832; odin_tt=6585e6c1c7ef3961c6c5844b2dda4055e5726c4ba862355f5801dde11fa1d36ac67963a58a6e6bcecb8dd009ae60826b; passport_fe_beating_status=false; SEARCH_RESULT_LIST_TYPE=%22single%22; s_v_web_id=verify_lwizw3tf_mb5hS8lF_TDxT_4kdp_9HYY_cdOJNT9gfoj9; passport_csrf_token=8a8595a7ba40296fc7677c52765ac95b; passport_csrf_token_default=8a8595a7ba40296fc7677c52765ac95b; dy_swidth=1707; dy_sheight=1067; FORCE_LOGIN=%7B%22videoConsumedRemainSeconds%22%3A180%7D; download_guide=%223%2F20240613%2F0%22; pwa2=%220%7C0%7C3%7C0%22; volume_info=%7B%22isUserMute%22%3Afalse%2C%22isMute%22%3Afalse%2C%22volume%22%3A0.956%7D; strategyABtestKey=%221718342776.951%22; xgplayer_device_id=6840998882; xgplayer_user_id=863629687864; bd_ticket_guard_client_data=eyJiZC10aWNrZXQtZ3VhcmQtdmVyc2lvbiI6MiwiYmQtdGlja2V0LWd1YXJkLWl0ZXJhdGlvbi12ZXJzaW9uIjoxLCJiZC10aWNrZXQtZ3VhcmQtcmVlLXB1YmxpYy1rZXkiOiJCTHNacURhNkRocy96Z2dwZTRqemdPNkQzN3QvMjY4R05uOFlIQnp0K2U5NXovUGJFdk11TW9YYkxOZmlUYk15a0tOaG1wVytNQi9IZXF3V1FuZlVsRXM9IiwiYmQtdGlja2V0LWd1YXJkLXdlYi12ZXJzaW9uIjoxfQ%3D%3D; msToken=5foJAadsmkOTekHJn9TOK5uR_p0f7YytooyVfX7trICffJpuhuE91AnbklaYHsw8gfIdmuDnNOuPqlFLn9lhcl2a2FBm7VtZGqHN2mjqhv4BUhrAtey70N3vYMZIvnLw; __ac_nonce=0666c1bdb0086b462330c; __ac_signature=_02B4Z6wo00f01RIXv9wAAIDAzm4gatozirESN7tAACLsHd2wt.IeUp90yIjd-uO6pjJGykdS6fcmStoWyBNdlw10yP8VAj3n3wgqwPboH3OaRLGhA0kswY0lbdLtLN3vIwo1AcIWoKZ0PLeA06; IsDouyinActive=true; home_can_add_dy_2_desktop=%220%22; stream_recommend_feed_params=%22%7B%5C%22cookie_enabled%5C%22%3Atrue%2C%5C%22screen_width%5C%22%3A1707%2C%5C%22screen_height%5C%22%3A1067%2C%5C%22browser_online%5C%22%3Atrue%2C%5C%22cpu_core_num%5C%22%3A16%2C%5C%22device_memory%5C%22%3A8%2C%5C%22downlink%5C%22%3A10%2C%5C%22effective_type%5C%22%3A%5C%224g%5C%22%2C%5C%22round_trip_time%5C%22%3A50%7D%22")

	resp, err := client.Do(req)
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
		videoInfo, _ := json.Marshal(VideoInfo{
			{
				Title: v.Desc,
				Url:   v.Video.Cover.UrlList[0],
			},
		})
		tracks[k] = personTrack{
			PersonId:       11,
			PersonIdInside: "83250291547",
			Content:        v.Desc,
			ImageInfo:      "[]",
			VideoInfo:      string(videoInfo),
			Link:           "https://www.douyin.com/video/" + v.AwemeId,
			PubTime:        v.CreateTime,
		}
	}
	return tracks, nil
}

func (cr Crawler) scrapeKeJiZaoZhiDao(ctx context.Context) ([]personTrack, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, cr.Timeout)
	defer cancel()

	fmt.Printf("%s [科技早知道]初始化Chromedp上下文成功\n", time.Now().Format("01-02 15:04:05"))
	// 访问文章列表页
	if err := chromedp.Run(timeoutCtx,
		chromedp.Navigate("https://www.xiaoyuzhoufm.com/podcast/5e74b52c418a84a046ecaceb"),
	); err != nil {
		return nil, err
	}
	fmt.Printf("%s [科技早知道]访问动态列表页成功\n", time.Now().Format("01-02 15:04:05"))
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
		tracks[k] = personTrack{
			PersonId:       12,
			PersonIdInside: "",
			Content:        content,
			ImageInfo:      "[]",
			VideoInfo:      "[]",
			Link:           link,
			PubTime:        int(t.Unix()),
		}
		fmt.Printf("%s [科技早知道]获取当前动态成功,内容%v，时间：%d\n", time.Now().Format("01-02 15:04:05"), content, t.Unix())
	}

	return tracks, nil
}

func (cr Crawler) scrapeZhenGeJiJin(ctx context.Context) ([]personTrack, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, cr.Timeout)
	defer cancel()

	fmt.Printf("%s [真格基金]初始化Chromedp上下文成功\n", time.Now().Format("01-02 15:04:05"))
	// 访问文章列表页
	if err := chromedp.Run(timeoutCtx,
		chromedp.Navigate("https://www.xiaoyuzhoufm.com/podcast/646f194853a5e5ea1408d97c"),
	); err != nil {
		return nil, err
	}
	fmt.Printf("%s [真格基金]访问动态列表页成功\n", time.Now().Format("01-02 15:04:05"))
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
		tracks[k] = personTrack{
			PersonId:       13,
			PersonIdInside: "",
			Content:        content,
			ImageInfo:      "[]",
			VideoInfo:      "[]",
			Link:           link,
			PubTime:        int(t.Unix()),
		}
		fmt.Printf("%s [真格基金]获取当前动态成功,内容%v，时间：%d\n", time.Now().Format("01-02 15:04:05"), content, t.Unix())
	}

	return tracks, nil
}

func (cr Crawler) scrapeAiJuNeiRen(ctx context.Context) ([]personTrack, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, cr.Timeout)
	defer cancel()

	fmt.Printf("%s [AI局内人]初始化Chromedp上下文成功\n", time.Now().Format("01-02 15:04:05"))
	// 访问文章列表页
	if err := chromedp.Run(timeoutCtx,
		chromedp.Navigate("https://www.xiaoyuzhoufm.com/podcast/643928f99361a4e7c38a9555"),
	); err != nil {
		return nil, err
	}
	fmt.Printf("%s [AI局内人]访问动态列表页成功\n", time.Now().Format("01-02 15:04:05"))
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
		tracks[k] = personTrack{
			PersonId:       14,
			PersonIdInside: "",
			Content:        content,
			ImageInfo:      "[]",
			VideoInfo:      "[]",
			Link:           link,
			PubTime:        int(t.Unix()),
		}
		fmt.Printf("%s [AI局内人]获取当前动态成功,内容%v，时间：%d\n", time.Now().Format("01-02 15:04:05"), content, t.Unix())
	}

	return tracks, nil
}

func (cr Crawler) scrape42ZhangJing(ctx context.Context) ([]personTrack, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, cr.Timeout)
	defer cancel()

	fmt.Printf("%s [42章经]初始化Chromedp上下文成功\n", time.Now().Format("01-02 15:04:05"))
	// 访问文章列表页
	if err := chromedp.Run(timeoutCtx,
		chromedp.Navigate("https://www.xiaoyuzhoufm.com/podcast/648b0b641c48983391a63f98"),
	); err != nil {
		return nil, err
	}
	fmt.Printf("%s [42章经]访问动态列表页成功\n", time.Now().Format("01-02 15:04:05"))
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
		tracks[k] = personTrack{
			PersonId:       15,
			PersonIdInside: "",
			Content:        content,
			ImageInfo:      "[]",
			VideoInfo:      "[]",
			Link:           link,
			PubTime:        int(t.Unix()),
		}
		fmt.Printf("%s [42章经]获取当前动态成功,内容%v，时间：%d\n", time.Now().Format("01-02 15:04:05"), content, t.Unix())
	}

	return tracks, nil
}
