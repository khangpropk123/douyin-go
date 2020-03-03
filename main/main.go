package main

import (
	"./model"
	"container/list"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)
var listPost = list.New()
func main(){
	var timenow = time.Now()
	var info = GetUserInfo("https://www.iesdouyin.com/share/user/24058267831")
	var sign = GetSignature(info)
	var maxCursor = 0
	post := GetPostData(info.Uid,sign.Signature,maxCursor)
	listPost.PushBack(post)
	for{

		if !post.HasMore {
			fmt.Println(listPost.Len())
			break
		}
		fmt.Println("Tiep ne")
		post = GetPostData(info.Uid,sign.Signature,int(post.MaxCursor))
		listPost.PushBack(post)
	}
	var listDesc = make([]*model.Description,0,0)
	var authorName = ""
	var wg sync.WaitGroup
	for l := listPost.Front(); l!=nil;l=l.Next() {
		for _, awe := range l.Value.(*model.DouyinPost).AwemeList{
			authorName = awe.Author.Nickname
			wg.Add(1)
			go DownloadVideo(awe.Author.Nickname,awe.Video.PlayAddr.URLList[0],awe.AwemeID,&wg)
			var desc = &model.Description{
				Id:   awe.AwemeID,
				Desc: awe.Desc,
			}
			listDesc = append(listDesc,desc)
		}

	}
	data ,_ := json.Marshal(listDesc)
	_ = ioutil.WriteFile("./Downloaded/"+authorName+"/Description.json", data, 0777)

	wg.Wait()
	fmt.Println(time.Now().Sub(timenow))
}

func GetUserInfo(url string) *model.SignaturePost  {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; U; Android 5.1.1; zh-cn; MI 4S Build/LMY47V) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/53.0.2785.146 Mobile Safari/537.36 XiaoMi/MiuiBrowser/9.1.3")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body,_ := ioutil.ReadAll(resp.Body)
	tacRe := regexp.MustCompile(`tac='([\s\S]*?)'</script>`)
	uidRe := regexp.MustCompile(`uid: "(\d+?)"`)

	var uid = strings.ReplaceAll(strings.ReplaceAll(uidRe.FindString(string(body)),`uid: "`,""),`"`,"")
	var tac = strings.ReplaceAll(strings.Split(tacRe.FindString(string(body)),"|")[0],`tac='`,"")
	var sign = &model.SignaturePost{
		Tac: tac,
		Uid: uid,
	}
	fmt.Println(sign)
	return sign
}

func GetSignature(sign *model.SignaturePost) *model.SignatureResp {
	data := url.Values{}
	data.Set("tac",sign.Tac)
	data.Set("user_id",sign.Uid)
	r, _ := http.NewRequest("POST", "http://127.0.0.1:5000/", strings.NewReader(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	//data,_ := json.Marshal(sign)
	//fmt.Println(string(data))

	resp,err := http.DefaultClient.Do(r)
	if err!=nil {
		return nil
	}
	defer resp.Body.Close()
	body,_ := ioutil.ReadAll(resp.Body)
	var signResp = &model.SignatureResp{}
	json.Unmarshal(body,signResp)
	return signResp
}

func GetPostData(uid string,sign string, maxCursor int) *model.DouyinPost {
	var baseUrl = fmt.Sprintf(`https://www.iesdouyin.com/web/api/v2/aweme/post/?user_id=%s&sec_uid=&count=21&max_cursor=%d&aid=1128&_signature=%s`,uid,maxCursor,sign)
	fmt.Println(baseUrl)
	req, err := http.NewRequest("GET",baseUrl,nil)
	if err != nil{
		return nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; U; Android 5.1.1; zh-cn; MI 4S Build/LMY47V) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/53.0.2785.146 Mobile Safari/537.36 XiaoMi/MiuiBrowser/9.1.3")
	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var name  = &model.DouyinPost{}
	err = json.Unmarshal(body, &name)
	fmt.Println(err)
	return name
}

func DownloadVideo(name string,url string, filename string,wg *sync.WaitGroup) error  {
	fmt.Println("Start downloading",filename+".mp4")
	defer wg.Done()
	rootPath := "./Downloaded/"+name
	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		os.Mkdir(rootPath,777)
	}
	req,_ := http.NewRequest("GET",url,nil)
	req.Header.Set("User-Agent","Mozilla/5.0 (Linux; U; Android 5.1.1; zh-cn; MI 4S Build/LMY47V) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/53.0.2785.146 Mobile Safari/537.36 XiaoMi/MiuiBrowser/9.1.3")
	resp, err := http.DefaultClient.Do(req)
	if err!=nil {
		return err
	}
	defer resp.Body.Close()
	buffer,err :=ioutil.ReadAll(resp.Body)
	if err !=nil{
		return err
	}
	err = ioutil.WriteFile(rootPath+"/"+filename+".mp4",buffer,777)
	if err != nil {
		return err
	}
	fmt.Println(filename+".mp4","downloaded!")
	return nil
}