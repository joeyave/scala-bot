package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joeyave/scala-bot/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type UserController struct {
	UserService *service.UserService
}

type User struct {
	ID     int64    `json:"id"`
	Name   string   `json:"name"`
	Events []*Event `json:"events"`

	//Events2 map[time.Weekday]map[string]int `json:"events2"`
}

type Event struct {
	ID      primitive.ObjectID `json:"id"`
	Date    string             `json:"date"`
	Weekday time.Weekday       `json:"weekday"`
	Name    string             `json:"name"`
	Roles   []*Role            `json:"roles"`
}

type Role struct {
	ID   primitive.ObjectID `json:"id"`
	Name string             `json:"name"`
}

func (c *UserController) UsersWithEvents(ctx *gin.Context) {

	fmt.Println(ctx.Request.URL)

	hex := ctx.Query("bandId")
	bandID, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		ctx.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}

	from := ctx.Query("from")
	fromDate, err := time.Parse("02.01.2006", from)
	if err != nil {
		fromDate = time.Date(time.Now().Year(), time.January, 1, 0, 0, 0, 0, time.Local)
	}

	users, err := c.UserService.FindManyExtraByBandID(bandID, fromDate, time.Now())
	if err != nil {
		return
	}

	var viewUsers []*User
	for _, user := range users {
		viewUser := &User{
			ID:   user.ID,
			Name: user.Name,
		}

		for _, event := range user.Events {
			viewEvent := &Event{
				ID:      event.ID,
				Date:    event.Time.Format("2006-01-02"),
				Weekday: event.Time.Weekday(),
				Name:    event.Name,
			}

			for _, membership := range event.Memberships {
				if membership.UserID == user.ID {
					viewRole := &Role{
						ID:   membership.Role.ID,
						Name: membership.Role.Name,
					}
					viewEvent.Roles = append(viewEvent.Roles, viewRole)
					break
				}
			}

			viewUser.Events = append(viewUser.Events, viewEvent)
		}

		viewUsers = append(viewUsers, viewUser)
	}

	ctx.JSON(200, gin.H{
		"users": viewUsers,
	})
}
