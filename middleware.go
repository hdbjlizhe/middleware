package middleware

import (
	"errors"
	"net/url"
	"os"
	"strings"

	"github.com/beego/beego/v2/client/httplib"
)

var Port string

func localUrl() string {
	return "http://localhost:" + Port + "/otto"
}

/**
 * @description: 设置端口号
 */
func SetPort() {
	Port = os.Args[1]
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
	params := url.Values{
		"imType":   {imType},
		"groupCode": {groupCode},
		"userID":    {userID},
		"title":     {title},
		"content":   {content},
	}
	if resp, err := httplib.Get(localUrl() + "/push?" + params.Encode()).String(); err != nil {
		return err
	} else if resp != "ok" {
		return errors.New("推送失败")
	} else {
		return nil
	}
}

/**
 * @description: 获取autMan名字
 */
func Name() string {
	resp, _ := httplib.Get(localUrl() + "/name").String()
	return resp
}

/**
 * @description: 获取autMan机器码
 */
func MachineId() string {
	resp, _ := httplib.Get(localUrl() + "/machineId").String()
	return resp
}

/**
 * @description: 获取autMan版本，结果是json字符串{"sn":"1.9.8","content":["版本更新内容1","版本更新内容2"]}
 */
func Version() string {
	resp, _ := httplib.Get(localUrl() + "/version").String()
	return resp
}

/**
 * @description: 获取用户otto数据库key-value的value值
 * @param {string} key
 */
func Get(key string) string {
	resp, _ := httplib.Get(localUrl() + "/get?key=" + url.QueryEscape(key)).String()
	return resp
}

/**
 * @description: 设置用户otto数据库key-value的value值
 * @param {string} key
 * @param {string} value
 */
func Set(key, value string) error {
	params := url.Values{
		"key":   {key},
		"value": {value},
	}
	if resp, err := httplib.Get(localUrl() + "/set?" + params.Encode()).String(); err != nil {
		return err
	} else if resp != "ok" {
		return errors.New("设置失败")
	} else {
		return nil
	}
}

/**
 * @description: 删除用户otto数据库key-value的value值
 * @param {string} key
 */
func Delete(key string) error {
	if resp, err := httplib.Get(localUrl() + "/delete?key=" + url.QueryEscape(key)).String(); err != nil {
		return err
	} else if resp != "ok" {
		return errors.New("删除失败")
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
	params := url.Values{
		"bucket": {bucket},
		"key":    {key},
	}
	resp, _ := httplib.Get(localUrl() + "/bucketGet?" + params.Encode()).String()
	return resp
}

/**
 * @description: 设置数据库key-value的value值
 * @param {string} bucket
 * @param {string} key
 * @param {string} value
 */
func BucketSet(bucket, key, value string) error {
	params := url.Values{
		"bucket": {bucket},
		"key":    {key},
		"value":  {value},
	}
	if resp, err := httplib.Get(localUrl() + "/bucketSet?" + params.Encode()).String(); err != nil {
		return err
	} else if resp != "ok" {
		return errors.New("设置失败")
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
	params := url.Values{
		"bucket": {bucket},
		"key":    {key},
	}
	if resp, err := httplib.Get(localUrl() + "/bucketDel?" + params.Encode()).String(); err != nil {
		return err
	} else if resp != "ok" {
		return errors.New("删除失败")
	} else {
		return nil
	}
}

/**
 * @description: 获取指定数据库的所有为value的keys
 * @param {string} bucket
 * @param {string} value
 */
func BucketKeys(bucket, value string) string {
	params := url.Values{
		"bucket": {bucket},
		"value":  {value},
	}
	resp, _ := httplib.Get(localUrl() + "/bucketKeys?" + params.Encode()).String()
	return resp
}

/**
 * @description: 获取指定数据桶所有的key集合
 * @param {string} bucket
 */
func BucketAllKeys(bucket string) []string {
	resp, _ := httplib.Get(localUrl() + "/bucketAllKeys?bucket=" + url.QueryEscape(bucket)).String()
	return strings.Split(resp, ",")
}

/**
 * @description: 通知管理员
 * @param {string} content
 * @param {string} imtypes
 */
func NotifyMasters(content string, imtypes []string) error {
	params := url.Values{
		"content": {content},
		"imtypes": imtypes,
	}
	if resp, err := httplib.Get(localUrl() + "/notifyMasters?" + params.Encode()).String(); err != nil {
		return err
	} else if resp != "ok" {
		return errors.New("通知失败")
	} else {
		return nil
	}
}

/**
 * @description: 当前系统授权的激活状态
 */
func Coffee() bool {
	resp, _ := httplib.Get(localUrl() + "/coffee").String()
	return resp == "true"
}

/**
 * @description: 京东、淘宝、拼多多的转链推广
 * @param {string} msg
 * @return {string} 转链后的信息
 */
func Promotion(msg string) string {
	resp, _ := httplib.Get(localUrl() + "/spread?msg=" + url.QueryEscape(msg)).String()
	return resp
}

type Sender struct {
	SenderID string
}

func (s *Sender) GetImtype() string {
	resp, _ := httplib.Get(localUrl() + "/getImtype?senderid=" + s.SenderID).String()
	return resp
}

func (s *Sender) GetUserID() string {
	resp, _ := httplib.Get(localUrl() + "/getUserID?senderid=" + s.SenderID).String()
	return strings.Trim(resp, "\"")
}

func (s *Sender) GetUsername() string {
	resp, _ := httplib.Get(localUrl() + "/getUsername?senderid=" + s.SenderID).String()
	return resp
}

func (s *Sender) GetUserAvatarUrl() string {
	resp, _ := httplib.Get(localUrl() + "/getUserAvatarUrl?senderid=" + s.SenderID).String()
	return resp
}

func (s *Sender) GetChatID() string {
	resp, _ := httplib.Get(localUrl() + "/getChatID?senderid=" + s.SenderID).String()
	return resp
}

func (s *Sender) GetChatName() string {
	resp, _ := httplib.Get(localUrl() + "/getChatName?senderid=" + s.SenderID).String()
	return resp
}

func (s *Sender) IsAdmin() bool {
	resp, _ := httplib.Get(localUrl() + "/isAdmin?senderid=" + s.SenderID).String()
	return resp == "true"
}

func (s *Sender) GetMessage() string {
	resp, _ := httplib.Get(localUrl() + "/getMessage?senderid=" + s.SenderID).String()
	return resp
}

func (s *Sender) GetMessageID() string {
	resp, _ := httplib.Get(localUrl() + "/getMessageID?senderid=" + s.SenderID).String()
	return resp
}

func (s *Sender) RecallMessage(messageid string) error {
	if resp, err := httplib.Get(localUrl() + "/recallMessage?senderid=" + s.SenderID + "&messageid=" + messageid).String(); err != nil {
		return err
	} else if resp != "ok" {
		return errors.New("撤回失败")
	} else {
		return nil
	}
}

func (s *Sender) BreakIn(content string) error {
	if resp, err := httplib.Get(localUrl() + "/breakIn?senderid=" + s.SenderID + "&text=" + url.QueryEscape(content)).String(); err != nil {
		return err
	} else if resp != "ok" {
		return errors.New("注入新消息失败")
	} else {
		return nil
	}
}

func (s *Sender) Param(index int) string {
	resp, _ := httplib.Get(localUrl() + "/param?senderid=" + s.SenderID + "&index=" + string(index)).String()
	return resp
}

func (s *Sender) Reply(text string) error {
	if resp, err := httplib.Get(localUrl() + "/sendText?senderid=" + s.SenderID + "&text=" + url.QueryEscape(text)).String(); err != nil {
		return err
	} else if resp != "ok" {
		return errors.New("回复失败")
	} else {
		return nil
	}
}

func (s *Sender) ReplyImage(imageurl string) error {
	if resp, err := httplib.Get(localUrl() + "/sendImage?senderid=" + s.SenderID + "&imageurl=" + url.QueryEscape(imageurl)).String(); err != nil {
		return err
	} else if resp != "ok" {
		return errors.New("回复图片失败")
	} else {
		return nil
	}
}

func (s *Sender) ReplyVoice(voiceurl string) error {
	if resp, err := httplib.Get(localUrl() + "/sendVoice?senderid=" + s.SenderID + "&voiceurl=" + url.QueryEscape(voiceurl)).String(); err != nil {
		return err
	} else if resp != "ok" {
		return errors.New("回复语音失败")
	} else {
		return nil
	}
}

func (s *Sender) ReplyVideo(videourl string) error {
	if resp, err := httplib.Get(localUrl() + "/sendVideo?senderid=" + s.SenderID + "&videourl=" + url.QueryEscape(videourl)).String(); err != nil {
		return err
	} else if resp != "ok" {
		return errors.New("回复视频失败")
	} else {
		return nil
	}
}

func (s *Sender) Listen(timeout int) string {
	resp, _ := httplib.Get(localUrl() + "/listen?senderid=" + s.SenderID + "&timeout=" + string(timeout)).String()
	return resp
}

func (s *Sender) WaitPay(eixtCode string, timeout int) string {
	resp, _ := httplib.Get(localUrl() + "/waitPay?senderid=" + s.SenderID + "&eixtCode=" + eixtCode + "&timeout=" + string(timeout)).String()
	return resp
}

func (s *Sender) AtWaitPay() bool {
	resp, _ := httplib.Get(localUrl() + "/atWaitPay?senderid=" + s.SenderID).String()
	return resp == "true"
}

func (s *Sender) GroupInviteIn(friend, group string) error {
	if resp, err := httplib.Get(localUrl() + "/groupInviteIn?senderid=" + s.SenderID + "&friend=" + friend + "&group=" + group).String(); err != nil {
		return err
	} else if resp != "ok" {
		return errors.New("邀请失败")
	} else {
		return nil
	}
}

func (s *Sender) GroupKick(userid string) error {
	if resp, err := httplib.Get(localUrl() + "/groupKick?senderid=" + s.SenderID + "&userid=" + userid).String(); err != nil {
		return err
	} else if resp != "ok" {
		return errors.New("踢出失败")
	} else {
		return nil
	}
}

func (s *Sender) GroupBan(userid string, timeout int) error {
	if resp, err := httplib.Get(localUrl() + "/groupBan?senderid=" + s.SenderID + "&userid=" + userid + "&timeout=" + string(timeout)).String(); err != nil {
		return err
	} else if resp != "ok" {
		return errors.New("禁言失败")
	} else {
		return nil
	}
}

func (s *Sender) GroupUnban(userid string) error {
	if resp, err := httplib.Get(localUrl() + "/groupUnban?senderid=" + s.SenderID + "&userid=" + userid).String(); err != nil {
		return err
	} else if resp != "ok" {
		return errors.New("解除禁言失败")
	} else {
		return nil
	}
}

func (s *Sender) GroupWholeBan(userid string) error {
	if resp, err := httplib.Get(localUrl() + "/groupWholeBan?senderid=" + s.SenderID + "&userid=" + userid).String(); err != nil {
		return err
	} else if resp != "ok" {
		return errors.New("全员禁言失败")
	} else {
		return nil
	}
}

func (s *Sender) GroupWholeUnban(userid string) error {
	if resp, err := httplib.Get(localUrl() + "/groupWholeUnban?senderid=" + s.SenderID + "&userid=" + userid).String(); err != nil {
		return err
	} else if resp != "ok" {
		return errors.New("解除全员禁言失败")
	} else {
		return nil
	}
}

func (s *Sender) GroupNoticeSend(notice string) error {
	if resp, err := httplib.Get(localUrl() + "/groupNoticeSend?senderid=" + s.SenderID + "&notice=" + url.QueryEscape(notice)).String(); err != nil {
		return err
	} else if resp != "ok" {
		return errors.New("发送失败")
	} else {
		return nil
	}
}

func (s *Sender) GetPluginName() string {
	resp, _ := httplib.Get(localUrl() + "/getPluginName?senderid=" + s.SenderID).String()
	return resp
}

func (s *Sender) GetPluginVersion() string {
	resp, _ := httplib.Get(localUrl() + "/getPluginVersion?senderid=" + s.SenderID).String()
	return resp
}
