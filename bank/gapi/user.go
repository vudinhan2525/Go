package gapi

import (
	"context"
	"database/sql"
	db "main/db/sqlc"
	"main/pb"
	"main/util"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserReq) (*pb.CreateUserRes, error) {
	password, err := util.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "password hash failed")
	}
	user, err := server.Store.CreateUser(ctx, db.CreateUserParams{
		Email:          req.GetEmail(),
		FullName:       req.GetFullname(),
		HashedPassword: password,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create user failed %v", err)
	}

	res := &pb.CreateUserRes{
		Status: "Create user successfully",
		User:   ConvertUser(user),
	}
	return res, nil
}
func (server *Server) LoginUser(ctx context.Context, req *pb.LoginUserReq) (*pb.LoginUserRes, error) {

	user, err := server.Store.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found %v", err)
		}
		return nil, status.Errorf(codes.Internal, "error when getting user %v", err)
	}

	err = util.VerifyPassword(req.Password, user.HashedPassword)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "password isn't correct %v", err)
	}
	accessToken, _, err := server.TokenMaker.CreateToken(strconv.FormatInt(user.UserID, 10), user.Email, server.Config.TokenDuration)
	if err != nil {

		return nil, status.Errorf(codes.Internal, "error when creating access token %v", err)
	}
	refreshToken, refreshPayload, err := server.TokenMaker.CreateToken(strconv.FormatInt(user.UserID, 10), user.Email, server.Config.RefreshTokenDuration)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "error when creating refresh token %v", err)
	}
	_, err = server.Store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		UserID:       int64(refreshPayload.UserID),
		RefreshToken: refreshToken,
		UserAgent:    "",
		ClientIp:     "",
		IsBlocked:    false,
		ExpiredAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error when creating session %v", err)
	}
	res := &pb.LoginUserRes{
		Status:       "Login successfully",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Data:         ConvertUser(user),
	}
	return res, nil
}
