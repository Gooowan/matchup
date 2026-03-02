package files

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Gooowan/matchup/modules/core/types"
	"github.com/Gooowan/matchup/modules/core/utils"
	filesgen "github.com/Gooowan/matchup/modules/files/gen"
)

type FileAdminController struct {
	fileService *FileService
}

func NewFileAdminController(fileService *FileService) *FileAdminController {
	return &FileAdminController{
		fileService: fileService,
	}
}

func (c *FileAdminController) UploadMaterial(ctx *gin.Context) {
	name := ctx.PostForm("name")
	if name == "" {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Material name is required"})
		return
	}

	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Failed to get file from request"})
		return
	}
	defer file.Close()

	// Validate file size (50MB max)
	if err := ValidateFileSize(header.Size, 50); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	// Extract file extension
	fileExt := ""
	if len(header.Filename) > 0 {
		for i := len(header.Filename) - 1; i >= 0 && i > len(header.Filename)-10; i-- {
			if header.Filename[i] == '.' {
				fileExt = header.Filename[i:]
				break
			}
		}
	}

	// Create database record first to get PostgreSQL-generated UUID
	material, err := c.fileService.Queries.CreateMaterial(ctx.Request.Context(), filesgen.CreateMaterialParams{
		Name:        name,
		FileKey:     "pending", // Temporary, will update after upload
		FileSize:    header.Size,
		ContentType: GetContentType(header.Filename),
		Visible:     false, // Default to not visible
	})
	if err != nil {
		utils.DebugPrint("Failed to create material record: %v", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to create material record"})
		return
	}

	// Generate unique file key using the PostgreSQL-generated UUID
	fileKey := fmt.Sprintf("%s_%d%s", utils.UUIDToString(material.ID), time.Now().Unix(), fileExt)

	uploadFile := File{
		Bucket:      "materials",
		Key:         fileKey,
		Size:        header.Size,
		Reader:      file,
		ContentType: GetContentType(header.Filename),
	}

	// Upload to S3
	if err := c.fileService.UploadFile(ctx.Request.Context(), uploadFile); err != nil {
		// Delete DB record on upload failure
		go func() {
			if err := c.fileService.Queries.DeleteMaterial(context.Background(), material.ID); err != nil {
				utils.DebugPrint("Failed to cleanup DB record: %v", err)
			}
		}()
		utils.DebugPrint("Failed to upload material: %v", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to upload file"})
		return
	}

	// Update the file key in the database
	if err := c.fileService.Queries.UpdateMaterialName(ctx.Request.Context(), filesgen.UpdateMaterialNameParams{
		ID:   material.ID,
		Name: name,
	}); err != nil {
		// Cleanup both S3 and DB on update failure
		go func() {
			c.fileService.DeleteFile(context.Background(), "materials", fileKey)
			c.fileService.Queries.DeleteMaterial(context.Background(), material.ID)
		}()
		utils.DebugPrint("Failed to update material file key: %v", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to finalize material"})
		return
	}

	// Update file_key separately since we don't have a dedicated query for it
	_, err = c.fileService.DB.Exec(ctx.Request.Context(),
		"UPDATE marketing_materials SET file_key = $1 WHERE id = $2",
		fileKey, material.ID)
	if err != nil {
		utils.DebugPrint("Failed to update file key: %v", err)
		// Continue anyway since the material is already uploaded
	}

	ctx.JSON(http.StatusOK, types.Resp{
		Data: gin.H{
			"id":           utils.UUIDToString(material.ID),
			"name":         material.Name,
			"file_size":    material.FileSize,
			"content_type": material.ContentType,
			"visible":      material.Visible,
			"created_at":   material.CreatedAt.Time,
		},
	})
}

func (c *FileAdminController) DeleteMaterial(ctx *gin.Context) {
	materialIDStr := ctx.Param("id")
	materialID, err := utils.StringToUUID(materialIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid material ID"})
		return
	}

	// Get material to retrieve file key
	material, err := c.fileService.Queries.GetMaterial(ctx.Request.Context(), materialID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, types.Resp{Error: "Material not found"})
		return
	}

	// Delete from database
	if err := c.fileService.Queries.DeleteMaterial(ctx.Request.Context(), materialID); err != nil {
		utils.DebugPrint("Failed to delete material from DB: %v", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to delete material"})
		return
	}

	// Delete from S3 (async, even if it fails we already deleted from DB)
	go func() {
		if err := c.fileService.DeleteFile(context.Background(), "materials", material.FileKey); err != nil {
			utils.DebugPrint("Failed to delete file %s from S3: %v", material.FileKey, err)
		}
	}()

	ctx.JSON(http.StatusOK, types.Resp{Data: "Material deleted successfully"})
}

func (c *FileAdminController) UpdateMaterialName(ctx *gin.Context) {
	materialIDStr := ctx.Param("id")
	materialID, err := utils.StringToUUID(materialIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid material ID"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required,min=1,max=255"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	if err := c.fileService.Queries.UpdateMaterialName(ctx.Request.Context(), filesgen.UpdateMaterialNameParams{
		ID:   materialID,
		Name: req.Name,
	}); err != nil {
		utils.DebugPrint("Failed to update material name: %v", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to update material name"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: "Material name updated successfully"})
}

func (c *FileAdminController) UpdateMaterialVisibility(ctx *gin.Context) {
	materialIDStr := ctx.Param("id")
	materialID, err := utils.StringToUUID(materialIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid material ID"})
		return
	}

	var req struct {
		Visible bool `json:"visible"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	if err := c.fileService.Queries.UpdateMaterialVisibility(ctx.Request.Context(), filesgen.UpdateMaterialVisibilityParams{
		ID:      materialID,
		Visible: req.Visible,
	}); err != nil {
		utils.DebugPrint("Failed to update material visibility: %v", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to update material visibility"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: "Material visibility updated successfully"})
}

func (c *FileAdminController) ListAllMaterials(ctx *gin.Context) {
	paginationParams := types.ParsePaginationParams(ctx)

	materials, err := c.fileService.Queries.ListMaterials(ctx.Request.Context(), filesgen.ListMaterialsParams{
		OffsetVal: int32(paginationParams.Offset()),
		LimitVal:  int32(paginationParams.Limit()),
	})
	if err != nil {
		utils.DebugPrint("Failed to list materials: %v", err)
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
			"visible":      material.Visible,
			"created_at":   material.CreatedAt.Time,
			"updated_at":   material.UpdatedAt.Time,
		})
	}

	response := types.NewPaginatedResp(paginationParams, responseMaterials, totalCount)
	ctx.JSON(http.StatusOK, response)
}

func (c *FileAdminController) RegisterRoutes(rg *gin.RouterGroup, adminAuthMiddleware gin.HandlerFunc) {
	rg.Use(adminAuthMiddleware)
	rg.POST("/marketing/upload", c.UploadMaterial)
	rg.DELETE("/marketing/:id", c.DeleteMaterial)
	rg.PUT("/marketing/:id/name", c.UpdateMaterialName)
	rg.PUT("/marketing/:id/visibility", c.UpdateMaterialVisibility)
	rg.GET("/marketing", c.ListAllMaterials)
}
