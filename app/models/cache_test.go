package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
)

func TestCache(t *testing.T) {
	t.Run("InitRedis", testInitRedis)
	t.Run("InitDB", testInitDB)
	t.Run("GetUserBaseInfo", testGetUserBaseInfo)
	t.Run("testAddLike", testAddLike)
	t.Run("DisconnectRedis", testDisconnectRedis)
	t.Run("DisconnectDB", testDisconnectDB)
}

func testAddLike(t *testing.T) {
	userID := primitive.NewObjectID()
	taskID := primitive.NewObjectID()
	err := redisInst.Cache.WillUpdate(userID, KindOfLikeTask)
	if err != nil {
		t.Error(err)
	}

	err = model.Set.AddToSet(userID, taskID, SetOfLikeTask)
	if err != nil {
		t.Error(err)
	}
	if  redisInst.Cache.IsLikeTask(userID, primitive.NewObjectID()) == true {
		t.Error()
	}
	if redisInst.Cache.IsLikeTask(userID, taskID) == false {
		t.Error()
	}
	taskID = primitive.NewObjectID()
	err = model.Set.AddToSet(userID, taskID, SetOfLikeTask)
	if err != nil {
		t.Error(err)
	}
	if redisInst.Cache.IsLikeTask(userID, taskID) == true {
		t.Error()
	}

	err = redisInst.Cache.WillUpdate(userID, KindOfLikeTask)
	if err != nil {
		t.Error(err)
	}

	if redisInst.Cache.IsLikeTask(userID, taskID) == false {
		t.Error()
	}


}

func testGetUserBaseInfo(t *testing.T) {
	id, err := model.User.AddUserByViolet(primitive.NewObjectID().Hex())
	if err != nil {
		t.Error(nil)
	}
	t.Log(id)

	if err := redisInst.Cache.WillUpdate(id, KindOfBaseInfo); err != nil {
		t.Error(err)
	}

	if err := model.User.SetUserInfoByID(id, UserInfoSchema{
		Nickname: "Show",
		Gender:   GenderWoman,
	}); err != nil {
		t.Error(nil)
	}

	info, err := redisInst.Cache.GetUserBaseInfo(id)
	if err != nil {
		t.Error(err)
	} else {
		t.Log(info)
	}

	info, err = redisInst.Cache.GetUserBaseInfo(primitive.NewObjectID())
	if err != nil {
		t.Log(err)
	} else {
		t.Error(info)
	}

	if err := model.User.SetUserInfoByID(id, UserInfoSchema{
		Nickname: "ShowShow",
		Gender:   GenderWoman,
	}); err != nil {
		t.Error(nil)
	}

	info, err = redisInst.Cache.GetUserBaseInfo(id)
	if err != nil {
		t.Error(err)
	} else {
		t.Log(info)
	}

	if err := redisInst.Cache.WillUpdate(id, KindOfBaseInfo); err != nil {
		t.Error(err)
	}

	info, err = redisInst.Cache.GetUserBaseInfo(id)
	if err != nil {
		t.Error(err)
	} else {
		t.Log(info)
	}

}
