package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
	"time"
)

func TestUserModel(t *testing.T) {
	t.Run("InitDB", testInitDB)

	t.Run("testUser", testUserModelAll)

	t.Run("DisconnectDB", testDisconnectDB)
}

func testUserModelAll(t *testing.T) {
	// 新建用户
	violetID := primitive.NewObjectID().Hex()
	id, err := model.User.AddUserByViolet(violetID)
	if err != nil {
		t.Error(err)
	}
	t.Log(id)

	user, err := model.User.GetUserByID(id)
	if err != nil {
		t.Error(err)
	}
	t.Log(user)

	user, err = model.User.GetUserByViolet(violetID)
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


}
