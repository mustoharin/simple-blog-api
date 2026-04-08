package http

import (
	"github.com/gin-gonic/gin"

	"simple-blog-api/internal/delivery/http/handler"
	"simple-blog-api/internal/delivery/http/middleware"
)

type RouterDeps struct {
	JWTSecret      string
	AllowedOrigins []string
	Auth           *handler.AuthHandler
	User           *handler.UserHandler
	Profile        *handler.ProfileHandler
	Post           *handler.PostHandler
	Tag            *handler.TagHandler
	Comment        *handler.CommentHandler
	Image          *handler.ImageHandler
	Audit          *handler.AuditHandler
	Dashboard      *handler.DashboardHandler
}

func SetupRouter(deps RouterDeps) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS(deps.AllowedOrigins))

	v1 := r.Group("/api/v1")

	authRoutes := v1.Group("/auth")
	{
		authRoutes.POST("/register", deps.Auth.Register)
		authRoutes.POST("/login", deps.Auth.Login)
		authRoutes.POST("/refresh", deps.Auth.Refresh)
		authRoutes.POST("/forgot-password", deps.Auth.ForgotPassword)
		authRoutes.POST("/reset-password", deps.Auth.ResetPassword)
		authRoutes.POST("/accept-invitation", deps.Auth.AcceptInvitation)
		authRoutes.POST("/logout", middleware.JWT(deps.JWTSecret), deps.Auth.Logout)
	}

	postRoutes := v1.Group("/posts")
	{
		postRoutes.GET("", deps.Post.ListPosts)
		postRoutes.GET("/:slug", deps.Post.GetPost)
		postRoutes.POST("", middleware.JWT(deps.JWTSecret), middleware.RequirePermission("post:create"), deps.Post.CreatePost)
		postRoutes.PUT("/:id", middleware.JWT(deps.JWTSecret), middleware.RequirePermission("post:edit"), deps.Post.UpdatePost)
		postRoutes.PATCH("/:id/publish", middleware.JWT(deps.JWTSecret), middleware.RequirePermission("post:publish"), deps.Post.TogglePublish)
		postRoutes.DELETE("/:id", middleware.JWT(deps.JWTSecret), middleware.RequirePermission("post:delete"), deps.Post.DeletePost)
		postRoutes.GET("/:id/comments", deps.Comment.ListComments)
		postRoutes.POST("/:id/comments", middleware.JWT(deps.JWTSecret), middleware.RequirePermission("comment:create"), deps.Comment.CreateComment)
	}

	tagRoutes := v1.Group("/tags")
	{
		tagRoutes.GET("", deps.Tag.ListTags)
		tagRoutes.POST("", middleware.JWT(deps.JWTSecret), middleware.RequirePermission("post:create"), deps.Tag.CreateTag)
		tagRoutes.DELETE("/:id", middleware.JWT(deps.JWTSecret), middleware.RequirePermission("post:delete"), deps.Tag.DeleteTag)
	}

	commentRoutes := v1.Group("/comments")
	{
		commentRoutes.PATCH("/:id/status", middleware.JWT(deps.JWTSecret), middleware.RequirePermission("comment:approve"), deps.Comment.UpdateCommentStatus)
		commentRoutes.DELETE("/:id", middleware.JWT(deps.JWTSecret), middleware.RequirePermission("comment:approve"), deps.Comment.DeleteComment)
	}

	imageRoutes := v1.Group("/images")
	{
		imageRoutes.POST("", middleware.JWT(deps.JWTSecret), middleware.RequirePermission("image:upload"), deps.Image.UploadImage)
		imageRoutes.DELETE("/:id", middleware.JWT(deps.JWTSecret), middleware.RequirePermission("image:upload"), deps.Image.DeleteImage)
	}

	userRoutes := v1.Group("/users", middleware.JWT(deps.JWTSecret))
	{
		userRoutes.GET("", middleware.RequirePermission("user:manage"), deps.User.ListUsers)
		userRoutes.POST("", middleware.RequirePermission("user:create"), deps.User.CreateUser)
		userRoutes.PUT("/:id", middleware.RequirePermission("user:update"), deps.User.UpdateUser)
		userRoutes.DELETE("/:id", middleware.RequirePermission("user:delete"), deps.User.DeleteUser)
		userRoutes.POST("/:id/roles", middleware.RequirePermission("user:manage"), deps.User.AssignRole)
		userRoutes.DELETE("/:id/roles/:roleId", middleware.RequirePermission("user:manage"), deps.User.RemoveRole)
		userRoutes.POST("/:id/resend-invitation", middleware.RequirePermission("user:manage"), deps.User.ResendInvitation)
	}

	meRoutes := v1.Group("/me", middleware.JWT(deps.JWTSecret))
	{
		meRoutes.GET("", deps.Profile.GetMe)
		meRoutes.PATCH("", deps.Profile.UpdateMe)
		meRoutes.POST("/change-password", deps.Profile.ChangePassword)
	}

	auditRoutes := v1.Group("/audit-logs", middleware.JWT(deps.JWTSecret), middleware.RequirePermission("audit:read"))
	{
		auditRoutes.GET("", deps.Audit.ListAuditLogs)
		auditRoutes.GET("/:id", deps.Audit.GetAuditLog)
	}

	v1.GET("/dashboard",
		middleware.JWT(deps.JWTSecret),
		middleware.RequirePermission("dashboard:read"),
		deps.Dashboard.GetDashboard,
	)

	return r
}
