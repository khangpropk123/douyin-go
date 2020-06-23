package tools

import (
	"../model"
	"archive/zip"
	"bytes"
	"container/list"
	"encoding/json"
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
	var link = GetUserInfo(data.Url)
	var sign = GetSignature(link)
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
	listPost.PushBack(post)
	for {
		if !post.HasMore {
			fmt.Println(listPost.Len())
			break
		}
		post = GetPostData(link.Uid, sign.Signature, int(post.MaxCursor))
		listPost.PushBack(post)
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
