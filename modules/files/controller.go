package files

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Gooowan/matchup/modules/core"
	"github.com/Gooowan/matchup/modules/core/auth"
	"github.com/Gooowan/matchup/modules/core/types"
	"github.com/Gooowan/matchup/modules/core/utils"
	filesgen "github.com/Gooowan/matchup/modules/files/gen"
	coremodels "github.com/Gooowan/matchup/modules/users/gen"
)

type FilesController struct {
	coreService *core.CoreService
	fileService *FileService
}

func NewFilesController(coreService *core.CoreService, fileService *FileService) *FilesController {
	return &FilesController{
		coreService: coreService,
		fileService: fileService,
	}
}

func (c *FilesController) UploadAvatar(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	// Store old avatar path for deletion later
	oldAvatarPath := ""
	if avatar, ok := user.ProfileData["avatar"].(string); ok && avatar != "" {
		oldAvatarPath = avatar
	}

	file, header, err := ctx.Request.FormFile("avatar")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Failed to get file from request"})
		return
	}
	defer file.Close()

	if !IsValidImageType(header.Filename) {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid file type. Only jpg, png, and webp files are allowed"})
		return
	}

	if err := ValidateFileSize(header.Size, 2); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	fileExt := header.Filename[len(header.Filename)-4:]
	uploadFile := File{
		Bucket:      "avatars",
		Key:         fmt.Sprintf("%s_%d%s", utils.UUIDToString(user.ID), time.Now().Unix(), fileExt),
		Size:        header.Size,
		Reader:      file,
		ContentType: GetContentType(header.Filename),
	}
	filePath := fmt.Sprintf("%s/%s/%s", c.fileService.PublicEndpoint, uploadFile.Bucket, uploadFile.Key)

	if err := c.fileService.UploadFile(ctx.Request.Context(), uploadFile); err != nil {
		utils.DebugPrint("Failed to upload avatart for user %s, error: %w", utils.UUIDToString(user.ID), err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to upload file"})
		return
	}

	err = c.coreService.Queries.UpdateUserProfileData(ctx.Request.Context(), coremodels.UpdateUserProfileDataParams{
		UserID: user.ID,
		ProfileData: map[string]any{
			"avatar": filePath,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to update user profile"})
		return
	}

	// Delete old avatar after successfully updating the profile
	if oldAvatarPath != "" {
		oldKey := c.fileService.ExtractKeyFromPath(oldAvatarPath)
		if oldKey != "" {
			// Delete asynchronously, don't fail the request if deletion fails
			go func() {
				if err := c.fileService.DeleteFile(context.Background(), "avatars", oldKey); err != nil {
					utils.DebugPrint("Failed to delete old avatar %s for user %s: %v", oldKey, utils.UUIDToString(user.ID), err)
				}
			}()
		}
	}

	ctx.JSON(http.StatusOK, types.Resp{
		Data: gin.H{
			"avatar": filePath,
		},
	})
}

func (c *FilesController) ListVisibleMaterials(ctx *gin.Context) {
	paginationParams := types.ParsePaginationParams(ctx)

	materials, err := c.fileService.Queries.ListVisibleMaterials(ctx.Request.Context(), filesgen.ListVisibleMaterialsParams{
		OffsetVal: int32(paginationParams.Offset()),
		LimitVal:  int32(paginationParams.Limit()),
	})
	if err != nil {
		utils.DebugPrint("Failed to list visible materials: %v", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to list materials"})
		return
	}

	var totalCount int64 = 0
	var responseMaterials []gin.H

	for _, material := range materials {
		totalCount = material.TotalCount
		responseMaterials = append(responseMaterials, gin.H{
			"id":           utils.UUIDToString(material.ID),
			"name":         material.Name,
			"file_size":    material.FileSize,
			"content_type": material.ContentType,
			"created_at":   material.CreatedAt.Time,
		})
	}

	response := types.NewPaginatedResp(paginationParams, responseMaterials, totalCount)
	ctx.JSON(http.StatusOK, response)
}

func (c *FilesController) GetMaterialDownloadURL(ctx *gin.Context) {
	materialIDStr := ctx.Param("id")
	materialID, err := utils.StringToUUID(materialIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid material ID"})
		return
	}

	// Get material and check visibility
	material, err := c.fileService.Queries.GetMaterial(ctx.Request.Context(), materialID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, types.Resp{Error: "Material not found"})
		return
	}

	if !material.Visible {
		ctx.JSON(http.StatusForbidden, types.Resp{Error: "Material is not available"})
		return
	}

	// Generate presigned URL (valid for 15 minutes)
	expiration := 15 * time.Minute
	url, err := c.fileService.GetPresignedURL(ctx.Request.Context(), "materials", material.FileKey, expiration)
	if err != nil {
		utils.DebugPrint("Failed to generate presigned URL: %v", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to generate download URL"})
		return
	}

	expiresAt := time.Now().Add(expiration)
	ctx.JSON(http.StatusOK, types.Resp{
		Data: gin.H{
			"url":        url,
			"expires_at": expiresAt,
		},
	})
}
