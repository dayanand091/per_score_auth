package models

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	pb "perScoreAuth/perScoreProto/user"

	"github.com/chuckpreslar/inflect"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // postgres dialect for gorm
	"github.com/pinzolo/casee"
	validator "gopkg.in/go-playground/validator.v9"
)

// User ...
type User struct {
	gorm.Model
	FirstName string `validate:"required"`
	LastName  string `validate:"required"`
	Email     string `gorm:"unique"`
	Password  string `validate:"required"`
	Age       int32  `validate:"required"`
	Role      string `validate:"required"`
	Location  Location
}

// Location ...
type Location struct {
	gorm.Model
	City    string `validate:"required"`
	Country string `validate:"required"`
	UserID  uint
}

// CreateInDB ...
func (user User) CreateInDB(ctx context.Context, in *pb.CreateUserRequest, db *gorm.DB) (*pb.CreateUserResponse, error) {
	var response = new(pb.CreateUserResponse)
	var fieldResponses []*pb.CreateUserResponse_Field

	fieldResponses, err := CreateUser(in, fieldResponses, db)

	if err != nil {
		response.Status = "FAILURE"
		response.Token = ""
		response.Message = "Signup failed. Please try again."
		response.Fields = fieldResponses
	} else {
		response.Status = "SUCCESS"
		response.Token = ""
		response.Message = "You have signed up successfully!"
	}

	return response, err
}

// CreateSession ...
func (user User) CreateSession(sctx context.Context, in *pb.GetSessionRequest, db *gorm.DB) (*pb.GetSessionResponse, error) {
	var err error
	const key = "fkzfgk0FY2CaYJhyXbshnPJaRrFtCwfj"
	var sessionInMinutes = "10"
	var response = new(pb.GetSessionResponse)
	result := db.Where("email = ? AND password = ?", in.Email, in.Password).First(&user).RecordNotFound()
	plaintext := user.Email + "," + user.Role + "," + sessionInMinutes
	byteKey := []byte(key)

	if result == true {
		response.Status = "FAILURE"
		response.Token = ""
		response.Message = "Invalid email and password combination!"
		err = errors.New(response.Message)
	} else {
		response.Status = "SUCCESS"
		response.Token = Encrypt(byteKey, plaintext)
		response.Message = "Logged in successfully!"
		err = nil
	}

	return response, err
}

// Encrypt ...
func Encrypt(key []byte, text string) string {
	plaintext := []byte(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	// convert to base64
	return base64.URLEncoding.EncodeToString(ciphertext)
}

// func EmailValidate(email string) bool {
// 	reg := regexp.MustCompile(".+@.+\\..+")
// 	matched := reg.Match([]byte(email))
// 	return matched
// }

// CreateUser ...
func CreateUser(in *pb.CreateUserRequest, fieldResponses []*pb.CreateUserResponse_Field, db *gorm.DB) ([]*pb.CreateUserResponse_Field, error) {
	validate := validator.New()
	var user User
	user.FirstName = in.FirstName
	user.LastName = in.LastName
	user.Email = in.Email
	user.Password = in.Password
	user.Age = in.Age
	user.Role = in.Role

	user.Location.City = in.Location.City
	user.Location.Country = in.Location.Country

	err := validate.Struct(user)
	if err != nil {
		for _, errV := range err.(validator.ValidationErrors) {
			fieldResponse := new(pb.CreateUserResponse_Field)
			fieldResponse.Name = casee.ToSnakeCase(errV.StructField())
			fieldResponse.Validation = inflect.Titleize(errV.Tag())
			fieldResponses = append(fieldResponses, fieldResponse)
			fmt.Println("*** Validation Errors ***")
			// fmt.Printf("*** Validation Error *** STRUCT: %s, FIELD: %s, VALIDATION: %s ====\n\n",
			// 	errV.Namespace(), errV.StructField(), errV.Tag())
			fmt.Println("NAMESPACE:", errV.Namespace())
			fmt.Println("FIELD:", errV.Field())
			fmt.Println("TAG:", errV.Tag())
			fmt.Println("TYPE:", errV.Type())
			fmt.Println("VALUE:", errV.Value())
			fmt.Println("PARAM:", errV.Param())
			fmt.Println()
		}

		fmt.Println("Fields:", fieldResponses)

		return fieldResponses, err
	}

	fmt.Println("Fields:", fieldResponses)

	err = db.Create(&user).Error
	if err != nil {
		return fieldResponses, err
	}

	return fieldResponses, err
}
