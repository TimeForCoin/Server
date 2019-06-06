package libs

import "testing"

func TestGetHash(t *testing.T) {
	res := GetHash(GetRandomString(64))
	t.Log(res)
}