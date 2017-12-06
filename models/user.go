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
	"regexp"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // postgres dialect for gorm
)

// User ...
type User struct {
	gorm.Model
	Email    string `gorm:"not null;unique"`
	Password string
	Age      int32    `gorm:"not null;"`
	Role     string   `gorm:"not null;"`
	Location Location `gorm:"not null;"`
}

// Location ...
type Location struct {
	gorm.Model
	City    string `gorm:"not null;"`
	Country string `gorm:"not null;"`
	UserID  uint   `gorm:"index"`
}

// CreateInDB ...
func (user User) CreateInDB(ctx context.Context, in *pb.CreateUserRequest, db *gorm.DB) (*pb.CreateUserResponse, error) {

	var response = new(pb.CreateUserResponse)
	var emailvalidate = EmailValidate(in.Email)
	fmt.Println("validate", emailvalidate)

	user.Email = in.Email
	user.Password = in.Password
	user.Age = in.Age
	user.Role = in.Role
	location := Location{
		City:    in.Location.City,
		Country: in.Location.Country,
		UserID:  user.ID,
	}
	err := db.Create(&user).Error
	errStr := fmt.Sprintf("%s", err) // To convert into string

	if err != nil {
		response.Status = "Failed"
		response.Token = ""
		response.Message = errStr
	} else {
		err := db.Create(&location)

		if err.Error != nil {
			response.Status = "Failed"
			response.Token = ""
			response.Message = errStr
		} else {
			response.Status = "Success"
			response.Token = ""
			response.Message = "Successfully Created"
		}
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
		response.Status = "Success"
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

func EmailValidate(email string) bool {
	reg := regexp.MustCompile(".+@.+\\..+")
	matched := reg.Match([]byte(email))
	return matched
}
