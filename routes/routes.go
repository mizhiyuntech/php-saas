package routes

import (
	"yuyue-auth/controllers"
	"yuyue-auth/middleware"

	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine) {
	api := r.Group("/api")

	api.GET("/install/status", controllers.GetInstallStatus)
	api.POST("/install", controllers.Install)

	api.GET("/site-info", controllers.GetSiteInfo)
	api.POST("/license/verify", controllers.VerifyLicense)
	api.POST("/piracy/check", controllers.CheckPiracy)

	api.GET("/payment/callback/:type/:action", controllers.HandlePaymentCallback)
	api.POST("/payment/callback/:type/:action", controllers.HandlePaymentCallback)

	pub := api.Group("/public")
	{
		pub.GET("/programs", controllers.PublicListPrograms)
		pub.GET("/packages", controllers.PublicListPackages)
		pub.GET("/payment-methods", controllers.PublicGetPaymentMethods)
		pub.POST("/order", controllers.PublicCreateOrder)
		pub.GET("/order/:order_no", controllers.PublicGetOrder)
	}

	auth := api.Group("")
	auth.Use(middleware.InstallCheck())
	{
		auth.POST("/auth/login", controllers.Login)
	}

	admin := api.Group("")
	admin.Use(middleware.InstallCheck(), middleware.AuthRequired())
	{
		admin.GET("/auth/profile", controllers.GetProfile)
		admin.POST("/auth/change-password", controllers.ChangePassword)

		admin.GET("/dashboard/stats", controllers.GetDashboardStats)

		admin.GET("/programs", controllers.ListPrograms)
		admin.GET("/programs/all", controllers.GetAllPrograms)
		admin.GET("/programs/:id", controllers.GetProgram)
		admin.POST("/programs", controllers.CreateProgram)
		admin.PUT("/programs/:id", controllers.UpdateProgram)
		admin.DELETE("/programs/:id", controllers.DeleteProgram)

		admin.GET("/licenses", controllers.ListLicenses)
		admin.POST("/licenses", controllers.CreateLicenses)
		admin.PUT("/licenses/:id", controllers.UpdateLicense)
		admin.DELETE("/licenses/:id", controllers.DeleteLicense)

		admin.GET("/packages", controllers.ListPackages)
		admin.POST("/packages", controllers.CreatePackage)
		admin.PUT("/packages/:id", controllers.UpdatePackage)
		admin.DELETE("/packages/:id", controllers.DeletePackage)

		admin.GET("/injection/languages", controllers.GetLanguages)
		admin.POST("/injection/upload", controllers.InjectAuthorization)

		admin.GET("/finance/payment-configs", controllers.GetPaymentConfigs)
		admin.PUT("/finance/payment-configs", controllers.UpdatePaymentConfig)
		admin.GET("/finance/orders", controllers.ListOrders)
		admin.GET("/finance/order-stats", controllers.GetOrderStats)

		admin.GET("/settings", controllers.GetSettings)
		admin.PUT("/settings", controllers.UpdateSettings)
		admin.POST("/settings/upload-icon", controllers.UploadIcon)
		admin.POST("/settings/upload-favicon", controllers.UploadFavicon)
		admin.GET("/settings/unauth-page", controllers.GetUnauthPageHTML)
		admin.PUT("/settings/unauth-page", controllers.UpdateUnauthPageHTML)
		admin.GET("/settings/piracy-page", controllers.GetPiracyPageHTML)
		admin.PUT("/settings/piracy-page", controllers.UpdatePiracyPageHTML)

		admin.GET("/piracy", controllers.ListPiracy)
		admin.POST("/piracy", controllers.CreatePiracy)
		admin.PUT("/piracy/:id", controllers.UpdatePiracy)
		admin.DELETE("/piracy/:id", controllers.DeletePiracy)
	}
}
