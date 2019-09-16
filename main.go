package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const PULL_DELAY = 1 //每次拉取一个page之后，延迟一段时间(单位秒)，避免反爬虫机制
const PAGE_SIZE = 20
const GROUP_URL_FMT = "https://api.zsxq.com/v1.10/groups/%d/topics?scope=%s&count=%d&end_time=%s"  //其中%s是星球GROUP_ID, 如水库
const GROUP_URL_PREFIX = "https://api.zsxq.com/v1.10/groups/%d/topics?"  //其中%s是星球GROUP_ID, 如水库
//const SCOPE = "all"
const SCOPE = "all"
const GROUP_ID_SHUIKU = 281542212511
const COOKIE_FILE = "./cookie.txt"
const TOPIC_FILE = "topics.txt"

type Owner struct {
	User_id uint64 	`json:"user_id"`  //提问者id
	Name string 	`json:"name"` //提问者名字
}

type ImageInfo struct {
	Url string 		`json:"url"`
	Width uint64 	`json:"width"`
	Height uint64 	`json:"height"`
	Size uint64		`json:"size"`
}

type Image struct {
	Image_id uint64 	`json:"image_id"`
	Type string 		`json:"type"`
	Thumbnail ImageInfo `json:"thumbnail"`
	Large ImageInfo 	`json:"large"`
	Original ImageInfo	`json:"original"`
}

type Question struct {
	Owner Owner 		`json:"owner"` //提问者
	Text string 		`json:"text"` //问题内容
	Images []Image 		`json:"images"` //问题图片
}

type Answer struct {
	Owner Owner 	`json:"owner"` //回答者
	Text string 	`json:"text"` //回答内容
	Images []Image 		`json:"images"` //回答图片
}

type ShowComment struct {
	Comment_id uint64 	`json:"comment_id"`
	Create_time string 	`json:"create_time"`
	Owner Owner 		`json:"owner"`
	Text string 		`json:"text"` //评论内容
}

//话题
type Topic struct {
	Topic_id uint64 	`json:"comment_id"`  //话题ID
	Question Question 	`json:"question"` //提问
	Answer Answer 		`json:"answer"` //回答
	Show_comments []ShowComment 	`json:"show_comments"` //评论列表
	Likes_count uint64 	`json:"likes_count"` //点赞数
	Comments_count uint64 	`json:"comments_count"` //评论数
	Reading_count uint64 `json:"reading_count"` //阅读数
	Create_time string 	`json:"create_time"` //创建时间2019-09-11T00:16:13.099+0800
}

type RespData struct {
	Topics []Topic `json:"topics"`
}
//url响应数据
type Resp struct {
	Succeeded bool `json:"succeeded"`
	Resp_data RespData `json:"resp_data"`
}


//type Cookie struct {
//	domain string `json:"domain"`
//	expirationDate float32 `json:"expirationDate"`
//	hostOnly bool `json:"hostOnly"`
//	path string `json:"path"`
//	sameSite string `json:"sameSite"`
//	secure bool `json:"secure"`
//	session bool `json:"session"`
//	storeId string `json:"storeId"`
//	id int `json:"id"`
//	name string `json:"name"`
//	value string `json:"value"`
//}

type Cookie struct {
	Name string `json:"name"`
	Value string `json:"value"`
}

type JsonStruct struct {
}

func NewJsonStruct() *JsonStruct {
	return &JsonStruct{}
}

//读取cookie json文件
func  LoadJsonFile(filename string, v interface{}) {
	//ReadFile函数会读取文件的全部内容，并将结果以[]byte类型返回
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("read file %v failed", filename)
	}

	err = json.Unmarshal(data, v)
	if err != nil {
		log.Fatal("json Unmarshal failed, %v", err)
	}
}
//获取知识星球cookie
func getZsxqCookie() ([]Cookie) {

	cookies := []Cookie{}

	LoadJsonFile(COOKIE_FILE, &cookies)
	return cookies
}

//将get请求的参数进行转义
func getParseParam(param string) string {
	return url.PathEscape(param)
}

func timeToStr(t time.Time) (string) {
	format := "2006-01-02T15:04:05.000+0800"
	return t.Format(format)
}

func strToTime(s string) (time.Time) {
	format := "2006-01-02T15:04:05.000+0800"
	t, _ := time.Parse(format, s)
	return t
}


//http api获取topic
func getOnePageTopic(endTime time.Time, pageSize int) ([]Topic) {

	client := http.Client{}
	format := "2006-01-02T15:04:05.000+0800"
	endTimeStr := endTime.Format(format)


	//urlVal := fmt.Sprintf(GROUP_URL_FMT, GROUP_ID_SHUIKU, SCOPE, pageSize, endTimeStr)
	//fmt.Printf("urlval=%v", urlVal)
	//urlArr := strings.Split(urlVal,"?")
	//if len(urlArr)  == 2 {
	//	urlVal = urlArr[0] + "?" + getParseParam(urlArr[1])
	//}

	urlVal := fmt.Sprintf(GROUP_URL_PREFIX, GROUP_ID_SHUIKU)
	values := url.Values{}
	values.Add("scope", SCOPE)

	values.Add("count", strconv.Itoa(pageSize))
	values.Add("end_time", endTimeStr)

	urlVal = urlVal + values.Encode()

	var req *http.Request

	req, _ = http.NewRequest("GET", urlVal, nil)

	//获取cookie
	cookies := getZsxqCookie()

	fmt.Printf("urlval=%v, cookie=%v \n", urlVal, cookies)

	for i := 0; i < len(cookies); i++ {
		req.AddCookie(&http.Cookie{Name:cookies[i].Name, Value:cookies[i].Value})
	}
	//添加， User-Agent否则会被认为是爬虫
	req.Header.Add("User-Agent","Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/76.0.3809.132 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		//url 请求失败
		log.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("url resp raw \n", string(b))

	respValue := Resp{}
	json.Unmarshal(b, &respValue)
	fmt.Println("url resp Unmarshal:", respValue)

	if respValue.Succeeded != true {
		//url api逻辑失败
		log.Fatal("api return not success")
		return nil
	}

	return respValue.Resp_data.Topics
}

//写入topic到文件
func saveTopicToFile(topics []Topic, filename string) () {
	fd, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	if err != nil {
		log.Fatal("saveTopic | open file err:%v", err)
		return
	}
	defer fd.Close()
	for i := 0; i < len(topics); i++ {
		topicBytes, err := json.Marshal(topics[i])
		if err != nil {
			log.Fatal("json marshal topics failed, err=%v", err)
			return
		}
		//TODO  保存到mysql，或者文件，或者都存储
		fmt.Printf("%v \n", string(topicBytes[:]))
		fd.Write(topicBytes)
		fd.Write([]byte("\n"))
	}

}

//获取topic列表的信息
//返回：
// lastTopicTime： 列表中最后一个topic创建时间
// returnPageSize: 列表长度
func getPageInfoFromTopics(topics []Topic) (time.Time, int) {

	topicCnt := len(topics)
	//创建时间格式2019-09-11T00:16:13.099+0800
	format := "2006-01-02T15:04:05.000+0800"
	//fmt.Printf("len=%v", topicCnt)
	lastTopicTime, _ := time.Parse(format, topics[topicCnt - 1].Create_time)

	return lastTopicTime, topicCnt
}

//从topic文件中获取最后一次拉取topic的create_time，作为lastTopicTime
func getLastTopicTime(filename string) (time.Time, error) {
	var lastTopicTime time.Time
	f, err := os.Open(filename)
	defer f.Close()
	var line string
	var lastLine string
	if nil == err {
		buff := bufio.NewReader(f)
		for {
			line, err = buff.ReadString('\n')
			if err != nil || io.EOF == err{
				break
			}
			//fmt.Println(line)
			//最后一行
			lastLine = line
		}
	} else {
		if os.IsNotExist(err) {
			//topic文件不存在, 则以当前时间开始拉取
			lastTopicTime = time.Now()
			return lastTopicTime, nil
		} else {
			//其他错误
			return lastTopicTime, err
		}


		return lastTopicTime, err
	}

	topic := Topic{}
	err = json.Unmarshal([]byte(lastLine), &topic)
	if err != nil {
		log.Fatal("getLastTopicTime| json Unmarshal err: %v", err)
		return lastTopicTime, err
	}
	lastTopicTime = strToTime(topic.Create_time)
	return lastTopicTime, nil
}


func main() {

	//TODO:从上次拉取的地方继续拉取
	lastTopicTime, err := getLastTopicTime(TOPIC_FILE)
	if err != nil {
		log.Fatal("getLastTopicTime| err:%v", err)
		return
	}
	//lastTopicTime := time.Now()
	returnPageSize := PAGE_SIZE

	for {
		fmt.Printf("lastTopicTime=%v, returnPageSize=%v \n", timeToStr(lastTopicTime), returnPageSize)
		if (returnPageSize < PAGE_SIZE) {
			//如果条数不足PAGE_SIZE说明拉取完了
			//TODO 加一些统计，总数，失败数等等
			fmt.Printf("Done。returnPageSize = %v less than PAGE_SIZE=%v, lastPageTime=%v。",
				returnPageSize, PAGE_SIZE, lastTopicTime)
			break
		}

		topicList := getOnePageTopic(lastTopicTime, PAGE_SIZE)

		go saveTopicToFile(topicList, TOPIC_FILE)
		//TODO: 获取最后一个topic的create_time时间作为下次拉取的lastTopicTime，以及page中条数
		lastTopicTime, returnPageSize = getPageInfoFromTopics(topicList)

		//每拉取一页，延迟1秒
		time.Sleep(time.Second * PULL_DELAY)
	}




}
