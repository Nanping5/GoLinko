package httpserver

import (
	v1 "GoLinko/api/v1"
	"GoLinko/internal/config"
	"GoLinko/middleware"
	"GoLinko/pkg/zlog"
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var router *gin.Engine

// InitRouter 初始化 HTTP 路由（需要在配置加载后调用）
func InitRouter() {
	router = gin.Default()
	//配置 CORS 中间件，允许跨域请求
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	router.Use(cors.New(corsConfig))

	//配置静态文件路径，提供头像和文件的访问
	router.Static("/static/avatars", config.GetConfig().StaticSrcConfig.StaticAvatarPath)
	router.Static("/static/files", config.GetConfig().StaticSrcConfig.StaticFilePath)

	//注册路由
	v1Group := router.Group("/v1")
	{
		v1Group.GET("/ping", func(c *gin.Context) {
			v1.JsonBack(c, "pong", 0, nil)

		})

		// 用户相关路由
		v1Group.POST("/register", v1.Register)
		v1Group.POST("/login", v1.Login)
		v1Group.POST("/send_email_code", v1.SendEmailCode)
		v1Group.POST("/login_by_code", v1.LoginByCode)
		v1Group.Use(middleware.JWT())

		v1Group.POST("/upload_file", v1.UploadFile)
		v1Group.PUT("/user_info", v1.UpdateUserInfo)
		v1Group.GET("/get_user_info", v1.GetUserInfo)

		// 群组相关路由
		v1Group.POST("/create_group", v1.CreateGroup)
		v1Group.GET("/load_my_groups", v1.LoadMyGroups)
		v1Group.GET("/check_group_add_mode", v1.CheckGroupAddMode)
		v1Group.POST("/enter_group_directly", v1.EnterGroupDirectly)
		v1Group.POST("/leave_group", v1.LeaveGroup)
		v1Group.POST("/dismiss_group", v1.DismissGroup)
		v1Group.GET("/get_group_info", v1.GetGroupInfo)
		v1Group.PUT("/update_group_info", v1.UpdateGroupInfo)
		v1Group.GET("/get_group_members", v1.GetGroupMembers)
		v1Group.DELETE("/remove_group_member", v1.RemoveGroupMember)

		//联系人相关路由
		v1Group.GET("/contact_user_list", v1.GetContactUserList)
		v1Group.GET("/contact_info", v1.GetContactInfo)
		v1Group.GET("/load_my_joined_groups", v1.LoadMyJoinedGroups)
		v1Group.DELETE("/delete_contact", v1.DeleteContact)
		v1Group.POST("/apply_add_contact", v1.ApplyAddContact)
		v1Group.GET("/contact_apply_list", v1.GetContactApplyList)
		v1Group.POST("/accept_contact_apply", v1.AcceptContactApply)
		v1Group.POST("/reject_contact_apply", v1.RejectContactApply)
		v1Group.POST("/black_contact", v1.BlackContact)
		v1Group.POST("/unblack_contact", v1.UnblackContact)

		//会话相关路由
		v1Group.GET("/check_open_session_allowed", v1.CheckOpenSessionAllowed)
		v1Group.POST("/open_session", v1.OpenSession)
		v1Group.GET("/session_list", v1.GetSessionList)
		v1Group.GET("/group_session_list", v1.GetGroupSessionList)
		v1Group.DELETE("/delete_session", v1.DeleteSession)

		//消息相关路由
		v1Group.GET("/message_list", v1.GetMessageList)
		v1Group.GET("/group_message_list", v1.GetGroupMessageList)
		v1Group.POST("/upload_avatar", v1.UploadAvatar)

		// 管理员相关路由（由服务端二次校验管理员权限）
		v1Group.GET("/admin/user_list", v1.GetUserList)
		v1Group.POST("/admin/able_user", v1.AbleUser)
		v1Group.POST("/admin/disable_user", v1.DisableUser)
		v1Group.DELETE("/admin/delete_user", v1.DeleteUser)
		v1Group.POST("/admin/set_admin", v1.SetAdmin)

		// WebSocket 路由
		v1Group.GET("/ws", v1.WsLogin)
	}

	// 前端静态文件服务（SPA）
	// 先注册 /assets 等静态资源，再用 NoRoute 把其他所有请求回退到 index.html
	router.Static("/assets", "./frontend/dist/assets")
	router.StaticFile("/vite.svg", "./frontend/dist/vite.svg")
	router.StaticFile("/favicon.ico", "./frontend/dist/favicon.ico")
	router.NoRoute(func(c *gin.Context) {
		c.File("./frontend/dist/index.html")
	})
	// 根路径直接返回 index.html
	router.GET("/", func(c *gin.Context) {
		c.File("./frontend/dist/index.html")
	})
}
func StartHTTPServer() {
	addr := fmt.Sprintf("%s:%d", config.GetConfig().MainConfig.Host, config.GetConfig().MainConfig.Port)
	err := router.Run(addr)
	if err != nil {
		zlog.GetLogger().Fatal("启动 HTTP 服务器失败", zap.Error(err))
		return
	}
}
