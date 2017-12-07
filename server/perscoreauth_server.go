package server

import (
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"perScoreAuth/models"
	pb "perScoreAuth/perScoreProto/user"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	User models.User
}

func (s *Server) CreateUser(ctx context.Context, in *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	fmt.Println("Request:", in)
	db, err := gorm.Open("postgres", "host=localhost port=5432 user=perscoreauth dbname=per_score_auth sslmode=disable password=perscoreauth-dm")
	defer db.Close()
	models.SetupDatabase(db)
	result, err := s.User.CreateInDB(ctx, in, db)
	if err != nil {
		log.Errorf("Error in CreateInDB: %+v", err)
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return result, nil
}

func (s *Server) GetSession(ctx context.Context, in *pb.GetSessionRequest) (*pb.GetSessionResponse, error) {
	db, err := gorm.Open("postgres", "host=localhost port=5432 user=perscoreauth dbname=per_score_auth sslmode=disable password=perscoreauth-dm")
	defer db.Close()
	result, err := s.User.CreateSession(ctx, in, db)
	if err != nil {
		log.Errorf("Error in CreateInDB: %+v", err)
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return result, nil
}
