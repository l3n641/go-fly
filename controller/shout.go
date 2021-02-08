package controller

import (
	"encoding/json"
	"fmt"
	"github.com/taoshihan1991/imaptool/models"
	"github.com/taoshihan1991/imaptool/tools"
	"log"
	"strconv"
	"time"
)

func SendServerJiang(content string) string {
	noticeServerJiang, err := strconv.ParseBool(models.FindConfig("NoticeServerJiang"))
	serverJiangAPI := models.FindConfig("ServerJiangAPI")
	if err != nil || !noticeServerJiang || serverJiangAPI == "" {
		log.Println("do not notice serverjiang:", serverJiangAPI, noticeServerJiang)
		return ""
	}
	sendStr := fmt.Sprintf("%s,访客来了", content)
	desp := "[登录](https://gofly.sopans.com/main)"
	url := serverJiangAPI + "?text=" + sendStr + "&desp=" + desp
	//log.Println(url)
	res := tools.Get(url)
	return res
}
func SendNoticeEmail(username, msg string) {
	smtp := models.FindConfig("NoticeEmailSmtp")
	email := models.FindConfig("NoticeEmailAddress")
	password := models.FindConfig("NoticeEmailPassword")
	if smtp == "" || email == "" || password == "" {
		return
	}
	err := tools.SendSmtp(smtp, email, password, []string{email}, "[通知]"+username, msg)
	if err != nil {
		log.Println(err)
	}
}
func SendAppGetuiPush(kefu string, title, content string) {
	token := models.FindConfig("GetuiToken")
	if token == "" {
		token = getGetuiToken()
		if token == "" {
			return
		}
	}
	appid := models.FindConfig("GetuiAppID")
	format := `
{
    "request_id":"%s",
    "settings":{
        "ttl":3600000
    },
    "audience":{
        "cid":[
            "%s"
        ]
    },
    "push_message":{
        "notification":{
            "title":"%s",
            "body":"%s",
            "click_type":"url",
            "url":"https//:xxx"
        }
    }
}
`
	clients := models.FindClients(kefu)
	if len(clients) == 0 {
		return
	}
	//clientIds := make([]string, 0)
	for _, client := range clients {
		//clientIds = append(clientIds, client.Client_id)
		req := fmt.Sprintf(format, tools.Md5(tools.Uuid()), client.Client_id, title, content)
		url := "https://restapi.getui.com/v2/" + appid + "/push/single/cid"
		headers := make(map[string]string)
		headers["Content-Type"] = "application/json;charset=utf-8"
		headers["token"] = token
		res, err := tools.PostHeader(url, []byte(req), headers)
		log.Println(url, req, err, res)
	}

}
func getGetuiToken() string {
	appid := models.FindConfig("GetuiAppID")
	appkey := models.FindConfig("GetuiAppKey")
	//appsecret := models.FindConfig("GetuiAppSecret")
	appmastersecret := models.FindConfig("GetuiMasterSecret")
	type req struct {
		Sign      string `json:"sign"`
		Timestamp string `json:"timestamp"`
		Appkey    string `json:"appkey"`
	}
	timestamp := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	reqJson := req{
		Sign:      tools.Sha256(appkey + timestamp + appmastersecret),
		Timestamp: timestamp,
		Appkey:    appkey,
	}
	reqStr, _ := json.Marshal(reqJson)
	url := "https://restapi.getui.com/v2/" + appid + "/auth"
	res, err := tools.Post(url, "application/json;charset=utf-8", reqStr)
	log.Println(url, string(reqStr), err, res)
	if err == nil && res != "" {
		var pushRes GetuiResponse
		json.Unmarshal([]byte(res), &pushRes)
		if pushRes.Code == 0 {
			token := pushRes.Data["token"].(string)
			models.UpdateConfig("GetuiToken", token)
			return token
		}
	}
	return ""
}
