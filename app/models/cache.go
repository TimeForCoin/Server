package models

import (
	"github.com/TimeForCoin/Server/app/utils"
	"time"

	"github.com/go-redis/redis"
	jsoniter "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 快速缓存

// - 用户基本信息(头像、昵称)
// - 用户点赞
// - 用户收藏
// - 用户关系

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
	KindOfFollower    DataKind = "follower-"
	KindOfFollowing   DataKind = "following-"
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

// IsCollectTask 用户是否收藏任务
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

// IsFollowerUser 用户是否被某人关注
func (c *CacheModel) IsFollowerUser(userID, otherID primitive.ObjectID) bool {
	setName := string(KindOfFollower) + userID.Hex()
	exist, err := c.Redis.Exists(setName).Result()
	if err != nil {
		return false
	}
	// 不存在记录
	if exist == 0 {
		// 从数据库读取
		set := GetModel().Set.GetSets(userID, SetOfFollowerUser)
		if len(set.FollowerUserID) > 0 {
			var setID []string
			for _, id := range set.FollowerUserID {
				setID = append(setID, id.Hex())
			}
			err = c.Redis.SAdd(setName, setID).Err()
			if err != nil {
				return false
			}
		}
	}
	val, err := c.Redis.SIsMember(setName, otherID.Hex()).Result()
	return val
}

// IsFollowingUser 用户是否已关注某人
func (c *CacheModel) IsFollowingUser(userID, otherID primitive.ObjectID) bool {
	setName := string(KindOfFollowing) + userID.Hex()
	exist, err := c.Redis.Exists(setName).Result()
	if err != nil {
		return false
	}
	// 不存在记录
	if exist == 0 {
		// 从数据库读取
		set := GetModel().Set.GetSets(userID, SetOfFollowingUser)
		if len(set.FollowingUserID) > 0 {
			var setID []string
			for _, id := range set.FollowingUserID {
				setID = append(setID, id.Hex())
			}
			err = c.Redis.SAdd(setName, setID).Err()
			if err != nil {
				return false
			}
		}
	}
	val, err := c.Redis.SIsMember(setName, otherID.Hex()).Result()
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
	rightCode := utils.GetHash(token + "&" + email)
	if rightCode != code {
		return true, false
	}
	err = c.Redis.Del("certification-" + userID.Hex()).Err()
	return true, true
}

// SetSessionUser 设置会话用户ID
func (c *CacheModel) SetSessionUser(session string, userID primitive.ObjectID) error{
	_, err := c.Redis.Set("login-"+session, userID.Hex(), time.Minute * 10).Result()
	if err != nil {
		return err
	}
	return nil
}

// GetSessionUser 获取用户ID
func (c *CacheModel) GetSessionUser(session string) (primitive.ObjectID, error) {
	user, err := c.Redis.Get("login-"+session).Result()
	if err != nil {
		return primitive.NilObjectID, err
	} else if user == "" {
		return primitive.NilObjectID, ErrNotExist
	}
	userID, err := primitive.ObjectIDFromHex(user)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return userID, nil
}