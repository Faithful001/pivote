package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	service *UserService
}


func NewUserController() *UserController {
	return &UserController{
		service: NewUserService(),
	}
}

func (ctrl *UserController) CreateUser(c *gin.Context) {
	var user User
	
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    err.Error(),
			"data":       nil,
		})
		return
	}
	
	createdUser, err := ctrl.service.CreateUser(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": http.StatusInternalServerError,
			"success":    false,
			"message":    err.Error(),
			"data":       nil,
		})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"statusCode": http.StatusCreated,
		"success":    true,
		"message":    "User created successfully",
		"data":       createdUser,
	})
}

func (ctrl *UserController) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":    false,
			"message":    "Invalid user ID",
			"data":       nil,
		})
		return
	}
	
	user, err := ctrl.service.GetUserByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"statusCode": http.StatusNotFound,
			"success":    false,
			"message":    err.Error(),
			"data":       nil,
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"success":    true,
		"message":    "User fetched successfully",
		"data":       user,
	})
}

func (ctrl *UserController) GetAllUsers(c *gin.Context) {
	users, err := ctrl.service.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": http.StatusInternalServerError,
			"success":  false,
			"message": "Failed to fetch users " + err.Error() ,
			"data":   nil,
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"success":    true,
		"message":    "Users fetched successfully",
		"data":       users,
	})
}

func (ctrl *UserController) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":  false,
			"message": "Invalid user ID " + err.Error(),
			"data":       nil,
		})
		return
	}
	
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":  false,
			"message": "Invalid request body",
			"data":   nil,
		})
		return
	}
	
	user.ID = uint(id)
	
	if err := ctrl.service.UpdateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": http.StatusInternalServerError,
			"success":  false,
			"message": "Failed to update user " + err.Error(),
			"data":       nil,
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"success":  false,
		"message": "User updated successfully",
		"data":    user,
	})
}

func (ctrl *UserController) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"success":  false,
			"message": "Invalid user ID",
			"data": nil,
		})
		return
	}
	
	if err := ctrl.service.DeleteUser(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"statusCode": http.StatusNotFound,
			"success":  false,
			"message": "User not found " + err.Error(),
			"data": nil,
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"success":  false,
		"message": "User deleted successfully",
		"data": nil,
	})
}
