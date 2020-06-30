package tools

import (
	"../model"
	"archive/zip"
	"bytes"
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	igdl "github.com/siongui/instago/download"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type DeleteChan struct {
	Delete chan string
}

type Req struct {
	Username string `json:"username"`
	Cookies  string `json:"cookies"`
	Url      string `json:"url"`
	Kind     int    `json:"kind"`
	Share    int    `json:"share"`
	Cmt      int    `json:"cmt"`
	Play     int    `json:"play"`
}
type ErrorRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
type Info struct {
	AuthorName string `json:"author_name"`
	Id         string `json:"id"`
	Follow     int    `json:"follow"`
	Region     string `json:"region"`
	Sign       string `json:"sign"`
	State      int    `json:"state"`
	Result     string `json:"result"`
	Total      int    `json:"total"`
	Progress   int    `json:"progress"`
}
type SignatureInfo struct {
	Signature string `json:"signature"`
	UserAgent string `json:"user-agent"`
}

const MOBILE_HEADER = "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1"
const NORMAL_HEADER = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.108 Safari/537.36"

var listPost = list.New()

func MainWorkFlow(data *Req, ws *websocket.Conn, mutex *sync.Mutex) string {
	if data.Url == "" {
		_ = ws.WriteJSON(&ErrorRes{
			Code:    400,
			Message: "Bad request url",
		})
		return ""
	}
	var link = GetUserInfoEx(data.Url)
	var sign = GetSignatureEx(link)
	var maxCursor = 0
	post := GetPostData(link.Uid, sign.Signature, maxCursor)
	//if fileExists("./Downloaded/" + post.AwemeList[0].Author.Nickname + ".zip") {
	//	var res = &Info{
	//		AuthorName: post.AwemeList[0].Author.Nickname,
	//		Id:         post.AwemeList[0].Author.ShortID,
	//		Follow:     post.AwemeList[0].Author.FollowingCount,
	//		Region:     post.AwemeList[0].Author.Region,
	//		Sign:       post.AwemeList[0].Author.Signature,
	//		State:      2,
	//		Result:     "http://65.52.184.198/download?kind=1&file=" + post.AwemeList[0].Author.Nickname + ".zip",
	//		Total:      0,
	//		Progress:   0,
	//	}
	//	_ = ws.WriteJSON(res)
	//	return ""
	//}
	if len(post.AwemeList)>0{
		listPost.PushBack(post)
	}
	var cusor = int(post.MaxCursor)
	for {

		if !post.HasMore {
			fmt.Println(listPost.Len())
			break
		}
		if int(post.MaxCursor) > 0{
			cusor = int(post.MaxCursor)
		}
		post = GetPostData(link.Uid, sign.Signature, int(cusor))
		if  len(post.AwemeList) > 0{
			listPost.PushBack(post)
		}
	}
	var listDesc = make([]*model.Description, 0, 0)
	var authorName = ""
	var wg sync.WaitGroup
	var info Info
	if listPost.Len() > 0 {
		authorName = listPost.Front().Value.(*model.DouyinPost).AwemeList[0].Author.Nickname
		var follow = listPost.Front().Value.(*model.DouyinPost).AwemeList[0].Author.FollowerCount
		var region = listPost.Front().Value.(*model.DouyinPost).AwemeList[0].Author.Region
		var id = listPost.Front().Value.(*model.DouyinPost).AwemeList[0].Author.ShortID
		var sign = listPost.Front().Value.(*model.DouyinPost).AwemeList[0].Author.Signature
		info = Info{
			AuthorName: authorName,
			Id:         id,
			Follow:     follow,
			Region:     region,
			Sign:       sign,
			State:      0,
		}
	}
	totalVideo := 0
	for l := listPost.Front(); l != nil; l = l.Next() {
		for _, awe := range l.Value.(*model.DouyinPost).AwemeList {
			if awe.Statistics.ShareCount >= data.Share && awe.Statistics.CommentCount >= data.Cmt && awe.Statistics.DiggCount >= data.Play {
				totalVideo++
			}
		}
	}
	info.Total = totalVideo

	for l := listPost.Front(); l != nil; l = l.Next() {
		for _, awe := range l.Value.(*model.DouyinPost).AwemeList {
			if awe.Statistics.ShareCount >= data.Share && awe.Statistics.CommentCount >= data.Cmt && awe.Statistics.DiggCount >= data.Play {
				authorName = awe.Author.Nickname
				//vBox.Append(widget.NewLabel("Author Name: "+authorName))
				wg.Add(1)
				go DownloadVideo(awe.Author.Nickname, awe.Video.PlayAddr.URLList[0], awe.AwemeID, &wg, ws, &info, mutex)
				var desc = &model.Description{
					Id:   awe.AwemeID,
					Desc: awe.Desc,
				}
				listDesc = append(listDesc, desc)
			}
		}
	}
	byteData, _ := json.Marshal(listDesc)
	var path = "./Downloaded/" + authorName + "/"
	_ = ioutil.WriteFile(path+"Description.json", byteData, 0777)
	wg.Wait()
	ZipFolder(path, "./Downloaded/"+authorName+".zip")
	info.State = 2
	info.Result = "http://65.52.184.198/download?kind=1&file=" + authorName + ".zip"
	_ = ws.WriteJSON(info)
	_ = os.RemoveAll(path)
	return "./Downloaded/" + authorName + ".zip"

}

func GetUserInfo(url string) *model.SignaturePost {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", NORMAL_HEADER)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	tacRe := regexp.MustCompile(`tac='([\s\S]*?)'</script>`)
	uidRe := regexp.MustCompile(`uid: "(\d+?)"`)

	var uid = strings.ReplaceAll(strings.ReplaceAll(uidRe.FindString(string(body)), `uid: "`, ""), `"`, "")
	var tac = strings.ReplaceAll(strings.Split(tacRe.FindString(string(body)), "|")[0], `tac='`, "")
	var sign = &model.SignaturePost{
		Tac: tac,
		Uid: uid,
	}
	fmt.Println(sign)
	return sign
}

func GetUserInfoEx(url string) *model.SignaturePost  {
	re := regexp.MustCompile(`/([0-9]+)\?`)
	uid := re.FindStringSubmatch(url)
	if len(uid)>0{
		var sign = &model.SignaturePost{
			Tac: "",
			Uid: strings.ReplaceAll(strings.ReplaceAll(uid[0],`/`,""),`?`,""),
		}
		fmt.Println(sign)
		return sign
	}
	return nil
}
func GetSignatureEx(sign *model.SignaturePost) *SignatureInfo {
	url := fmt.Sprintf(`http://49.233.200.77:5001/sign/%s/`,sign.Uid)
	r, _ := http.NewRequest("GET", url,nil)
	//data,_ := json.Marshal(sign)
	//fmt.Println(string(data))

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var signResp = &SignatureInfo{}
	_ = json.Unmarshal(body, signResp)
	fmt.Println(signResp)
	return signResp
}
func GetSignature(sign *model.SignaturePost) *model.SignatureResp {
	data := url.Values{}
	data.Set("tac", sign.Tac)
	data.Set("user_id", sign.Uid)
	r, _ := http.NewRequest("POST", "http://49.233.200.77:5001/", strings.NewReader(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	//data,_ := json.Marshal(sign)
	//fmt.Println(string(data))

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var signResp = &model.SignatureResp{}
	json.Unmarshal(body, signResp)
	return signResp
}

func GetPostData(uid string, sign string, maxCursor int) *model.DouyinPost {
	var baseUrl = fmt.Sprintf(`https://www.iesdouyin.com/web/api/v2/aweme/post/?user_id=%s&sec_uid=&count=21&max_cursor=%d&aid=1128&_signature=%s`, uid, maxCursor, sign)
	fmt.Println(baseUrl)
	req, err := http.NewRequest("GET", baseUrl, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", NORMAL_HEADER)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var name = &model.DouyinPost{}
	err = json.Unmarshal(body, &name)
	fmt.Println(err)
	return name
}

func DownloadVideo(name string, url string, filename string, wg *sync.WaitGroup, ws *websocket.Conn, info *Info, mutex *sync.Mutex) error {
	defer wg.Done()
	rootPath := "./Downloaded/" + name
	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		os.Mkdir(rootPath, 0777)
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(rootPath+"/"+filename+".mp4", buffer, 0777)
	if err != nil {
		return err
	}
	info.Progress++
	info.State = 1
	mutex.Lock()
	_ = ws.WriteJSON(info)
	mutex.Unlock()
	return nil
}

func ZipFolder(dest string, source string) {
	baseFolder := dest

	// Get a Buffer to Write To
	outFile, err := os.Create(source)
	if err != nil {
		fmt.Println(err)
	}
	defer outFile.Close()

	// Create a new zip archive.
	w := zip.NewWriter(outFile)

	// Add some files to the archive.
	addFiles(w, baseFolder, "")

	if err != nil {
		fmt.Println(err)
	}

	// Make sure to check the error on Close.
	err = w.Close()
	if err != nil {
		fmt.Println(err)
	}
}

func addFiles(w *zip.Writer, basePath, baseInZip string) {
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		fmt.Println(err)
	}

	for _, file := range files {
		if !file.IsDir() {
			dat, err := ioutil.ReadFile(basePath + file.Name())
			if err != nil {
				fmt.Println(err)
			}

			// Add some files to the archive.
			f, err := w.Create(baseInZip + file.Name())
			if err != nil {
				fmt.Println(err)
			}
			_, err = f.Write(dat)
			if err != nil {
				fmt.Println(err)
			}
		} else if file.IsDir() {

			// Recurse
			newBase := basePath + file.Name() + "/"
			fmt.Println("Recursing and Adding SubDir: " + file.Name())
			fmt.Println("Recursing and Adding SubDir: " + newBase)

			addFiles(w, newBase, baseInZip+file.Name()+"/")
		}
	}

}

func GetFb() string {
	url := "https://www.facebook.com/nga.duong97/videos/1690792357728370"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}
	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	reg_hd := regexp.MustCompile(`hd_src:(.*?),`)
	reg_sd := regexp.MustCompile(`sd_src:(.*?),`)
	hd_src := reg_hd.FindStringSubmatch(string(data))
	sd_src := reg_sd.FindStringSubmatch(string(data))
	if len(hd_src) >= 2 {
		fmt.Println(hd_src[1])
		reqhd, err := http.NewRequest("GET", strings.ReplaceAll(hd_src[1], `"`, ""), nil)
		video, err := http.DefaultClient.Do(reqhd)
		if err != nil {
			return ""
		}
		defer video.Body.Close()
		fileName := strings.ReplaceAll(url, `/`, "") + "_HD.mp4"
		data, err = ioutil.ReadAll(video.Body)
		err = ioutil.WriteFile(fileName, data, 0755)
		return fileName

	} else if len(sd_src) >= 2 {
		reqsd, err := http.NewRequest("GET", strings.ReplaceAll(sd_src[1], `"`, ""), nil)
		video, err := http.DefaultClient.Do(reqsd)
		if err != nil {
			return ""
		}
		defer video.Body.Close()
		fileName := strings.ReplaceAll(url, `/`, "") + "_SD.mp4"
		data, err = ioutil.ReadAll(video.Body)
		err = ioutil.WriteFile(fileName, data, 0755)
		return fileName
	} else {
		return ""
	}

}

func GetInstaCookie(username string, password string) (int, string, int64) {
	url := "https://www.instagram.com/accounts/login/ajax/"
	var params = []byte(fmt.Sprintf(`username=%s&password=%s&queryParams={}`, username, password))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(params))
	req.Header.Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Ubuntu Chromium/66.0.3359.139 Chrome/66.0.3359.139 Safari/537.36")
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.Header.Set("x-requested-with", "XMLHttpRequest")
	req.Header.Set("x-csrftoken", "Xz7lgke9oJh9CV8c2HjY2o0tB4EvQedb")
	req.Header.Set("cookie", "csrftoken=Xz7lgke9oJh9CV8c2HjY2o0tB4EvQedb; rur=FTW; mid=Wv1kwwAEAAF10fw8f1pJ9zUc4HcT")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Cookies:", resp.Cookies())
	return 1, "", 7
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func DownloadFileIG(user string, cookies string) (string, error) {
	ig, err := igdl.NewInstagramDownloadManager(cookies)
	if err != nil {
		return "", err
	}
	err = ig.DownloadUserStoryHighlightsByName(user)
	if err != nil {
		return "", err
	}
	ZipFolder("Instagram/"+user+"/", "File/Instagram/"+user+".zip")
	_ = os.RemoveAll("Instagram/" + user)
	return user + ".zip", nil
}

func DownloadFileFb(link string) string {
	url := link
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}
	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	reg_hd := regexp.MustCompile(`hd_src:(.*?),`)
	reg_sd := regexp.MustCompile(`sd_src:(.*?),`)
	hd_src := reg_hd.FindStringSubmatch(string(data))
	sd_src := reg_sd.FindStringSubmatch(string(data))
	if len(hd_src) >= 2 {
		fmt.Println(hd_src[1])
		reqhd, err := http.NewRequest("GET", strings.ReplaceAll(hd_src[1], `"`, ""), nil)
		video, err := http.DefaultClient.Do(reqhd)
		if err != nil {
			return ""
		}
		defer video.Body.Close()
		fileName := strings.ReplaceAll(url, `/`, "") + "_HD.mp4"
		data, err = ioutil.ReadAll(video.Body)
		err = ioutil.WriteFile("File/Facebook/"+fileName, data, 0755)
		return fileName

	} else if len(sd_src) >= 2 {
		reqsd, err := http.NewRequest("GET", strings.ReplaceAll(sd_src[1], `"`, ""), nil)
		video, err := http.DefaultClient.Do(reqsd)
		if err != nil {
			return ""
		}
		defer video.Body.Close()
		fileName := strings.ReplaceAll(url, `/`, "") + "_SD.mp4"
		data, err = ioutil.ReadAll(video.Body)
		err = ioutil.WriteFile("File/Facebook/"+fileName, data, 0755)
		return fileName
	} else {
		return ""
	}

}

type VideoDetail struct {
	StatusCode int `json:"status_code"`
	ItemList   []struct {
		Video struct {
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
			Ratio        string      `json:"ratio"`
			HasWatermark bool        `json:"has_watermark"`
			BitRate      interface{} `json:"bit_rate"`
			Vid          string      `json:"vid"`
			PlayAddr     struct {
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
			} `json:"play_addr"`
			Height   int `json:"height"`
			Width    int `json:"width"`
			Duration int `json:"duration"`
		} `json:"video"`
		UniqidPosition interface{} `json:"uniqid_position"`
		CommentList    interface{} `json:"comment_list"`
		Desc           string      `json:"desc"`
		Music          struct {
			Title   string `json:"title"`
			Author  string `json:"author"`
			CoverHd struct {
				URLList []string `json:"url_list"`
				URI     string   `json:"uri"`
			} `json:"cover_hd"`
			CoverMedium struct {
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
			} `json:"cover_medium"`
			Position   interface{} `json:"position"`
			Status     int         `json:"status"`
			ID         int64       `json:"id"`
			Mid        string      `json:"mid"`
			CoverLarge struct {
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
			} `json:"cover_large"`
			CoverThumb struct {
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
			} `json:"cover_thumb"`
			PlayURL struct {
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
			} `json:"play_url"`
			Duration int `json:"duration"`
		} `json:"music"`
		Statistics struct {
			CommentCount int    `json:"comment_count"`
			DiggCount    int    `json:"digg_count"`
			AwemeID      string `json:"aweme_id"`
		} `json:"statistics"`
		ShareInfo struct {
			ShareDesc      string `json:"share_desc"`
			ShareTitle     string `json:"share_title"`
			ShareWeiboDesc string `json:"share_weibo_desc"`
		} `json:"share_info"`
		AwemeType int         `json:"aweme_type"`
		Position  interface{} `json:"position"`
		LongVideo interface{} `json:"long_video"`
		ChaList   []struct {
			ChaName        string      `json:"cha_name"`
			Desc           string      `json:"desc"`
			UserCount      int         `json:"user_count"`
			Cid            string      `json:"cid"`
			ConnectMusic   interface{} `json:"connect_music"`
			Type           int         `json:"type"`
			ViewCount      int         `json:"view_count"`
			HashTagProfile string      `json:"hash_tag_profile"`
			IsCommerce     bool        `json:"is_commerce"`
		} `json:"cha_list"`
		IsLiveReplay bool        `json:"is_live_replay"`
		Duration     int         `json:"duration"`
		LabelTopText interface{} `json:"label_top_text"`
		Author       struct {
			TypeLabel        interface{} `json:"type_label"`
			Nickname         string      `json:"nickname"`
			PlatformSyncInfo interface{} `json:"platform_sync_info"`
			Signature        string      `json:"signature"`
			AvatarLarger     struct {
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
			} `json:"avatar_larger"`
			AvatarThumb struct {
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
			} `json:"avatar_thumb"`
			AvatarMedium struct {
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
			} `json:"avatar_medium"`
			UniqueID        string      `json:"unique_id"`
			FollowersDetail interface{} `json:"followers_detail"`
			UID             string      `json:"uid"`
			ShortID         string      `json:"short_id"`
			Geofencing      interface{} `json:"geofencing"`
			PolicyVersion   interface{} `json:"policy_version"`
		} `json:"author"`
		RiskInfos struct {
			Warn    bool   `json:"warn"`
			Type    int    `json:"type"`
			Content string `json:"content"`
		} `json:"risk_infos"`
		AuthorUserID int64       `json:"author_user_id"`
		Geofencing   interface{} `json:"geofencing"`
		GroupID      int64       `json:"group_id"`
		ShareURL     string      `json:"share_url"`
		TextExtra    []struct {
			Start       int    `json:"start"`
			End         int    `json:"end"`
			Type        int    `json:"type"`
			HashtagName string `json:"hashtag_name"`
			HashtagID   int64  `json:"hashtag_id"`
		} `json:"text_extra"`
		Promotions  interface{} `json:"promotions"`
		ForwardID   string      `json:"forward_id"`
		AwemeID     string      `json:"aweme_id"`
		CreateTime  int         `json:"create_time"`
		VideoLabels interface{} `json:"video_labels"`
		ImageInfos  interface{} `json:"image_infos"`
		VideoText   interface{} `json:"video_text"`
		IsPreview   int         `json:"is_preview"`
	} `json:"item_list"`
	Extra struct {
		Now   int64  `json:"now"`
		Logid string `json:"logid"`
	} `json:"extra"`
}

func DownloadDouyin(link string) (string,error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		} }


	r, err := http.NewRequest("GET", link, nil)
	if err != nil{
		return "",err
	}
	//r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Host","v.douyin.com")
	r.Header.Add("Accept","text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	r.Header.Add("Accept-Language","zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
	//r.Header.Add("Accept-Encoding","gzip, deflate")
	r.Header.Add("Connection","keep-alive")
	r.Header.Add("Upgrade-Insecure-Requests","1")
	r.Header.Add("User-Agent","Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:78.0) Gecko/20100101 Firefox/78.0")
	resp, err := client.Do(r)
	if err != nil{
		return "",err
	}
	defer resp.Body.Close()
	nextlink,_ := resp.Location()
	url_id := strings.Split(nextlink.String(),"/")
	var file = ""
	if len(url_id) > 6{
		r2 , err := http.NewRequest("GET","https://www.iesdouyin.com/web/api/v2/aweme/iteminfo/?item_ids="+url_id[5],nil)
		if err != nil{
			return "",err
		}
		r2.Header.Add("Host","www.iesdouyin.com")
		r2.Header.Add("Accept","*/*")
		r2.Header.Add("X-Requested-With","XMLHttpRequest")
		r2.Header.Add("Referer",nextlink.String())
		resp2, err := client.Do(r2)
		if err != nil{
			return "",err
		}
		defer resp2.Body.Close()
		data, _ := ioutil.ReadAll(resp2.Body)
		var detail  VideoDetail
		err = json.Unmarshal(data, &detail)
		if err != nil{
			return "",err
		}
		if len(detail.ItemList)>0{
			fmt.Println(detail.ItemList[0].Video.PlayAddr.URLList[0])
			file, err = DownLoad(strings.ReplaceAll(detail.ItemList[0].Video.PlayAddr.URLList[0],"playwm","play"),detail.ItemList[0].AwemeID+".mp4")
			if err != nil{
				return "",err
			}
		} else {

				return "",errors.New("Error")
		}
		}

	////regex := regexp.MustCompile(`([0-9]+)`)
	//a := doc.Find("a").Nodes
	////b := regex.FindStringSubmatch(a)
	////if len(b)>0{
	return file, nil
}

func DownLoad(link string, name string) (string, error) {
	//client := &http.Client{
	//	CheckRedirect: func(req *http.Request, via []*http.Request) error {
	//		return http.ErrUseLastResponse
	//	} }
	r2 , _:= http.NewRequest("GET",link,nil)

	r2.Header.Add("Host","aweme.snssdk.com")
	r2.Header.Add("User-Agent","Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1")
	r2.Header.Add("Accept","ext/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	resp, err:= http.DefaultClient.Do(r2)
	if err != nil{
		return "",err
	}
	defer resp.Body.Close()
	data1, err:= ioutil.ReadAll(resp.Body)
	if err != nil{
		return "",err
	}
	err = ioutil.WriteFile("File/Douyin/"+name, data1, 0777)
	if err != nil{
		return "",err
	}
	return name,nil
}