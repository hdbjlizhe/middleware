package middleware

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/httplib"
	"github.com/buger/jsonparser"
)

// /////////////////////////////////////////////////////////////////////////////////////////////////////
// /////////////////////////////////////////////////////////////////////////////////////////////////////
// /////////////////////////////////////////////////////////////////////////////////////////////////////
var Port string

const socketPath = "/tmp/autMan.sock"

var transport = &http.Transport{
	DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
		return net.Dial("unix", socketPath)
	},
}

func httpUrl() string {
	return "http://127.0.0.1:" + Port + "/otto"
}

func sockUrl() string {
	return "http://127.0.0.1/sock"
}

/**
 * @description: 设置端口号
 */
func SetPort() {
	Port = os.Args[1]
}

/**
 * @description: 添加消息监听句柄
 * @param {string} chatid 群组ID
 * @param {string} userid 用户ID
 * @param {func(string)} func 消息监听句柄，回调函数
 */
func AddMsgListener(imtype, chatid, userid string, exitChannel chan struct{}, function func(string)) {
	//创建ess连接
	url := fmt.Sprintf("%s/msghook", httpUrl())

	// 向SSE服务端发起POST请求
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}
	// 设置Header
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(strings.NewReader(fmt.Sprintf(`{"imtype":"%s","chatid":"%s","userid":"%s"}`, imtype, chatid, userid)))

	// 发起请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// 使用bufio.NewReader来按行读取服务端发送的数据
	reader := bufio.NewReader(resp.Body)

	go func() {
		for {
			// 按行读取数据
			data, err := reader.ReadBytes('\n')
			if err != nil {
				fmt.Printf("Read error: %v\n", err)
				break
			} else {
				fmt.Printf("Read: %s\n", data)
			}
			// 这里简单地打印出来，实际应用中可能需要根据行的内容进行解析
			msg := strings.TrimSuffix(string(data), "\n")
			fmt.Printf("Received: %s", msg)
			if strings.Contains(msg, "event:message") {
				continue
			} else {
				msg = strings.TrimLeft(msg, "data:")
			}

			// 处理消息
			msg = strings.ReplaceAll(msg, "\\n", "\n")
			function(msg)
		}
	}()

	<-exitChannel
}

/**
 * @description: 获取消息发送者ID
 * @return {string}
 */
func GetSenderID() string {
	return os.Args[2]
}

/**
 * @description: 推送消息
 * @param {string} imtType 包括：qq/qb/wx/wb/tg/tb/wxmp/wxsv
 * @param {string} groupCode 群号
 * @param {string} userID 用户ID
 * @param {string} title 标题
 * @param {string} content 内容
 * @return {*}
 */
func Push(imType, groupCode, userID, title, content string) error {
	params := map[string]interface{}{
		"imType":    imType,
		"groupCode": groupCode,
		"userID":    userID,
		"title":     title,
		"content":   content,
	}
	body, _ := json.Marshal(params)
	_, err := httplib.Post(sockUrl()+"/push").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes()
	if err != nil {
		return err
	} else {
		return nil
	}
}

/**
 * @description: 获取autMan名字
 */
func Name() string {
	resp, _ := httplib.Post(sockUrl()+"/name").Header("Content-Type", "application/json").Bytes()
	name, _ := jsonparser.GetString(resp, "data")
	return name
}

/**
 * @description: 获取autMan机器码
 */
func MachineId() string {
	resp, _ := httplib.Post(sockUrl()+"/machineId").Header("Content-Type", "application/json").Bytes()
	rlt, _ := jsonparser.GetString(resp, "data")
	return rlt
}

/**
 * @description: 获取autMan版本，结果是json字符串{"sn":"1.9.8","content":["版本更新内容1","版本更新内容2"]}
 */
func Version() string {
	resp, _ := httplib.Post(sockUrl()+"/version").Header("Content-Type", "application/json").Bytes()
	rlt, _ := jsonparser.GetString(resp, "data")
	return rlt
}

/**
 * @description: 获取用户otto数据库key-value的value值
 * @param {string} key
 */
func Get(key string, defaultValue ...string) string {
	params := map[string]interface{}{
		"key": key,
	}
	body, _ := json.Marshal(params)
	resp, _ := httplib.Post(sockUrl()+"/get").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes()
	rlt, _ := jsonparser.GetString(resp, "data")
	if rlt == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return rlt
}

/**
 * @description: 设置用户otto数据库key-value的value值
 * @param {string} key
 * @param {string} value
 */
func Set(key, value string) error {
	params := map[string]interface{}{
		"key":   key,
		"value": value,
	}
	body, _ := json.Marshal(params)
	if _, err := httplib.Post(sockUrl()+"/set").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes(); err != nil {
		return err
	} else {
		return nil
	}
}

/**
 * @description: 删除用户otto数据库key-value的value值
 * @param {string} key
 */
func Delete(key string) error {
	params := map[string]interface{}{
		"key": key,
	}
	body, _ := json.Marshal(params)
	if _, err := httplib.Post(sockUrl()+"/delete").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes(); err != nil {
		return err
	} else {
		return nil
	}
}

/**
 * @description: 获取数据库key-value的value值
 * @param {string} bucket
 * @param {string} key
 */
func BucketGet(bucket, key string) string {
	params := map[string]interface{}{
		"bucket": bucket,
		"key":    key,
	}
	body, _ := json.Marshal(params)
	resp, _ := httplib.Post(sockUrl()+"/bucketGet").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes()
	rlt, _ := jsonparser.GetString(resp, "data")
	return rlt
}

/**
 * @description: 设置数据库key-value的value值
 * @param {string} bucket
 * @param {string} key
 * @param {string} value
 */
func BucketSet(bucket, key, value string) error {
	params := map[string]interface{}{
		"bucket": bucket,
		"key":    key,
		"value":  value,
	}
	body, _ := json.Marshal(params)
	if _, err := httplib.Post(sockUrl()+"/bucketSet").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes(); err != nil {
		return err
	} else {
		return nil
	}
}

/**
 * @description: 删除数据库key-value的value值
 * @param {string} bucket
 * @param {string} key
 */
func BucketDelete(bucket, key string) error {
	params := map[string]interface{}{
		"bucket": bucket,
		"key":    key,
	}
	body, _ := json.Marshal(params)
	if _, err := httplib.Post(sockUrl()+"/bucketDel").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes(); err != nil {
		return err
	} else {
		return nil
	}
}

/**
 * @description: 获取指定数据库的所有为value的keys
 * @param {string} bucket
 * @param {string} value
 */
func BucketKeys(bucket, value string) []string {
	params := map[string]interface{}{
		"bucket": bucket,
		"value":  value,
	}
	body, _ := json.Marshal(params)
	resp, _ := httplib.Post(sockUrl()+"/bucketKeys").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes()
	data, _ := jsonparser.GetUnsafeString(resp, "data")
	rlt := []string{}
	json.Unmarshal([]byte(data), &rlt)
	return rlt
}

/**
 * @description: 获取指定数据桶所有的key集合
 * @param {string} bucket
 */
func BucketAllKeys(bucket string) []string {
	params := map[string]interface{}{
		"bucket": bucket,
	}
	body, _ := json.Marshal(params)
	resp, _ := httplib.Post(sockUrl()+"/bucketAllKeys").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes()
	data, _ := jsonparser.GetUnsafeString(resp, "data")
	rlt := []string{}
	json.Unmarshal([]byte(data), &rlt)
	return rlt
}

/**
 * @description: 通知管理员
 * @param {string} content
 * @param {string} imtypes
 */
func NotifyMasters(content string, imtypes []string) error {
	params := map[string]interface{}{
		"content": content,
		"imtypes": imtypes,
	}
	body, _ := json.Marshal(params)
	if _, err := httplib.Post(sockUrl()+"/notifyMasters").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes(); err != nil {
		return err
	} else {
		return nil
	}
}

/**
 * @description: 当前系统授权的激活状态
 */
func Coffee() bool {
	resp, _ := httplib.Post(sockUrl()+"/coffee").Header("Content-Type", "application/json").Bytes()
	rlt, _ := jsonparser.GetBoolean(resp, "data")
	return rlt
}

/**
 * @description: 京东、淘宝、拼多多的转链推广
 * @param {string} msg
 * @return {string} 转链后的信息
 */
func Promotion(msg string) string {
	params := map[string]interface{}{
		"msg": msg,
	}
	body, _ := json.Marshal(params)
	resp, _ := httplib.Post(sockUrl()+"/spread").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes()
	rlt, _ := jsonparser.GetString(resp, "data")
	return rlt
}

type Sender struct {
	SenderID string
}

/**
 * @description: 获取数据库key-value的value值
 * @param {string} bucket
 * @param {string} key
 */
func (s *Sender) BucketGet(bucket, key string) string {
	params := map[string]interface{}{
		"senderid": s.SenderID,
		"bucket":   bucket,
		"key":      key,
	}
	body, _ := json.Marshal(params)
	resp, _ := httplib.Post(sockUrl()+"/bucketGet").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes()
	rlt, _ := jsonparser.GetString(resp, "data")
	return rlt
}

/**
 * @description: 设置数据库key-value的value值
 * @param {string} bucket
 * @param {string} key
 * @param {string} value
 */
func (s *Sender) BucketSet(bucket, key, value string) error {
	params := map[string]interface{}{
		"senderid": s.SenderID,
		"bucket":   bucket,
		"key":      key,
		"value":    value,
	}
	body, _ := json.Marshal(params)
	if _, err := httplib.Post(sockUrl()+"/bucketSet").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes(); err != nil {
		return err
	} else {
		return nil
	}
}

/**
 * @description: 删除数据库key-value的value值
 * @param {string} bucket
 * @param {string} key
 */
func (s *Sender) BucketDelete(bucket, key string) error {
	params := map[string]interface{}{
		"senderid": s.SenderID,
		"bucket":   bucket,
		"key":      key,
	}
	body, _ := json.Marshal(params)
	if _, err := httplib.Post(sockUrl()+"/bucketDel").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes(); err != nil {
		return err
	} else {
		return nil
	}
}

/**
 * @description: 获取指定数据库的所有为value的keys
 * @param {string} bucket
 * @param {string} value
 */
func (s *Sender) BucketKeys(bucket, value string) []string {
	params := map[string]interface{}{
		"senderid": s.SenderID,
		"bucket":   bucket,
		"value":    value,
	}
	body, _ := json.Marshal(params)
	resp, _ := httplib.Post(sockUrl()+"/bucketKeys").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes()
	data, _ := jsonparser.GetUnsafeString(resp, "data")
	rlt := []string{}
	json.Unmarshal([]byte(data), &rlt)
	return rlt
}

/**
 * @description: 获取指定数据桶所有的key集合
 * @param {string} bucket
 */
func (s *Sender) BucketAllKeys(bucket string) []string {
	params := map[string]interface{}{
		"senderid": s.SenderID,
		"bucket":   bucket,
	}
	body, _ := json.Marshal(params)
	resp, _ := httplib.Post(sockUrl()+"/bucketAllKeys").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes()
	data, _ := jsonparser.GetUnsafeString(resp, "data")
	rlt := []string{}
	json.Unmarshal([]byte(data), &rlt)
	return rlt
}

func (s *Sender) SetContinue() bool {
	params := map[string]interface{}{
		"senderid": s.SenderID,
	}
	body, _ := json.Marshal(params)
	resp, _ := httplib.Post(sockUrl()+"/continue").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes()
	rlt, _ := jsonparser.GetBoolean(resp, "data")
	return rlt
}

func (s *Sender) GetImtype() string {
	params := map[string]interface{}{
		"senderid": s.SenderID,
	}
	body, _ := json.Marshal(params)
	resp, _ := httplib.Post(sockUrl()+"/getImtype").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes()
	rlt, _ := jsonparser.GetString(resp, "data")
	return rlt
}

func (s *Sender) GetUserID() string {
	params := map[string]interface{}{
		"senderid": s.SenderID,
	}
	body, _ := json.Marshal(params)
	resp, _ := httplib.Post(sockUrl()+"/getUserID").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes()
	rlt, _ := jsonparser.GetString(resp, "data")
	return rlt
}

func (s *Sender) GetUsername() string {
	params := map[string]interface{}{
		"senderid": s.SenderID,
	}
	body, _ := json.Marshal(params)
	resp, _ := httplib.Post(sockUrl()+"/getUserName").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes()
	rlt, _ := jsonparser.GetString(resp, "data")
	return rlt
}

func (s *Sender) GetUserAvatarUrl() string {
	params := map[string]interface{}{
		"senderid": s.SenderID,
	}
	body, _ := json.Marshal(params)
	resp, _ := httplib.Post(sockUrl()+"/getUserAvatarUrl").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes()
	rlt, _ := jsonparser.GetString(resp, "data")
	return rlt
}

func (s *Sender) GetChatID() string {
	params := map[string]interface{}{
		"senderid": s.SenderID,
	}
	body, _ := json.Marshal(params)
	resp, _ := httplib.Post(sockUrl()+"/getChatID").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes()
	rlt, _ := jsonparser.GetString(resp, "data")
	return rlt
}

func (s *Sender) GetChatName() string {
	params := map[string]interface{}{
		"senderid": s.SenderID,
	}
	body, _ := json.Marshal(params)
	resp, _ := httplib.Post(sockUrl()+"/getChatName").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes()
	rlt, _ := jsonparser.GetString(resp, "data")
	return rlt
}

func (s *Sender) IsAdmin() bool {
	params := map[string]interface{}{
		"senderid": s.SenderID,
	}
	body, _ := json.Marshal(params)
	resp, _ := httplib.Post(sockUrl()+"/isAdmin").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes()
	rlt, _ := jsonparser.GetBoolean(resp, "data")
	return rlt
}

func (s *Sender) GetMessage() string {
	params := map[string]interface{}{
		"senderid": s.SenderID,
	}
	body, _ := json.Marshal(params)
	resp, _ := httplib.Post(sockUrl()+"/getMessage").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes()
	rlt, _ := jsonparser.GetString(resp, "data")
	return rlt
}

/*
* @description: 获取消息ID
* @return {string} 消息ID
 */
func (s *Sender) GetMessageID() string {
	params := map[string]interface{}{
		"senderid": s.SenderID,
	}
	body, _ := json.Marshal(params)
	resp, _ := httplib.Post(sockUrl()+"/getMessageID").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes()
	rlt, _ := jsonparser.GetString(resp, "data")
	return rlt
}

/*
* @description: 撤回用户消息
* @param {string} messageid 消息ID
 */
func (s *Sender) RecallMessage(messageid string) error {
	params := map[string]interface{}{
		"senderid":  s.SenderID,
		"messageid": messageid,
	}
	body, _ := json.Marshal(params)
	if _, err := httplib.Post(sockUrl()+"/recallMessage").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes(); err != nil {
		return err
	} else {
		return nil
	}
}

/*
* @description: 即，模拟当前用户的身份，修改用户输入的内容，将新内容注入到消息队列中，多用于通过关键词拉起其他插件或任务
* @param {string} content 消息内容
 */
func (s *Sender) BreakIn(content string) error {
	params := map[string]interface{}{
		"senderid": s.SenderID,
		"text":     content,
	}
	body, _ := json.Marshal(params)
	if _, err := httplib.Post(sockUrl()+"/breakIn").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes(); err != nil {
		return err
	} else {
		return nil
	}
}

/*
* @description: 获取用户触发的关键词，对应头注中rule规则中的小括号或问号
* @param {int} index 参数索引
* @return {string} 参数值
 */
func (s *Sender) Param(index int) string {
	params := map[string]interface{}{
		"senderid": s.SenderID,
		"index":    index,
	}
	body, _ := json.Marshal(params)
	resp, _ := httplib.Post(sockUrl()+"/param").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes()
	rlt, _ := jsonparser.GetString(resp, "data")
	return rlt
}

/*
* @description: 回复文本
* @param {string} text 文本内容，文本中可以使用CQ码，例如：[CQ:at,qq=123456]，[CQ:image,file=xxx.jpg]
 */
func (s *Sender) Reply(text string) ([]string, error) {
	params := map[string]interface{}{
		"senderid": s.SenderID,
		"text":     text,
	}
	body, _ := json.Marshal(params)
	var msgIds []string
	if resp, err := httplib.Post(sockUrl()+"/sendText").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes(); err == nil {
		if data, err := jsonparser.GetUnsafeString(resp, "data"); err == nil {
			json.Unmarshal([]byte(data), &msgIds)
			return msgIds, nil
		}
	}
	return nil, errors.New("回复失败")
}

/*
* @description: 回复markdown
* @param {string} text markdown字符串
 */
func (s *Sender) ReplyMarkdown(text string) ([]string, error) {
	params := map[string]interface{}{
		"senderid": s.SenderID,
		"markdown": text,
	}
	body, _ := json.Marshal(params)
	var msgIds []string
	if resp, err := httplib.Post(sockUrl()+"/sendMarkdown").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes(); err == nil {
		if data, err := jsonparser.GetUnsafeString(resp, "data"); err == nil {
			json.Unmarshal([]byte(data), &msgIds)
			return msgIds, nil
		}
	}
	return nil, errors.New("回复失败")
}

/*
* @description: 回复图片
* @param {string} imageurl 图片链接
* @return {[]string} 消息ID
 */
func (s *Sender) ReplyImage(imageurl string) ([]string, error) {
	params := map[string]interface{}{
		"senderid": s.SenderID,
		"imageurl": imageurl,
	}
	body, _ := json.Marshal(params)
	var msgIds []string
	if resp, err := httplib.Post(sockUrl()+"/sendImage").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes(); err == nil {
		if data, err := jsonparser.GetUnsafeString(resp, "data"); err == nil {
			json.Unmarshal([]byte(data), &msgIds)
			return msgIds, nil
		}
	}
	return nil, errors.New("回复失败")
}

/*
* @description: 回复语音
* @param {string} voiceurl 语音链接
* @return {[]string} 消息ID
 */
func (s *Sender) ReplyVoice(voiceurl string) ([]string, error) {
	params := map[string]interface{}{
		"senderid": s.SenderID,
		"voiceurl": voiceurl,
	}
	body, _ := json.Marshal(params)
	var msgIds []string
	if resp, err := httplib.Post(sockUrl()+"/sendVoice").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes(); err == nil {
		if data, err := jsonparser.GetUnsafeString(resp, "data"); err == nil {
			json.Unmarshal([]byte(data), &msgIds)
			return msgIds, nil
		}
	}
	return nil, errors.New("回复失败")
}

/*
* @description: 回复视频
* @param {string} videourl 视频链接
* @return {[]string} 消息ID
 */
func (s *Sender) ReplyVideo(videourl string) ([]string, error) {
	params := map[string]interface{}{
		"senderid": s.SenderID,
		"videourl": videourl,
	}
	body, _ := json.Marshal(params)
	var msgIds []string
	if resp, err := httplib.Post(sockUrl()+"/sendVideo").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes(); err == nil {
		if data, err := jsonparser.GetUnsafeString(resp, "data"); err == nil {
			json.Unmarshal([]byte(data), &msgIds)
			return msgIds, nil
		}
		return msgIds, nil
	}
	return nil, errors.New("回复失败")
}

/*
* @description: 等待用户输入
* @param {string} timeout 超时，单位：毫秒
* @return {string} 用户输入的消息
 */
func (s *Sender) Listen(timeout int) string {
	params := map[string]interface{}{
		"senderid": s.SenderID,
		"timeout":  timeout,
	}
	body, _ := json.Marshal(params)
	resp, _ := httplib.Post(sockUrl()+"/listen").Header("Content-Type", "application/json").Body(body).SetTransport(transport).SetTimeout(time.Millisecond*time.Duration(timeout), time.Millisecond*time.Duration(timeout)).Bytes()
	rlt, _ := jsonparser.GetString(resp, "data")
	return rlt
}

/*
* @description: 等待用户支付
* @param {string} timeout 超时，单位：毫秒
* @return {string} 用户支付信息json字符串
 */
func (s *Sender) WaitPay(exitCode string, timeout int) string {
	params := map[string]interface{}{
		"senderid": s.SenderID,
		"exitCode": exitCode,
		"timeout":  timeout,
	}
	body, _ := json.Marshal(params)
	resp, _ := httplib.Post(sockUrl()+"/waitPay").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes()
	rlt, _ := jsonparser.GetString(resp, "data")
	return rlt
}

/*
* @description: 判断当前是否处于等待用户支付状态
 */
func (s *Sender) AtWaitPay() bool {
	params := map[string]interface{}{
		"senderid": s.SenderID,
	}
	body, _ := json.Marshal(params)
	resp, _ := httplib.Post(sockUrl()+"/atWaitPay").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes()
	rlt, _ := jsonparser.GetBoolean(resp, "data")
	return rlt
}

func (s *Sender) GroupInviteIn(friend, group string) error {
	params := map[string]interface{}{
		"senderid": s.SenderID,
		"friend":   friend,
		"group":    group,
	}
	body, _ := json.Marshal(params)
	if _, err := httplib.Post(sockUrl()+"/groupInviteIn").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes(); err != nil {
		return err
	} else {
		return nil
	}
}

func (s *Sender) GroupKick(userid string) error {
	params := map[string]interface{}{
		"senderid": s.SenderID,
		"userid":   userid,
	}
	body, _ := json.Marshal(params)
	if _, err := httplib.Post(sockUrl()+"/groupKick").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes(); err != nil {
		return err
	} else {
		return nil
	}
}

func (s *Sender) GroupBan(userid string, timeout int) error {
	params := map[string]interface{}{
		"senderid": s.SenderID,
		"userid":   userid,
		"timeout":  timeout,
	}
	body, _ := json.Marshal(params)
	if _, err := httplib.Post(sockUrl()+"/groupBan").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes(); err != nil {
		return err
	} else {
		return nil
	}
}

func (s *Sender) GroupUnban(userid string) error {
	params := map[string]interface{}{
		"senderid": s.SenderID,
		"userid":   userid,
	}
	body, _ := json.Marshal(params)
	if _, err := httplib.Post(sockUrl()+"/groupUnban").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes(); err != nil {
		return err
	} else {
		return nil
	}
}

func (s *Sender) GroupWholeBan(userid string) error {
	params := map[string]interface{}{
		"senderid": s.SenderID,
		"userid":   userid,
	}
	body, _ := json.Marshal(params)
	if _, err := httplib.Post(sockUrl()+"/groupWholeBan").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes(); err != nil {
		return err
	} else {
		return nil
	}
}

func (s *Sender) GroupWholeUnban(userid string) error {
	params := map[string]interface{}{
		"senderid": s.SenderID,
		"userid":   userid,
	}
	body, _ := json.Marshal(params)
	if _, err := httplib.Post(sockUrl()+"/groupWholeUnban").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes(); err != nil {
		return err
	} else {
		return nil
	}
}

func (s *Sender) GroupNoticeSend(notice string) error {
	params := map[string]interface{}{
		"senderid": s.SenderID,
		"notice":   notice,
	}
	body, _ := json.Marshal(params)
	if _, err := httplib.Post(sockUrl()+"/groupNoticeSend").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes(); err != nil {
		return err
	} else {
		return nil
	}
}

func (s *Sender) GetPluginName() string {
	params := map[string]interface{}{
		"senderid": s.SenderID,
	}
	body, _ := json.Marshal(params)
	resp, _ := httplib.Post(sockUrl()+"/getPluginName").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes()
	rlt, _ := jsonparser.GetString(resp, "data")
	return rlt
}

func (s *Sender) GetPluginVersion() string {
	params := map[string]interface{}{
		"senderid": s.SenderID,
	}
	body, _ := json.Marshal(params)
	resp, _ := httplib.Post(sockUrl()+"/getPluginVersion").Header("Content-Type", "application/json").Body(body).SetTransport(transport).Bytes()
	rlt, _ := jsonparser.GetString(resp, "data")
	return rlt
}
