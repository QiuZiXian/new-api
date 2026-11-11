package controller

import (
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

func GetModelTags(c *gin.Context) {
	tags, err := model.GetAllModelTags()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, tags)
}

func CreateModelTag(c *gin.Context) {
	var payload model.ModelTag
	if err := c.ShouldBindJSON(&payload); err != nil {
		common.ApiError(c, err)
		return
	}
	name := strings.TrimSpace(payload.Name)
	if !model.IsValidModelTagName(name) {
		common.ApiErrorMsg(c, "标签名称无效或超过 64 个字符")
		return
	}
	if duplicated, err := model.IsModelTagDuplicated(0, name); err != nil {
		common.ApiError(c, err)
		return
	} else if duplicated {
		common.ApiErrorMsg(c, "标签已存在")
		return
	}

	sort := payload.Sort
	if sort == 0 {
		existing, err := model.GetAllModelTags()
		if err != nil {
			common.ApiError(c, err)
			return
		}
		for _, tag := range existing {
			if tag.Sort >= sort {
				sort = tag.Sort + 1
			}
		}
	}
	tag, err := model.InsertModelTag(name, sort)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, tag)
}

func UpdateModelTag(c *gin.Context) {
	var payload model.ModelTag
	if err := c.ShouldBindJSON(&payload); err != nil {
		common.ApiError(c, err)
		return
	}
	if payload.Id == 0 {
		common.ApiErrorMsg(c, "缺少标签 ID")
		return
	}
	name := strings.TrimSpace(payload.Name)
	if !model.IsValidModelTagName(name) {
		common.ApiErrorMsg(c, "标签名称无效或超过 64 个字符")
		return
	}
	if duplicated, err := model.IsModelTagDuplicated(payload.Id, name); err != nil {
		common.ApiError(c, err)
		return
	} else if duplicated {
		common.ApiErrorMsg(c, "标签已存在")
		return
	}
	if err := model.UpdateModelTagById(payload.Id, name, payload.Sort); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, payload)
}

func DeleteModelTag(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if err := model.DeleteModelTagById(id); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

func ImportModelTags(c *gin.Context) {
	added, err := model.ImportModelTagsFromModels()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"added": added})
}
