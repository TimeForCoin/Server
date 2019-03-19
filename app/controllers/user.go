package controllers

// UserController 用户控制
type UserController struct{}

// GetPing serves
// Method:   GET
// Resource: http://localhost:port/ping
func (c *UserController) GetPing() string {
	return "pong"
}
