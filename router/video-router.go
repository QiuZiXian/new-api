package router

import (
	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/middleware"

	"github.com/gin-gonic/gin"
)

func SetVideoRouter(router *gin.Engine) {
	// Video proxy: accepts either session auth (dashboard) or token auth (API clients)
	videoProxyRouter := router.Group("/v1")
	videoProxyRouter.Use(middleware.RouteTag("relay"))
	videoProxyRouter.Use(middleware.TokenOrUserAuth())
	{
		videoProxyRouter.GET("/videos/:task_id/content", controller.VideoProxy)
	}

	videoV1Router := router.Group("/v1")
	videoV1Router.Use(middleware.RouteTag("relay"))
	videoV1Router.Use(middleware.TokenAuth(), middleware.Distribute())
	{
		videoV1Router.POST("/video/generations", controller.RelayTask)
		videoV1Router.GET("/video/generations/:task_id", controller.RelayTaskFetch)
		videoV1Router.POST("/videos/:video_id/remix", controller.RelayTask)
	}
	// openai compatible API video routes
	// docs: https://platform.openai.com/docs/api-reference/videos/create
	{
		videoV1Router.POST("/videos", controller.RelayTask)
		videoV1Router.GET("/videos/:task_id", controller.RelayTaskFetch)
	}

	// Seedance/CII-style content generation task list & cancel. Backed by the
	// caller's own local task data (see controller/task_content.go), not a proxy
	// to a shared upstream account.
	contentTaskV1Router := router.Group("/v1/contents/generations/tasks")
	contentTaskV1Router.Use(middleware.RouteTag("relay"))
	contentTaskV1Router.Use(middleware.TokenAuth())
	{
		contentTaskV1Router.GET("", controller.ListContentTasks)
		contentTaskV1Router.DELETE("/:task_id", controller.CancelContentTask)
	}

	// 字幕擦除（独立路由，不与视频生成共享 /v1/videos）。
	// Distribute 按 model 字段把请求分到 cii-subtitle-erase 模型对应的
	// ChannelTypeCiiSubtitleErase 渠道。复用 controller.RelayTask / RelayTaskFetch
	// 走标准任务提交/查询流程，最终命中 ciisubtitleerase.TaskAdaptor。
	subtitleEraseV1Router := router.Group("/v1/videos/subtitle-erase/tasks")
	subtitleEraseV1Router.Use(middleware.RouteTag("relay"))
	subtitleEraseV1Router.Use(middleware.TokenAuth(), middleware.Distribute())
	{
		subtitleEraseV1Router.POST("", controller.RelayTask)
		subtitleEraseV1Router.GET("/:task_id", controller.RelayTaskFetch)
	}

	// 素材组 / 素材 CRUD
	// 仅 TokenAuth：与 controller/task_content.go 一样，不挂 Distribute()
	// （素材 CRUD 不需要按模型分发渠道）。
	assetV1Router := router.Group("/v1")
	assetV1Router.Use(middleware.RouteTag("relay"))
	assetV1Router.Use(middleware.TokenAuth())
	{
		// 素材组
		assetV1Router.POST("/asset-groups", controller.CreateAssetGroup)
		assetV1Router.GET("/asset-groups", controller.ListAssetGroups)
		assetV1Router.GET("/asset-groups/:group_id", controller.GetAssetGroup)
		assetV1Router.PUT("/asset-groups/:group_id", controller.UpdateAssetGroup)
		assetV1Router.DELETE("/asset-groups/:group_id", controller.DeleteAssetGroup)

		// 素材（强依赖素材组，path 上保留 group_id）
		assetV1Router.POST("/asset-groups/:group_id/assets", controller.CreateAsset)
		assetV1Router.GET("/asset-groups/:group_id/assets", controller.ListAssets)
		assetV1Router.GET("/assets/:asset_id", controller.GetAsset)
		assetV1Router.PUT("/assets/:asset_id", controller.UpdateAsset)
		assetV1Router.DELETE("/assets/:asset_id", controller.DeleteAsset)
	}

	klingV1Router := router.Group("/kling/v1")
	klingV1Router.Use(middleware.RouteTag("relay"))
	klingV1Router.Use(middleware.KlingRequestConvert(), middleware.TokenAuth(), middleware.Distribute())
	{
		klingV1Router.POST("/videos/text2video", controller.RelayTask)
		klingV1Router.POST("/videos/image2video", controller.RelayTask)
		klingV1Router.GET("/videos/text2video/:task_id", controller.RelayTaskFetch)
		klingV1Router.GET("/videos/image2video/:task_id", controller.RelayTaskFetch)
	}

	// Jimeng official API routes - direct mapping to official API format
	jimengOfficialGroup := router.Group("jimeng")
	jimengOfficialGroup.Use(middleware.RouteTag("relay"))
	jimengOfficialGroup.Use(middleware.JimengRequestConvert(), middleware.TokenAuth(), middleware.Distribute())
	{
		// Maps to: /?Action=CVSync2AsyncSubmitTask&Version=2022-08-31 and /?Action=CVSync2AsyncGetResult&Version=2022-08-31
		jimengOfficialGroup.POST("/", controller.RelayTask)
	}
}
