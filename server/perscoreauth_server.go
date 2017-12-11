package server

import (
	"fmt"

	"golang.org/x/net/context"

	"perScoreAuth/models"
	pb "perScoreAuth/perScoreProto/user"

	"github.com/jinzhu/gorm"
)

type Server struct {
	User models.User
}

func (s *Server) CreateUser(ctx context.Context, in *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	fmt.Println("Request:", in)
	db, _ := gorm.Open("postgres", "host=localhost port=5432 user=perscoreauth dbname=per_score_auth sslmode=disable password=perscoreauth-dm")
	defer db.Close()
	models.SetupDatabase(db)
	result, _ := s.User.CreateInDB(ctx, in, db)
	fmt.Println("Result", result)
	// if err != nil {
	// 	log.Errorf("Error in CreateInDB: %+v", err)
	// 	return nil, status.Errorf(codes.Internal, err.Error())
	// }
	fmt.Println("Result:", result)
	return result, nil
}

func (s *Server) GetSession(ctx context.Context, in *pb.GetSessionRequest) (*pb.GetSessionResponse, error) {
	fmt.Println("Request:", in)
	db, _ := gorm.Open("postgres", "host=localhost port=5432 user=perscoreauth dbname=per_score_auth sslmode=disable password=perscoreauth-dm")
	defer db.Close()
	result, _ := s.User.CreateSession(ctx, in, db)
	// if err != nil {
	// 	log.Errorf("Error in CreateInDB: %+v", err)
	// 	return nil, status.Errorf(codes.Internal, err.Error())
	// }
	fmt.Println("Result:", result)
	return result, nil
}
