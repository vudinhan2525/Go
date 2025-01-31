package gapi

import (
	db "main/db/sqlc"
	"main/pb"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func ConvertUser(user db.User) *pb.User {
	return &pb.User{
		UserId:            user.UserID,
		FullName:          user.FullName,
		Email:             user.Email,
		Role:              string(user.Role),
		PasswordChangedAt: timestamppb.New(user.PasswordChangedAt),
		CreatedAt:         timestamppb.New(user.CreatedAt),
	}
}
