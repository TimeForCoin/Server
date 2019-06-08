package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
	"time"
)

func TestUserModel(t *testing.T) {
	t.Run("InitDB", testInitDB)
	t.Run("InitRedis", testInitRedis)

	t.Run("testUser", testUserModelAll)

	ctx, finish := GetCtx()
	defer finish()
	err := model.User.Collection.Drop(ctx)
	if err != nil {
		t.Error(err)
	}

	t.Run("DisconnectRedis", testDisconnectRedis)
	t.Run("DisconnectDB", testDisconnectDB)
}

func testUserModelAll(t *testing.T) {
	// 新建用户
	violetID := primitive.NewObjectID()

	user, err := model.User.GetUserByID(violetID)
	if err == nil {
		t.Error("has conflict")
	}
	t.Log(user)

	id, err := model.User.AddUserByViolet(violetID.Hex())
	if err != nil {
		t.Error(err)
	}
	t.Log(id)

	user, err = model.User.GetUserByID(id)
	if err != nil {
		t.Error(err)
	}
	t.Log(user)

	user, err = model.User.GetUserByViolet(violetID.Hex())
	if err != nil {
		t.Error(err)
	}
	t.Log(user)
	// 修改信息
	err = model.User.SetUserInfoByID(id, UserInfoSchema{
		Gender:GenderMan,
		Email: "abc@qq.com",
		Phone:"188",
		Avatar:"bbb",
		Nickname:"ccc",
		Bio:"ddd",
		Location:"kkk",
		Birthday: time.Now().Unix(),
	})
	if err != nil {
		t.Error(err)
	}
	user, err = model.User.GetUserByID(id)
	if err != nil {
		t.Error(err)
	}
	t.Log(user)
	err = model.User.SetUserInfoByID(id, UserInfoSchema{
		Gender: GenderWoman,
		Email: "abc@qq.com",
		Phone:"199",
		Location:"location",
		Birthday: time.Now().Unix(),
	})
	if err != nil {
		t.Error(err)
	}
	user, err = model.User.GetUserByID(id)
	if err != nil {
		t.Error(err)
	}
	t.Log(user)

	if err = model.User.UpdateUserDataCount(id, UserDataCount{
		Money: 10,
		Credit: -4,
		FollowerCount: 1,
		FollowingCount: -1,
	}); err != nil {
		t.Error(err)
	}

	if err = model.User.SetUserType(id, UserTypeAdmin); err != nil {
		t.Error(err)
	}

	if err = model.User.SetUserAttend(id); err != nil {
		t.Error(err)
	}

	user, err = model.User.GetUserByID(id)
	if err != nil {
		t.Error(err)
	}
	t.Log(user)

}
