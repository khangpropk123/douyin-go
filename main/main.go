package main

import (
	"./controllers"
	"./model"
	"./tools"
	"archive/zip"
	"bytes"
	"container/list"
	"encoding/json"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	"github.com/astaxie/beego"
	igdl "github.com/siongui/instago/download"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const MOBILE_HEADER = "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1"
const NORMAL_HEADER = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.108 Safari/537.36"

var listPost = list.New()

func main() {
	var cookies = `{
  "ig_did": "8FEEA7BB-44E9-401B-ABAC-3509E2210FCB",
  "mid": "Xu2XzgAEAAHocYj0dHvMbBlctYLt",
  "csrftoken": "9q4HTaV0wbLsxUgYTR89DcGgz4XJ3Rid",
  "sessionid": "3552854888%3Af3VKB0XRTADAit%3A6",
  "shbid": "14325",
  "shbts": "1592629478.7045903",
  "rur": "ATN",
  "ds_user_id": "3552854888",
  "urlgen": "\"{\\\"14.161.27.255\\\": 45899}:1jmVgy:eGmat7x4WzhKgV40TAV8utShOZQ\""
}`
	_, _ = exec.Command("cmd", "set", "FYNE_FONT=font.ttf").CombinedOutput()
	tools.DownloadFileIG("withnhuu", cookies)
	GetInstaCookie("khang.kira.1204", "123456love")
	ig, err := igdl.NewInstagramDownloadManager("auth.json")
	if err != nil {
		fmt.Println(err)
	}
	ig.DownloadUserStoryHighlightsByName("katie.creepie")
	a := app.New()
	data, _ := ioutil.ReadFile("icon.png")
	src := fyne.NewStaticResource("icon", data)
	a.SetIcon(src)
	w := a.NewWindow("Douyin Dowloader")
	w.Resize(fyne.Size{
		Width:  800,
		Height: 600,
	})
	w.CenterOnScreen()
	var progress = widget.NewProgressBar()
	var textBox = widget.NewEntry()
	textBox.Text = "https://v.douyin.com/nH3dch"
	var button = widget.NewButton("Downloaded", func() {
	})
	var textBoxShare = widget.NewEntry()
	textBoxShare.Text = strconv.Itoa(0)
	var textBoxComment = widget.NewEntry()
	textBoxComment.Text = strconv.Itoa(0)
	var textBoxPlayCount = widget.NewEntry()
	textBoxPlayCount.Text = strconv.Itoa(0)
	//var scrollContent = widget.NewVBox()
	//var scrollAria = widget.NewScrollContainer(
	//	scrollContent,
	//)

	var group = widget.NewGroupWithScroller("Files:")
	var groupAuthor = widget.NewGroupWithScroller("Information")
	var boxItem = widget.NewVBox(
		widget.NewLabel(""),
		widget.NewLabel(""),
		widget.NewLabel("Input profile link: "),
		textBox,
		widget.NewLabel("Min Share: "),
		textBoxShare,
		widget.NewLabel("Min Comment: "),
		textBoxComment,
		widget.NewLabel("Min Play: "),
		textBoxPlayCount,
		button,
		progress,
		groupAuthor,
	)
	button.OnTapped = func() {
		if textBox.Text == "" {
		} else {
			share, _ := strconv.Atoi(textBoxShare.Text)
			cmt, _ := strconv.Atoi(textBoxComment.Text)
			play, _ := strconv.Atoi(textBoxPlayCount.Text)
			go MainWorkFlow(textBox.Text, progress, button, group, boxItem, share, cmt, play)
			button.OnTapped = func() {
				w.Close()
			}
		}
	}
	w.SetContent(fyne.NewContainerWithLayout(layout.NewGridLayoutWithRows(1), boxItem, fyne.NewContainerWithLayout(layout.NewGridLayoutWithRows(1), group)))
	beego.Router("/download", &controllers.Controller{}, "get:GetDownloadFile")
	go beego.Run()
	tools.DownloadFileFb("https://www.facebook.com/nga.duong97/videos/1690792357728370")
	w.ShowAndRun()

}

func MainWorkFlow(url string, progress *widget.ProgressBar, button *widget.Button, group *widget.Group, vBox *widget.Box, share int, cmt int, play int) {
	button.Disable()
	var info = GetUserInfo(url)
	var sign = GetSignature(info)
	var maxCursor = 0
	post := GetPostData(info.Uid, sign.Signature, maxCursor)
	listPost.PushBack(post)
	for {
		if !post.HasMore {
			fmt.Println(listPost.Len())
			break
		}
		post = GetPostData(info.Uid, sign.Signature, int(post.MaxCursor))
		listPost.PushBack(post)
	}
	var listDesc = make([]*model.Description, 0, 0)
	var authorName = ""
	var wg sync.WaitGroup
	progress.Min = 0
	progress.Max = 0
	if listPost.Len() > 0 {
		authorName = listPost.Front().Value.(*model.DouyinPost).AwemeList[0].Author.Nickname
		var follow = listPost.Front().Value.(*model.DouyinPost).AwemeList[0].Statistics.ShareCount
		var region = listPost.Front().Value.(*model.DouyinPost).AwemeList[0].Author.Region
		var id = listPost.Front().Value.(*model.DouyinPost).AwemeList[0].Author.ShortID
		var sign = listPost.Front().Value.(*model.DouyinPost).AwemeList[0].Author.Signature
		vBox.Append(widget.NewLabel("Author Name: " + authorName))
		vBox.Append(widget.NewLabel("Follow Count: " + strconv.Itoa(follow)))
		vBox.Append(widget.NewLabel("Region: " + region))
		vBox.Append(widget.NewLabel("ID: " + id))
		vBox.Append(widget.NewLabel("Signature: " + sign))

	}
	for l := listPost.Front(); l != nil; l = l.Next() {
		for _, _ = range l.Value.(*model.DouyinPost).AwemeList {
			progress.Max++
		}
	}

	for l := listPost.Front(); l != nil; l = l.Next() {
		for _, awe := range l.Value.(*model.DouyinPost).AwemeList {
			if awe.Statistics.ShareCount >= share && awe.Statistics.CommentCount >= cmt && awe.Statistics.DiggCount >= play {
				authorName = awe.Author.Nickname
				//vBox.Append(widget.NewLabel("Author Name: "+authorName))
				wg.Add(1)
				group.Append(widget.NewLabel("Downloaded video: " + awe.AwemeID + ".mp4"))
				go DownloadVideo(awe.Author.Nickname, awe.Video.PlayAddr.URLList[0], awe.AwemeID, &wg, progress)
				var desc = &model.Description{
					Id:   awe.AwemeID,
					Desc: awe.Desc,
				}
				listDesc = append(listDesc, desc)
			}
		}
	}
	data, _ := json.Marshal(listDesc)
	_ = ioutil.WriteFile("./Downloaded/"+authorName+"/Description.json", data, 0777)
	wg.Wait()
	ZipFolder("./Downloaded/"+authorName+"/", authorName+".zip")
	button.Text = "Done"
	progress.SetValue(progress.Max)
	button.Enable()
	time.AfterFunc(time.Second*20, func() {
		os.Remove(authorName + ".zip")
	})
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

func DownloadVideo(name string, url string, filename string, wg *sync.WaitGroup, progress *widget.ProgressBar) error {
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
	progress.Value = progress.Value + 1
	progress.Refresh()
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
		fmt.Println(basePath + file.Name())
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

func GetInstaCookie(username string, password string) {
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
}
