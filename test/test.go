package main

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"time"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type Cookie struct {
	domain string `json:"domain"`
	expirationDate float32 `json:"expirationDate"`
	hostOnly bool `json:"hostOnly"`
	path string `json:"path"`
	sameSite string `json:"sameSite"`
	secure bool `json:"secure"`
	session bool `json:"session"`
	storeId string `json:"storeId"`
	id int `json:"id"`
	name string `json:"name"`
	value string `json:"value"`
}
type Topics struct {
	ID 				int64 		`gorm:"primary_key"`
	GroupId 		int64		`gorm:"type:bigint(20);not null;unique_index:idx_group_time_id;"`
	TopicContent 	string 		`gorm:"type:text;not null;"`
	UpdateAt 		time.Time	`gorm:"type:timestamp"`
	TopicTime 		time.Time 	`gorm:"type:datetime(3);unique_index:idx_group_time_id;"`
	CreateAt 		time.Time	`gorm:"type:timestamp"`
}
//话题
type Topic struct {
	Question Question 	`json:"question"` //提问
	Answer Answer 		`json:"answer"` //回答
	Show_comments []ShowComment 	`json:"show_comments"` //评论列表
	Likes_count int64 	`json:"likes_count"` //点赞数
	Comments_count int64 	`json:"comments_count"` //评论数
	Reading_count int64 `json:"reading_count"` //阅读数
	Create_time string 	`json:"create_time"` //创建时间2019-09-11T00:16:13.099+0800
}

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
	Comment_id int64 	`json:"comment_id"`
	Create_time string 	`json:"create_time"`
	Owner Owner 		`json:"owner"`
	Text string 		`json:"text"` //评论内容
}
//func main() {
//	cookieStr := `[
//{
//    "domain": ".zsxq.com",
//    "expirationDate": 1568431615.21024,
//    "hostOnly": false,
//    "httpOnly": false,
//    "name": "abtest_env",
//    "path": "/",
//    "sameSite": "unspecified",
//    "secure": false,
//    "session": false,
//    "storeId": "0",
//    "value": "product",
//    "id": 1
//},
//{
//    "domain": ".zsxq.com",
//    "expirationDate": 1583915272,
//    "hostOnly": false,
//    "httpOnly": false,
//    "name": "UM_distinctid",
//    "path": "/",
//    "sameSite": "unspecified",
//    "secure": false,
//    "session": false,
//    "storeId": "0",
//    "value": "16d1f6fa00eec0-0d8bbe2bdcc961-38637501-1aeaa0-16d1f6fa00f435",
//    "id": 2
//}]`
//	cookies := []Cookie{}
//	json.Unmarshal([]byte(cookieStr), cookies)
//
//	fmt.Printf("cookies %v", cookies)
//
//}
const TIME_ZONE_SHANGHAI 	= "Asia/Shanghai"
const TIME_ZONE_0 			= ""
var db *gorm.DB

func init() {
	var err error
	//链接mysql
	db, err = gorm.Open("mysql", "root:root@tcp(127.0.0.1:3306)/zsxq?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}

}

func timeToStr(t time.Time, timeZone string) (string) {
	format := "2006-01-02T15:04:05.000+0800"
	//time.ParseInLocation("2006-01-02 15:04:05", t, time.Local)
	//var cstSh, _ = time.LoadLocation("Asia/Shanghai") //上海
	//var cstSh, _ = time.LoadLocation(timeZone) //0时区
	//return t.In(cstSh).Format(format)
	return t.Format(format)
}

func strToTime(s string, timeZone string) (time.Time) {
	format := "2006-01-02T15:04:05.000+0800"
	t, _ := time.Parse(format, s)
	//var cst0, _ = time.LoadLocation(timeZone) //0时区
	//t, _ := time.ParseInLocation(format, s, cst0)
	return t
}


func main ( ) {
	var jsonBlob = [ ] byte ( ` [
        { "Name" : "Platypus" , "Order" : "Monotremata" } ,
        { "Name" : "Quoll" ,     "Order" : "Dasyuromorphia" }
    ] ` )
	type Animal struct {
		//Name  string
		Order string
	}
	var animals [ ] Animal
	err := json.Unmarshal ( jsonBlob , & animals )
	if err != nil {
		fmt.Println ( "error:" , err )
	}
	fmt.Printf ( "%+v" , animals )

	testTime := time.Time{}

	fmt.Printf("%v\n", testTime.IsZero())


	//timeStr := "2019-09-29T17:02:01.335+0800"
	//fmt.Printf("time %v\n", timeToStr(strToTime(timeStr)))

	timeStr := timeToStr(time.Now(), TIME_ZONE_0)
	fmt.Printf("time %v \n", timeStr)
	timeObj := strToTime(timeStr, TIME_ZONE_0)
	fmt.Printf("time %v \n", timeToStr(timeObj, TIME_ZONE_0))

	emojiStr := `{"question":{"owner":{"user_id":0,"name":""},"text":"欧神，您好！看到之前有人问“亲戚入股买房”，您说1.债权融资，2.把对方当仇人，写的清清楚楚。请教您：如果入股权，该写清楚哪几样协议？","images":null},"answer":{"owner":{"user_id":225442182841,"name":"yevon_ou"},"text":"最关键二条；\n\n1) 如果想卖，到哪个价位卖，谁说了算\n\n2) 如果想撤股，如何处理\n\n\n\n\n1½) 如果要装修，民宿，维修，谁说了算。合同细则谁能拍板。","images":null},"show_comments":[{"comment_id":554582154854,"create_time":"2017-04-30T15:54:26.871+0800","owner":{"user_id":222211288481,"name":"azorro"},"text":"（1）划清边界；（2）禁止互相购买，一致对外；（3）约定退出条件；（4）调整自我心态，做好吃亏的准备；（5）手续齐全"},{"comment_id":452514851218,"create_time":"2017-04-30T15:54:42.832+0800","owner":{"user_id":222211288481,"name":"azorro"},"text":"血的教训"},{"comment_id":824215258582,"create_time":"2017-04-30T17:25:48.549+0800","owner":{"user_id":455511518428,"name":"信步小园闲庭"},"text":"说的太对了，现在和亲戚一起投的钱退也退不出，只能过桥+抵押贷。"},{"comment_id":141482421222,"create_time":"2017-04-30T17:37:35.336+0800","owner":{"user_id":481418411128,"name":"zhiwen"},"text":"福州大神吗😁"},{"comment_id":824218855222,"create_time":"2017-04-30T22:44:23.781+0800","owner":{"user_id":458511554248,"name":"郑钦文 Enzo"},"text":"GP和LP要分清楚"}],"likes_count":11,"comments_count":5,"reading_count":118,"create_time":"2017-04-30T15:49:11.423+0800"}`

	topic := Topic{}

	json.Unmarshal([]byte(emojiStr), &topic)
	fmt.Printf("json1 :%v\n", topic)
	jsonBytes, _ := json.Marshal(topic)
	fmt.Printf("json2 :%v\n", string(jsonBytes[:]))

	var topics []Topics
	db.Where("topic_content like ?", "%福州大神吗%").Find(&topics)
	fmt.Printf("-----topics %v\n", topics)
}