package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
	"models"
)

const PULL_DELAY = 100 //每次拉取一个page之后，延迟一段时间(单位毫秒)，避免反爬虫机制
const PAGE_SIZE = 20
//const GROUP_URL_FMT = "https://api.zsxq.com/v1.10/groups/%d/topics?scope=%s&count=%d&end_time=%s"  //其中%s是星球GROUP_ID, 如水库
const GROUP_URL_PREFIX = "https://api.zsxq.com/v1.10/groups/%d/topics?"  //其中%s是星球GROUP_ID, 如水库
//const SCOPE = "all"
const SCOPE = "all"
const GROUP_ID_SHUIKU = 281542212511
const COOKIE_FILE = "./cookie.txt"
const GROUP_DIR_FMT = "group_%d"
const TOPIC_FILE = "topics.txt"
const IMAGE_DIR = "images"
const IMAGE_PREFIX = "large_"
const USER_AGENT = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/76.0.3809.132 Safari/537.36"
const MOST_LONG_LONG_AGO = "2006-01-02T15:04:05.000+08000"
const TOPIC_SAVE_TYPE_BIT_FILE = 1   //bitmap, 0位 存储topics到文件
const TOPIC_SAVE_TYPE_BIT_MYSQL = 2   //bitmap, 1位 存储topics到mysql

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

var db *gorm.DB
var topicSaveType int

func init() {
	var err error
	//链接mysql
	db, err = gorm.Open("mysql", "root:900616/zsxq?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
	if db.HasTable(models.Topics{}) {
		//判断表是否存在, 不存在就创建
		db.CreateTable(models.Topics{})
	}

	topicSaveType = TOPIC_SAVE_TYPE_BIT_MYSQL
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

//cookie 转字符串
func convCookieToStr(cookies []Cookie) (string){
	var cookieStr string
	firstFlag := true
	for i := 0; i < len(cookies); i++ {
		if firstFlag == true {
			cookieStr += cookies[i].Name + "=" + cookies[i].Value
			firstFlag = false
		} else {
			cookieStr += ";" + cookies[i].Name + "=" + cookies[i].Value
		}
	}
	return cookieStr
}

//http get url
func httpGetWithCookie(urlStr string) []byte {
	client := http.Client{}
	var req *http.Request
	//获取cookie
	cookies := getZsxqCookie()
	req, _ = http.NewRequest("GET", urlStr, nil)

	//cookie
	for i := 0; i < len(cookies); i++ {
		req.AddCookie(&http.Cookie{Name:cookies[i].Name, Value:cookies[i].Value})
	}
	//header
	//添加 User-Agent否则会被认为是爬虫
	req.Header.Add("User-Agent",USER_AGENT)
	cookieStr := convCookieToStr(cookies)

	curlStr := fmt.Sprintf("curl \"%s\" --cookie \"%s\" -H \"User-Agent: %s\" ", urlStr, cookieStr, USER_AGENT)
	fmt.Printf("%s \n", curlStr)

	//发起请求
	resp, err := client.Do(req)
	if err != nil {
		//url 请求失败
		log.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)
	return b
}


//http api获取topic
func getOnePageTopic(groupId int, endTime time.Time, pageSize int) ([]Topic) {
	format := "2006-01-02T15:04:05.000+0800"
	endTimeStr := endTime.Format(format)


	//urlVal := fmt.Sprintf(GROUP_URL_FMT, GROUP_ID_SHUIKU, SCOPE, pageSize, endTimeStr)
	//fmt.Printf("urlval=%v", urlVal)
	//urlArr := strings.Split(urlVal,"?")
	//if len(urlArr)  == 2 {
	//	urlVal = urlArr[0] + "?" + getParseParam(urlArr[1])
	//}

	//拼接url
	urlStr := fmt.Sprintf(GROUP_URL_PREFIX, groupId)
	values := url.Values{}
	values.Add("scope", SCOPE)

	values.Add("count", strconv.Itoa(pageSize))
	values.Add("end_time", endTimeStr)
	urlStr = urlStr + values.Encode()


	respBytes := httpGetWithCookie(urlStr)

	respValue := Resp{}
	json.Unmarshal(respBytes, &respValue)
	fmt.Println("url resp Unmarshal:", respValue)

	if respValue.Succeeded != true {
		//url api逻辑失败
		log.Fatal("api return not success")
		return nil
	}

	return respValue.Resp_data.Topics
}

// 判断文件夹是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}


//保存图片到文件
func saveImageToFile(waitGroup *sync.WaitGroup, image Image, imageDir string) {
	//当前只用large图
	urlStr := image.Large.Url
	//imageFileName :=     IMAGE_PREFIX + image.Image_id + "." + image.Type
	imageFileName := fmt.Sprintf("%s/%s%d.%s", imageDir, IMAGE_PREFIX, image.Image_id, image.Type)
	fmt.Printf("saveImageToFile |url=%s, filename=%s \n", urlStr, imageFileName)

	//拉取image
	imageData := httpGetWithCookie(urlStr)

	imageFd, err := os.Create(imageFileName)
	if err != nil {
		log.Fatal("create file failed:", imageFileName, err)
		return
	}
	defer imageFd.Close()
	imageFd.Write(imageData)

	waitGroup.Done()
}

//从topic获取图片列表
func getImageUrlFromTopic(topic Topic) ([]Image){
	var imageList []Image
	//问题图片 + 回答图片
	for i := 0; i < len(topic.Question.Images); i++ {
		imageList = append(imageList, topic.Question.Images[i])
	}
	for i := 0; i < len(topic.Answer.Images); i++ {
		imageList = append(imageList, topic.Answer.Images[i])
	}

	return imageList
}

//保存图片到目录
func saveImage(imageList []Image, imageDir string) {
	waitGroup := sync.WaitGroup{}
	for i := 0; i < len(imageList); i++ {
		fmt.Printf("image =%v \n", imageList)
		waitGroup.Add(1)
		//从url拉取图片到文件夹
		go saveImageToFile(&waitGroup, imageList[i], imageDir)
	}
	waitGroup.Wait()
}

func saveTopicToFile(topics []Topic, filename string, imageDir string) {
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
		//保存到者文件，或者都存储
		fmt.Printf("%v \n", string(topicBytes[:]))
		fd.Write(topicBytes)
		fd.Write([]byte("\n"))


		//获取图片列表
		imageList := getImageUrlFromTopic(topics[i])
		//图片保存到文件夹
		saveImage(imageList, imageDir)
	}
}

func saveTopicToMysql(topics []Topic) {

}

//写入topic到文件
func saveTopic(waitGroup *sync.WaitGroup, topics []Topic, filename string, imageDir string) () {

	//保存到mysql
	if topicSaveType | TOPIC_SAVE_TYPE_BIT_MYSQL != 0 {
		saveTopicToMysql(topics)
	}
	//保存到文件
	if topicSaveType | TOPIC_SAVE_TYPE_BIT_FILE != 0  {
		saveTopicToFile(topics, filename, imageDir)
	}

	waitGroup.Done()
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

func createDir(dir string) {
	exist, err := PathExists(dir)
	if err != nil {
		fmt.Printf("get dir error![%v]\n", err)
		return
	}

	if exist {
		fmt.Printf("has dir![%v]\n", dir)
	} else {
		fmt.Printf("no dir![%v]\n", dir)
		// 创建文件夹
		err := os.Mkdir(dir, os.ModePerm)
		if err != nil {
			fmt.Printf("mkdir failed![%v]\n", err)
		} else {
			fmt.Printf("mkdir %s success!\n", dir)
		}
	}
}

//拉取指定时间段的topic
func pullZsxqTopicByPeriod(groupId int, startTime time.Time, endTime time.Time) {
	//创建group目录
	groupDir := fmt.Sprintf(GROUP_DIR_FMT, groupId)
	createDir(groupDir)
	//创建图片目录
	imageDir := groupDir + "/" + IMAGE_DIR
	createDir(imageDir)

	topicFile := groupDir + "/" + TOPIC_FILE
	waitGroup := sync.WaitGroup{}

	//从上次拉取的地方继续拉取
	//lastTopicTime, err := getLastTopicTime(topicFile)
	//if err != nil {
	//	log.Fatal("getLastTopicTime| err:%v", err)
	//	return
	//}
	//lastTopicTime := time.Now()
	returnPageSize := PAGE_SIZE

	for {
		fmt.Printf("lastTopicTime=%v, returnPageSize=%v \n", timeToStr(startTime), returnPageSize)
		if (returnPageSize < PAGE_SIZE || startTime.Before(endTime)) {
			//结束拉取
			//1. 如果条数不足PAGE_SIZE说明拉取完了。
			//2. 如果startTime在endTime之前，说明这个时间段拉取完了。如startTime=00:14:00 , endTime=00:14:01, 应该停止
			//TODO： 加一些统计，总数，失败数等等
			fmt.Printf("Done。returnPageSize = %v less than PAGE_SIZE=%v, lastPageTime=%v \n",
				returnPageSize, PAGE_SIZE, startTime)
			break
		}
		//注意每次lastTopicTime都需要加一个微小的时间偏移，比如1s
		topicList := getOnePageTopic(groupId, startTime.Add(-1 * time.Second), PAGE_SIZE)
		waitGroup.Add(1)
		go saveTopic(&waitGroup, topicList, topicFile, imageDir)
		//获取最后一个topic的create_time时间作为下次拉取的lastTopicTime，以及page中条数
		startTime, returnPageSize = getPageInfoFromTopics(topicList)

		//每拉取一页，延迟一定的ms
		time.Sleep(time.Millisecond * PULL_DELAY)
	}

	fmt.Printf("waiting pull finish....\n")
	waitGroup.Wait()
}

//获取已经存在的时间段
func getExistPeriod(groupId int) (time.Time, time.Time) {

}

//拉取指定group_id星球的topic
//应当拉取的时间段，分为头尾2段时间：
// 1. [当前时间, 已拉取的最新时间]
// 2. [已拉取的最新时间, 2000-01-01 00:00:00]
func pullZsxqTopic(groupId int) {

	//获取当前已拉取的时间段
	existStartTime, existEndTime := getExistPeriod(groupId)

	if existStartTime == nil || existEndTime == nil {
		//未存在已拉取数据，说明需要从now到MOST_LONG_LONG_AGO
		pullZsxqTopicByPeriod(groupId, time.Now(), strToTime(MOST_LONG_LONG_AGO))
	} else {
		//有存量数据
		//1. 拉取前半段
		pullZsxqTopicByPeriod(groupId, time.Now(), existStartTime)

		//2. 拉取后半段
		pullZsxqTopicByPeriod(groupId, existEndTime, strToTime(MOST_LONG_LONG_AGO))
	}



}
//拉取指定group_id星球的topic
//func pullZsxqTopic(groupId int) {
//	//创建group目录
//	groupDir := fmt.Sprintf(GROUP_DIR_FMT, groupId)
//	createDir(groupDir)
//	//创建图片目录
//	imageDir := groupDir + "/" + IMAGE_DIR
//	createDir(imageDir)
//
//	topicFile := groupDir + "/" + TOPIC_FILE
//	waitGroup := sync.WaitGroup{}
//
//	//从上次拉取的地方继续拉取
//	lastTopicTime, err := getLastTopicTime(topicFile)
//	if err != nil {
//		log.Fatal("getLastTopicTime| err:%v", err)
//		return
//	}
//	//lastTopicTime := time.Now()
//	returnPageSize := PAGE_SIZE
//
//	for {
//		fmt.Printf("lastTopicTime=%v, returnPageSize=%v \n", timeToStr(lastTopicTime), returnPageSize)
//		if (returnPageSize < PAGE_SIZE) {
//			//如果条数不足PAGE_SIZE说明拉取完了
//			//加一些统计，总数，失败数等等
//			fmt.Printf("Done。returnPageSize = %v less than PAGE_SIZE=%v, lastPageTime=%v \n",
//				returnPageSize, PAGE_SIZE, lastTopicTime)
//			break
//		}
//		//注意每次lastTopicTime都需要加一个微小的时间偏移，比如1s
//		topicList := getOnePageTopic(groupId, lastTopicTime.Add(-1 * time.Second), PAGE_SIZE)
//		waitGroup.Add(1)
//		go saveTopicToFile(&waitGroup, topicList, topicFile, imageDir)
//		//获取最后一个topic的create_time时间作为下次拉取的lastTopicTime，以及page中条数
//		lastTopicTime, returnPageSize = getPageInfoFromTopics(topicList)
//
//		//每拉取一页，延迟1秒
//		time.Sleep(time.Millisecond * PULL_DELAY)
//	}
//
//	fmt.Printf("waiting pull finish....\n")
//	waitGroup.Wait()
//}


func main() {





	//拉取水库 知识星球
	pullZsxqTopic(GROUP_ID_SHUIKU)
}
