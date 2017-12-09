package models

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	pb "perScoreAuth/perScoreProto/user"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // postgres dialect for gorm
	validator "gopkg.in/go-playground/validator.v9"
)

// User ...
type User struct {
	gorm.Model
	FirstName string
	LastName  string
	Email     string `gorm:"unique"`
	Password  string
	Age       int32
	Role      string
	Location  Location
}

// Location ...
type Location struct {
	gorm.Model
	City    string
	Country string
	UserID  uint
}

// CreateInDB ...
func (user User) CreateInDB(ctx context.Context, in *pb.CreateUserRequest, db *gorm.DB) (*pb.CreateUserResponse, error) {
	var response = new(pb.CreateUserResponse)
	var fieldResponses []*pb.CreateUserResponse_Field

	_, err := CreateUser(in, fieldResponses, db)

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
	const key = "fkzfgk0FY2CaYJhyXbshnPJaRrFtCwfj"
	var response = new(pb.GetSessionResponse)
	email := in.Email
	password := in.Password
	err := db.Where("Email = ? AND Password = ? ", email, password).First(&user).Error
	plaintext := email + "," + "10"
	byteKey := []byte(key)

	errStr := fmt.Sprintf("%s", err) // To convert into string

	if db.RecordNotFound() {
		response.Status = "NOT_FOUND"
		response.Token = ""
		response.Message = errStr
	} else {
		response.Status = "SUCCESS"
		response.Token = Encrypt(byteKey, plaintext)
		response.Message = "Successfully created session"
	}

	return response, err
}

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
func CreateUser(in *pb.CreateUserRequest, fieldResponses []*pb.CreateUserResponse_Field, db *gorm.DB) (User, error) {
	validate := validator.New()
	fieldResponse := new(pb.CreateUserResponse_Field)
	var user User
	var location Location
	user.Email = in.Email
	user.Password = in.Password
	user.Age = in.Age
	user.Role = in.Role

	err := validate.Struct(user)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			fieldResponse.Name = err.StructField()
			fieldResponse.Validation = err.Tag()
			fieldResponses = append(fieldResponses, fieldResponse)
			fmt.Println()
			fmt.Printf("*** Validation Error *** FIELD: %s, TYPE: %s, VALIDATION: %s ====\n\n",
				err.StructField(), err.Type(), err.Tag())
		}

		return user, err
	}

	location.City = in.Location.City
	location.Country = in.Location.Country

	err = validate.Struct(location)

	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			fieldResponse.Name = err.StructField()
			fieldResponse.Validation = err.Tag()
			fieldResponses = append(fieldResponses, fieldResponse)
			fmt.Println()
			fmt.Printf("*** Validation Error *** FIELD: %s, TYPE: %s, VALIDATION: %s ====\n\n",
				err.StructField(), err.Type(), err.Tag())
		}

		return user, err
	}

	err = db.Create(&user).Error
	if err != nil {
		return user, err
	}

	location.UserID = user.ID
	err = db.Create(&location).Error

	return user, err
}
