package models

import (
	"time"

	"github.com/TimeForCoin/Server/app/libs"
	"github.com/go-redis/redis"
	jsoniter "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 快速缓存

// - 热门任务列表
// - 任务浏览量(热度)
// - 用户基本信息(头像、昵称)
// - 点赞用户记录

// CacheModel 缓存数据库
type CacheModel struct {
	Redis *redis.Client
}

// DataKind 数据类型
type DataKind string

// DataKind 缓存数据类型
const (
	KindOfLikeTask    DataKind = "like-task-"
	KindOfLikeComment DataKind = "like-comment-"
	KindOfCollectTask DataKind = "collect-task-"
	KindOfBaseInfo    DataKind = "info-"
)

// WillUpdate 更新缓存数据
func (c *CacheModel) WillUpdate(userID primitive.ObjectID, kind DataKind) error {
	return c.Redis.Del(string(kind) + userID.Hex()).Err()
}

// IsLikeTask 用户是否点赞任务
func (c *CacheModel) IsLikeTask(userID, taskID primitive.ObjectID) bool {
	setName := string(KindOfLikeTask) + userID.Hex()
	exist, err := c.Redis.Exists(setName).Result()
	if err != nil {
		return false
	}
	// 不存在记录
	if exist == 0 {
		// 从数据库读取
		set := GetModel().Set.GetSets(userID, SetOfLikeTask)
		if len(set.LikeTaskID) > 0 {
			var setID []string
			for _, id := range set.LikeTaskID {
				setID = append(setID, id.Hex())
			}
			err = c.Redis.SAdd(setName, setID).Err()
			if err != nil {
				return false
			}
		}
	}
	val, err := c.Redis.SIsMember(setName, taskID.Hex()).Result()
	return val
}

// IsLikeComment 用户是否点赞评论
func (c *CacheModel) IsLikeComment(userID, commentID primitive.ObjectID) bool {
	setName := string(KindOfLikeComment) + userID.Hex()
	exist, err := c.Redis.Exists(setName).Result()
	if err != nil {
		return false
	}
	// 不存在记录
	if exist == 0 {
		// 从数据库读取
		set := GetModel().Set.GetSets(userID, SetOfLikeComment)
		if len(set.LikeCommentID) > 0 {
			var setID []string
			for _, id := range set.LikeCommentID {
				setID = append(setID, id.Hex())
			}
			err = c.Redis.SAdd(setName, setID).Err()
			if err != nil {
				return false
			}
		}
	}
	val, err := c.Redis.SIsMember(setName, commentID.Hex()).Result()
	return val
}

// IsLikeComment 用户是否收藏任务
func (c *CacheModel) IsCollectTask(userID, taskID primitive.ObjectID) bool {
	setName := string(KindOfCollectTask) + userID.Hex()
	exist, err := c.Redis.Exists(setName).Result()
	if err != nil {
		return false
	}
	// 不存在记录
	if exist == 0 {
		// 从数据库读取
		set := GetModel().Set.GetSets(userID, SetOfCollectTask)
		if len(set.CollectTaskID) > 0 {
			var setID []string
			for _, id := range set.CollectTaskID {
				setID = append(setID, id.Hex())
			}
			err = c.Redis.SAdd(setName, setID).Err()
			if err != nil {
				return false
			}
		}
	}
	val, err := c.Redis.SIsMember(setName, taskID.Hex()).Result()
	return val
}

// UserBaseInfo 用户基本信息数据
type UserBaseInfo struct {
	ID       string `json:"id"`
	Nickname string
	Avatar   string
	Gender   UserGender
	Type     UserType
}

// GetUserBaseInfo 获取用户基本信息
func (c *CacheModel) GetUserBaseInfo(id primitive.ObjectID) (UserBaseInfo, error) {
	baseInfo := UserBaseInfo{}
	val, err := c.Redis.Get("info-" + id.Hex()).Result()
	// 不存在记录
	if err != nil {
		// 从数据库读取
		user, err := GetModel().User.GetUserByID(id)
		if err != nil {
			return baseInfo, err
		}
		baseInfo.ID = user.ID.Hex()
		baseInfo.Nickname = user.Info.Nickname
		baseInfo.Avatar = user.Info.Avatar
		baseInfo.Gender = user.Info.Gender
		baseInfo.Type = user.Data.Type
		str, err := jsoniter.Marshal(baseInfo)
		if err != nil {
			return baseInfo, err
		}
		return baseInfo, c.Redis.Set("info-"+id.Hex(), str, time.Hour*24).Err()
	}
	err = jsoniter.Unmarshal([]byte(val), &baseInfo)
	return baseInfo, err
}

// SetCertification 设置认证
func (c *CacheModel) SetCertification(userID primitive.ObjectID, code string) error {
	return c.Redis.Set("certification-"+userID.Hex(), code, time.Minute*30).Err()
}

// CheckCertification 检查认证
func (c *CacheModel) CheckCertification(userID primitive.ObjectID, email, code string, use bool) (exist bool, right bool) {
	token, err := c.Redis.Get("certification-" + userID.Hex()).Result()
	if err != nil {
		return false, false
	}
	rightCode := libs.GetHash(token + "&" + email)
	if rightCode != code {
		return true, false
	}
	err = c.Redis.Del("certification-" + userID.Hex()).Err()
	return true, true
}
