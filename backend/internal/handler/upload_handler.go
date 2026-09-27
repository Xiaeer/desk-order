package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"deskorder/pkg/utils"

	"github.com/gin-gonic/gin"
)

const (
	maxProductImageSize int64 = 5 << 20
	maxUserAvatarSize   int64 = 5 << 20
)

var allowedProductImageExt = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".webp": {},
	".gif":  {},
}

// UploadProductImage 上传商品图片
func UploadProductImage(c *gin.Context) {
	uploadImage(c, "products", maxProductImageSize)
}

// UploadUserAvatar 上传用户头像
func UploadUserAvatar(c *gin.Context) {
	uploadImage(c, "avatars", maxUserAvatarSize)
}

func uploadImage(c *gin.Context, subDir string, maxSize int64) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.Fail(c, 400, "请选择要上传的图片")
		return
	}

	if file.Size > maxSize {
		utils.Fail(c, 400, "图片不能超过 5MB")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if _, ok := allowedProductImageExt[ext]; !ok {
		utils.Fail(c, 400, "仅支持 jpg、jpeg、png、webp、gif 图片")
		return
	}

	contentType := strings.ToLower(file.Header.Get("Content-Type"))
	if !strings.HasPrefix(contentType, "image/") {
		utils.Fail(c, 400, "上传文件必须是图片")
		return
	}

	uploadDir := filepath.Join("uploads", subDir)
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		utils.Fail(c, 500, "创建上传目录失败")
		return
	}

	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join(uploadDir, filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		utils.Fail(c, 500, "保存图片失败")
		return
	}

	urlPath := "/uploads/" + subDir + "/" + filename
	utils.Success(c, gin.H{
		"url": buildPublicFileURL(c, urlPath),
	})
}

func buildPublicFileURL(c *gin.Context, path string) string {
	scheme := "http"
	if c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	host := c.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = c.Request.Host
	}
	return fmt.Sprintf("%s://%s%s", scheme, host, path)
}
