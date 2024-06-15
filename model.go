package main

type douyinGetVideoListResp struct {
	AwemeList []struct {
		AwemeId           string `json:"aweme_id"`    // 视频id->跳转链接：https://www.douyin.com/video/{id}
		Desc              string `json:"desc"`        // content
		CreateTime        int    `json:"create_time"` // pub time
		FriendInteraction int    `json:"friend_interaction"`
		Video             struct {
			Cover struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"` // video cover
				Width   int      `json:"width"`
				Height  int      `json:"height"`
			} `json:"cover"`
		} `json:"video"`
	} `json:"aweme_list"`
}
