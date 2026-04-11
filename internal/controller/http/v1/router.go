package v1
import (
	"github.com/gin-gonic/gin"
	"practice-7/internal/usecase"
	"practice-7/pkg/logger"
)


func RegisterRoutes(rg *gin.RouterGroup, uc usecase.UserInterface, log logger.Interface) {
	newUserRoutes(rg, uc, log)
}
