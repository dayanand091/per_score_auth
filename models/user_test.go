package models_test

import (
	"context"
	"fmt"
	"strings"

	"perScoreAuth/models"

	pb "perScoreAuth/perScoreProto/user"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/bxcodec/faker"
	"github.com/jinzhu/gorm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("User", func() {
	Describe("CreateInDb", func() {
		mocket.Catcher.Register()
		var user models.User
		var userRequest pb.CreateUserRequest
		var userRequestLocation pb.CreateUserRequest_Location

		db, err := gorm.Open("postgres", "host=localhost user=perscoreauth dbname=per_score_auth sslmode=disable password=perscoreauth-dm")
		if err != nil {
			fmt.Println("Error:", err)
		}

		userRequestfake := faker.FakeData(&userRequest)
		userRequestLocationfake := faker.FakeData(&userRequestLocation)
		if userRequestfake != nil {
			fmt.Println(userRequestfake)
		}
		if userRequestLocationfake != nil {
			fmt.Println(userRequestLocationfake)
		}
		// userRequest.Email = "12"
		userRequest.Location = &userRequestLocation

		Context("With valid Response", func() {
			It("when the responds is not nil", func() {
				response, err := user.CreateInDB(context.Background(), &userRequest, db)
				data := strings.TrimSpace(response.Message)
				fmt.Println("response message ::", data)
				if err != nil {
					fmt.Println("Error", err)
				}
				Expect(response).NotTo(BeNil())
			})
		})
	})
})
