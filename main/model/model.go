package model

type DouyinPost struct {
	MinCursor int64 `json:"min_cursor"`
	HasMore   bool  `json:"has_more"`
	Extra     struct {
		Now   int64  `json:"now"`
		Logid string `json:"logid"`
	} `json:"extra"`
	StatusCode int `json:"status_code"`
	AwemeList  []struct {
		Author struct {
			UniqueID            string      `json:"unique_id"`
			PolicyVersion       interface{} `json:"policy_version"`
			FollowersDetail     interface{} `json:"followers_detail"`
			Geofencing          interface{} `json:"geofencing"`
			WithFusionShopEntry bool        `json:"with_fusion_shop_entry"`
			ShortID             string      `json:"short_id"`
			IsAdFake            bool        `json:"is_ad_fake"`
			HasOrders           bool        `json:"has_orders"`
			IsEnterpriseVip     bool        `json:"is_enterprise_vip"`
			Rate                int         `json:"rate"`
			Nickname            string      `json:"nickname"`
			FollowingCount      int         `json:"following_count"`
			Region              string      `json:"region"`
			IsGovMediaVip       bool        `json:"is_gov_media_vip"`
			AvatarLarger        struct {
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
			} `json:"avatar_larger"`
			AvatarMedium struct {
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
			} `json:"avatar_medium"`
			StoryOpen         bool   `json:"story_open"`
			WithCommerceEntry bool   `json:"with_commerce_entry"`
			FavoritingCount   int    `json:"favoriting_count"`
			TotalFavorited    int    `json:"total_favorited"`
			WithShopEntry     bool   `json:"with_shop_entry"`
			SecUID            string `json:"sec_uid"`
			AvatarThumb       struct {
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
			} `json:"avatar_thumb"`
			PlatformSyncInfo interface{} `json:"platform_sync_info"`
			FollowerCount    int         `json:"follower_count"`
			Secret           int         `json:"secret"`
			VideoIcon        struct {
				URI     string        `json:"uri"`
				URLList []interface{} `json:"url_list"`
			} `json:"video_icon"`
			Signature    string `json:"signature"`
			AwemeCount   int    `json:"aweme_count"`
			UserCanceled bool   `json:"user_canceled"`
			UID          string `json:"uid"`
			FollowStatus int    `json:"follow_status"`
		} `json:"author"`
		ChaList interface{} `json:"cha_list"`
		Video   struct {
			DownloadAddr struct {
				URLList []string `json:"url_list"`
				URI     string   `json:"uri"`
			} `json:"download_addr"`
			BitRate  interface{} `json:"bit_rate"`
			Duration int         `json:"duration"`
			Vid      string      `json:"vid"`
			PlayAddr struct {
				URLList []string `json:"url_list"`
				URI     string   `json:"uri"`
			} `json:"play_addr"`
			Cover struct {
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
			} `json:"cover"`
			DynamicCover struct {
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
			} `json:"dynamic_cover"`
			OriginCover struct {
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
			} `json:"origin_cover"`
			Ratio         string `json:"ratio"`
			HasWatermark  bool   `json:"has_watermark"`
			PlayAddrLowbr struct {
				URLList []string `json:"url_list"`
				URI     string   `json:"uri"`
			} `json:"play_addr_lowbr"`
			Height int `json:"height"`
			Width  int `json:"width"`
		} `json:"video"`
		VideoLabels interface{} `json:"video_labels"`
		Statistics  struct {
			AwemeID      string `json:"aweme_id"`
			CommentCount int    `json:"comment_count"`
			DiggCount    int    `json:"digg_count"`
			PlayCount    int    `json:"play_count"`
			ShareCount   int    `json:"share_count"`
			ForwardCount int    `json:"forward_count"`
		} `json:"statistics"`
		TextExtra []struct {
			Start       int    `json:"start"`
			End         int    `json:"end"`
			Type        int    `json:"type"`
			HashtagName string `json:"hashtag_name"`
			HashtagID   int64  `json:"hashtag_id"`
		} `json:"text_extra"`
		ImageInfos     interface{} `json:"image_infos"`
		Position       interface{} `json:"position"`
		AwemeID        string      `json:"aweme_id"`
		UniqidPosition interface{} `json:"uniqid_position"`
		Geofencing     interface{} `json:"geofencing"`
		VideoText      interface{} `json:"video_text"`
		Desc           string      `json:"desc"`
		AwemeType      int         `json:"aweme_type"`
		CommentList    interface{} `json:"comment_list"`
		LabelTopText   interface{} `json:"label_top_text"`
		Promotions     interface{} `json:"promotions"`
		LongVideo      interface{} `json:"long_video"`
	} `json:"aweme_list"`
	MaxCursor int64 `json:"max_cursor"`
}

type SignaturePost struct{
	Tac string `json:"tac"`
	Uid string `json:"uid"`
}
type SignatureResp struct {
	Signature string `json:"signature"`
	UserAgent string `json:"user-agent"`
}
type Description struct {
	Id string `json:"id"`
	Desc string `json:"desc"`
}