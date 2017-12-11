package server

import (
	"fmt"
	"os"

	"golang.org/x/net/context"

	"perScoreAuth/models"
	pb "perScoreAuth/perScoreProto/user"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// Server ...
type Server struct {
	User models.User
}

// CreateUser ...
func (s *Server) CreateUser(ctx context.Context, in *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	fmt.Println("Request:", in)
	var result *pb.CreateUserResponse
	dbString := fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=%s", os.Getenv("DEV_HOST"), os.Getenv("DEV_DBNAME"), os.Getenv("DEV_USERNAME"), os.Getenv("DEV_PASSWORD"), os.Getenv("DEV_SSLMODE"))
	db, err := gorm.Open(os.Getenv("DEV_DB_DRIVER"), dbString)
	defer db.Close()
	if err != nil {
		log.Errorf("Error opening DB connection: %+v", err)
	} else {
		result, _ = s.User.CreateInDB(ctx, in, db)
	}
	fmt.Println("Result", result)
	return result, nil
}

// GetSession ...
func (s *Server) GetSession(ctx context.Context, in *pb.GetSessionRequest) (*pb.GetSessionResponse, error) {
	fmt.Println("Request:", in)
	var result *pb.GetSessionResponse
	dbString := fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=%s", os.Getenv("DEV_HOST"), os.Getenv("DEV_DBNAME"), os.Getenv("DEV_USERNAME"), os.Getenv("DEV_PASSWORD"), os.Getenv("DEV_SSLMODE"))
	db, err := gorm.Open(os.Getenv("DEV_DB_DRIVER"), dbString)
	defer db.Close()
	if err != nil {
		log.Errorf("Error opening DB connection: %+v", err)
	} else {
		models.SetupDatabase(db)
		result, _ = s.User.CreateSession(ctx, in, db)
	}
	fmt.Println("Result", result)
	return result, nil
}
