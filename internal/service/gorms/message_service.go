package gormss

import (
	"GoLinko/internal/config"
	"GoLinko/internal/dao"
	"GoLinko/internal/dto/response"
	"GoLinko/internal/model"
	myredis "GoLinko/internal/service/redis"
	constants "GoLinko/pkg/const"
	"GoLinko/pkg/enum/contact/contact_status_enum"
	"GoLinko/pkg/enum/contact/contact_type_enum"
	"GoLinko/pkg/zlog"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type messageService struct {
}

var MessageService = &messageService{}

// GetMessageList 获取消息列表（通过 sessionId）
func (m *messageService) GetMessageList(ctx context.Context, userId string, sessionId string) (string, []response.GetMessageListResponse, int) {
	db := dao.NewDbClient(ctx)

	// 单聊/群聊统一会话鉴权：只有会话参与者或群成员可以读取历史
	allowed, msg, ret := m.canAccessSessionMessages(ctx, userId, sessionId)
	if !allowed {
		return msg, nil, ret
	}

	// 1. 尝试从缓存获取
	respString, err := myredis.GetKeyNilIsError("message_list_" + userId + "_" + sessionId)
	if err == nil && respString != "" {
		var resp []response.GetMessageListResponse
		if err := json.Unmarshal([]byte(respString), &resp); err == nil {
			return "查询消息列表成功", resp, 0
		}
	}

	// 2. 缓存失效，查询数据库
	var messages []model.Message
	if err := db.Where("session_id = ?", sessionId).Order("created_at asc").Find(&messages).Error; err != nil {
		zlog.GetLogger().Error("查询消息列表失败", zap.String("session_id", sessionId), zap.Error(err))
		return constants.SYSTEM_ERROR, nil, -1
	}

	// 3. 转换为响应格式
	var resp []response.GetMessageListResponse
	for _, msg := range messages {
		resp = append(resp, response.GetMessageListResponse{
			SendId:     msg.SendId,
			SendName:   msg.SendName,
			SendAvatar: msg.SendAvatar,
			ReceiveId:  msg.ReceiveId,
			Content:    msg.Content,
			Url:        msg.Url,
			Type:       msg.Type,
			FileType:   msg.FileType,
			FileName:   msg.FileName,
			FileSize:   msg.FileSize,
			AVdata:     msg.AVdata,
		})
	}

	// 4. 写入缓存（有效期 5 分钟）
	if respData, err := json.Marshal(resp); err == nil {
		_ = myredis.SetKeyEx("message_list_"+userId+"_"+sessionId, string(respData), 5*60)
	}

	return "查询消息列表成功", resp, 0
}

// GetGroupMessageList 获取群聊消息列表（通过 groupId）
func (m *messageService) GetGroupMessageList(ctx context.Context, userId string, groupId string) (string, []response.GetGroupMessageListResponse, int) {
	db := dao.NewDbClient(ctx)

	if !m.isGroupMember(ctx, userId, groupId) {
		return "无权访问该群消息", nil, -2
	}

	// 1. 尝试从缓存获取（群聊消息与请求者无关，key 只用 groupId）
	respString, err := myredis.GetKeyNilIsError("group_message_list_" + groupId)
	if err == nil && respString != "" {
		var resp []response.GetGroupMessageListResponse
		if err := json.Unmarshal([]byte(respString), &resp); err == nil {
			return "查询群聊消息列表成功", resp, 0
		}
	}

	// 2. 缓存失效，查询数据库
	var messages []model.Message
	if err := db.Where("receive_id = ?", groupId).Order("created_at asc").Find(&messages).Error; err != nil {
		zlog.GetLogger().Error("查询群聊消息列表失败", zap.String("group_id", groupId), zap.Error(err))
		return constants.SYSTEM_ERROR, nil, -1
	}

	// 3. 转换为响应格式
	var resp []response.GetGroupMessageListResponse
	for _, msg := range messages {
		resp = append(resp, response.GetGroupMessageListResponse{
			SendId:     msg.SendId,
			SendName:   msg.SendName,
			SendAvatar: msg.SendAvatar,
			GroupId:    msg.ReceiveId,
			Content:    msg.Content,
			Url:        msg.Url,
			Type:       msg.Type,
			FileType:   msg.FileType,
			FileName:   msg.FileName,
			FileSize:   msg.FileSize,
			AVdata:     msg.AVdata,
		})
	}

	// 4. 写入缓存
	if respData, err := json.Marshal(resp); err == nil {
		_ = myredis.SetKeyEx("group_message_list_"+groupId, string(respData), 5*60)
	}

	return "查询群聊消息列表成功", resp, 0
}

func (m *messageService) canAccessSessionMessages(ctx context.Context, userId, sessionId string) (bool, string, int) {
	db := dao.NewDbClient(ctx)
	session := model.Session{}
	if err := db.Where("uuid = ?", sessionId).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, "会话不存在", -2
		}
		zlog.GetLogger().Error("查询会话失败", zap.String("session_id", sessionId), zap.String("user_id", userId), zap.Error(err))
		return false, constants.SYSTEM_ERROR, -1
	}

	if session.PairKey != "" && session.PairKey[0] == 'G' {
		if !m.isGroupMember(ctx, userId, session.PairKey) {
			return false, "无权访问该群消息", -2
		}
		return true, "", 0
	}

	parts := strings.Split(session.PairKey, "_")
	if len(parts) != 2 {
		zlog.GetLogger().Warn("非法会话键格式", zap.String("session_id", sessionId), zap.String("pair_key", session.PairKey))
		return false, "会话数据异常", -1
	}
	if parts[0] != userId && parts[1] != userId {
		return false, "无权访问该会话消息", -2
	}
	return true, "", 0
}

func (m *messageService) isGroupMember(ctx context.Context, userId, groupId string) bool {
	db := dao.NewDbClient(ctx)
	contact := model.UserContact{}
	err := db.Where("user_id = ? AND contact_id = ? AND contact_type = ? AND status NOT IN (?, ?)",
		userId, groupId, contact_type_enum.GROUP, contact_status_enum.QUIT_GROUP, contact_status_enum.KICK_OUT_GROUP).
		First(&contact).Error
	return err == nil
}

// allowedFileExts 是允许上传的文件扩展名白名单
var allowedFileExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
	".txt": true, ".zip": true, ".mp4": true, ".mp3": true,
	".wav": true, ".m4a": true, ".aac": true, ".ogg": true, ".webm": true,
}

func (m *messageService) UploadFile(c *gin.Context) (string, int, map[string]string) {
	userId := c.GetString("userId")
	if userId == "" {
		return "用户未登录", -1, nil
	}

	if err := c.Request.ParseMultipartForm(constants.FFILE_MAX_SIZE); err != nil {
		zlog.GetLogger().Error("解析上传文件失败", zap.Error(err))
		return constants.SYSTEM_ERROR, -1, nil
	}
	mForm := c.Request.MultipartForm
	var result map[string]string

	for key := range mForm.File {
		msg, code, res := func() (string, int, map[string]string) {
			file, fileHeader, err := c.Request.FormFile(key)
			if err != nil {
				zlog.GetLogger().Error("获取上传文件失败", zap.Error(err))
				return constants.SYSTEM_ERROR, -1, nil
			}
			defer file.Close()

			// 防止路径遍历攻击：只取文件名，去掉目录部分
			safeName := filepath.Base(fileHeader.Filename)
			ext := strings.ToLower(filepath.Ext(safeName))
			if !allowedFileExts[ext] {
				zlog.GetLogger().Warn("不允许的文件类型", zap.String("ext", ext))
				return "不支持的文件类型", -1, nil
			}

			zlog.GetLogger().Info("接收到上传文件", zap.String("filename", safeName), zap.Int64("size", fileHeader.Size))

			// 生成唯一文件名，防止重名
			fileName := userId + "_" + time.Now().Format("20060102150405") + "_" + safeName
			staticFilePath := config.GetConfig().StaticSrcConfig.StaticFilePath
			if err := os.MkdirAll(staticFilePath, 0755); err != nil {
				zlog.GetLogger().Error("创建文件目录失败", zap.Error(err))
				return constants.SYSTEM_ERROR, -1, nil
			}
			localFileName := staticFilePath + "/" + fileName
			out, err := os.Create(localFileName)
			if err != nil {
				zlog.GetLogger().Error("创建本地文件失败", zap.Error(err))
				return constants.SYSTEM_ERROR, -1, nil
			}
			defer out.Close()
			if _, err := io.Copy(out, file); err != nil {
				zlog.GetLogger().Error("保存上传文件失败", zap.Error(err))
				return constants.SYSTEM_ERROR, -1, nil
			}
			zlog.GetLogger().Info("文件保存成功", zap.String("local_file_name", localFileName))

			downloadUrl := "/static/files/" + fileName
			return "", 0, map[string]string{
				"url":      downloadUrl,
				"filename": safeName,
			}
		}()
		if code != 0 {
			return msg, code, nil
		}
		result = res
		break // 目前只处理第一个文件
	}
	return "文件上传成功", 0, result
}

func (m *messageService) UploadAvatar(c *gin.Context) (string, int, string) {
	userId := c.GetString("userId")
	if userId == "" {
		return "用户未登录", -1, ""
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		zlog.GetLogger().Error("获取上传头像失败", zap.Error(err))
		return "获取文件失败", -1, ""
	}
	file, err := fileHeader.Open()
	if err != nil {
		zlog.GetLogger().Error("打开头像文件失败", zap.Error(err))
		return constants.SYSTEM_ERROR, -1, ""
	}
	defer file.Close()

	// 限制文件大小 10MB
	if fileHeader.Size > 10*1024*1024 {
		return "文件大小不能超过 10MB", -1, ""
	}

	// 验证头像文件扩展名
	allowedAvatarExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	}
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext == "" {
		ext = ".png"
	} else if !allowedAvatarExts[ext] {
		return "只支持 jpg、jpeg、png、gif、webp 格式的头像", -1, ""
	}

	// 生成唯一文件名，防止重名
	fileName := userId + "_" + time.Now().Format("20060102150405") + ext
	localPath := config.GetConfig().StaticSrcConfig.StaticAvatarPath + "/" + fileName

	// 确保目录存在
	if err := os.MkdirAll(config.GetConfig().StaticSrcConfig.StaticAvatarPath, 0755); err != nil {
		zlog.GetLogger().Error("创建头像目录失败", zap.Error(err))
		return constants.SYSTEM_ERROR, -1, ""
	}

	out, err := os.Create(localPath)
	if err != nil {
		zlog.GetLogger().Error("创建本地头像文件失败", zap.Error(err))
		return constants.SYSTEM_ERROR, -1, ""
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		zlog.GetLogger().Error("保存头像文件失败", zap.Error(err))
		return constants.SYSTEM_ERROR, -1, ""
	}

	// 更新数据库中的用户头像路径
	// 这里的路径应该是前端可以直接访问的相对路径，例如 /static/avatars/xxx.png
	avatarUrl := "/static/avatars/" + fileName
	db := dao.NewDbClient(c.Request.Context())
	if err := db.Model(&model.UserInfo{}).Where("uuid = ?", userId).Update("avatar", avatarUrl).Error; err != nil {
		zlog.GetLogger().Error("更新用户头像失败", zap.Error(err), zap.String("userId", userId))
		return "更新数据库失败", -1, ""
	}

	zlog.GetLogger().Info("头像上传成功", zap.String("userId", userId), zap.String("url", avatarUrl))
	return "头像上传成功", 0, avatarUrl
}
