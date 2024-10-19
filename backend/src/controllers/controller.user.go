package controllers

import (
	"net/http"
	"context"
	User "common/models/user"
	Batch "common/models/batch"
	Test "common/models/test"
	"server/src/helper"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func (c *ControllerClass) UserLoginHandler(ctx *gin.Context, loginRequest *User.UserLoginRequest) {
	var foundUser User.User

	err := c.UserCollection.FindOne(context.Background(), bson.M{"username": loginRequest.Username}).Decode(&foundUser)
    if err != nil {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

	if foundUser.Password != loginRequest.Password {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

	var batch Batch.Batch
    err = c.BatchCollection.FindOne(context.Background(), bson.M{"name": foundUser.Batch}).Decode(&batch)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch batch information"})
        return
    }

	var tests []Test.Test
    cursor, err := c.TestCollection.Find(context.Background(), bson.M{"_id": bson.M{"$in": batch.Tests}})
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tests"})
        return
    }
    defer cursor.Close(context.Background())

    if err = cursor.All(context.Background(), &tests); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode tests"})
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "message": "Login successful",
        "user": foundUser,
        "tests": tests,
    })
}

func (this *ControllerClass) UpdateUserData(ctx *gin.Context, userUpdateRequest *User.UserUpdateRequest) {
	userCollection := this.UserCollection
	err := helper.UpdateUserData(userCollection, userUpdateRequest)

	if err != nil {
		ctx.JSON(500, gin.H{
			"message": "Error in updating user data",
			"error":   err,
		})
		return
	}

	ctx.JSON(200, gin.H{
		"message": "User data updated successfully",
	})
}

func (this *ControllerClass) Increase_Time(ctx *gin.Context, param string, username []string, time_to_increase int64) {
	userCollection := this.UserCollection

	if len(username) == 0 {
		ctx.JSON(500, gin.H{
			"message": "Empty username",
		})
		return
	}

	if len(username) > 1 {
		param = "batch"
	}

	switch param {
	case "user":
		err := helper.UpdateUserTestTime(userCollection, username[0], time_to_increase)
		if err != nil {
			ctx.JSON(500, gin.H{
				"message": "Error in increasing time",
				"error":   err,
			})
		}
		ctx.JSON(200, gin.H{
			"message": "Time increased successfully",
		})

	case "batch":

		err := helper.UpdateBatchTestTime(userCollection, username, time_to_increase)
		if err != nil {
			ctx.JSON(500, gin.H{
				"message": "Error in increasing time",
				"error":   err,
			})
		}
		ctx.JSON(200, gin.H{
			"message": "Time increased successfully",
		})

	default:
		ctx.JSON(500, gin.H{
			"message": "Invalid parameter",
		})
	}

}

func (this *ControllerClass) GetBatchWiseData(ctx *gin.Context, param string, BatchNumber string, Ranges []int) {
	userCollection := this.UserCollection

	switch param {
	case "batch":
		result, err := helper.GetBatchWiseList(userCollection, BatchNumber)
		if err != nil {
			ctx.JSON(500, gin.H{
				"message": "Error in fetching batch wise data",
				"error":   err,
			})
		}
		ctx.JSON(200, gin.H{
			"message": "Batch wise data fetched successfully",
			"data":    result,
		})

	case "roll":
		From := Ranges[0]
		To := Ranges[1]
		result, err := helper.GetBatchWiseListRoll(userCollection, BatchNumber, From, To)
		if err != nil {
			ctx.JSON(500, gin.H{
				"message": "Error in fetching batch wise data",
				"error":   err,
			})
		}

		ctx.JSON(200, gin.H{
			"message": "Batch wise data fetched successfully",
			"data":    result,
		})

	case "frontend":
		result, err := helper.GetBatchDataForFrontend(userCollection, BatchNumber)
		if err != nil {
			ctx.JSON(500, gin.H{
				"message": "Error in fetching batch wise data",
				"error":   err,
			})
		}

		ctx.JSON(200, gin.H{
			"message": "Batch wise data fetched successfully",
			"data":    result,
		})

	default:
		ctx.JSON(500, gin.H{
			"message": "Invalid parameter",
		})
	}
}

func (this *ControllerClass) SetUserData(ctx *gin.Context, param string, userRequest *User.UserBatchRequestData, Username string) {
	userCollection := this.UserCollection

	switch param {
	case "download":
		err := helper.SetUserResultToDownloaded(userCollection, userRequest)
		if err != nil {
			ctx.JSON(500, gin.H{
				"message": "Error in setting user data",
				"error":   err,
			})
		}

		ctx.JSON(200, gin.H{
			"message": "User data set successfully",
		})

	case "reset":
		err := helper.ResetUserData(userCollection, Username)
		if err != nil {
			ctx.JSON(500, gin.H{
				"message": "Error in resetting user data",
				"error":   err,
			})
		}

		ctx.JSON(200, gin.H{
			"message": "User data reset successfully",
		})

	default:
		ctx.JSON(500, gin.H{
			"message": "Invalid parameter",
		})
	}

}

func (self *ControllerClass) UpdateUser(ctx *gin.Context, userRequest *User.UserModelUpdateRequest) error{
	userCollection := self.UserCollection

	err := helper.UpdateUser(userCollection, userRequest)
	if err != nil{
		return err
	}

	return nil
}


func (self *ControllerClass) DeleteUser(ctx *gin.Context, userId string) error{
	userCollection := self.UserCollection

	err := helper.Delete_Model_By_ID(userCollection, userId)

	if err != nil{
		return err
	}

	return nil;
}
