package gapi

import (
	"context"
	"database/sql"
	db "main/db/sqlc"
	"main/pb"
	"main/pkg/val"
	"main/util"
	"main/worker"
	"strconv"
	"time"

	"github.com/hibiken/asynq"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func validateCreateUserRequest(req *pb.CreateUserReq) (violations []*errdetails.BadRequest_FieldViolation) {

	if err := val.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}

	if err := val.ValidateFullName(req.GetFullname()); err != nil {
		violations = append(violations, fieldViolation("fullname", err))
	}

	if err := val.ValidateEmail(req.GetEmail()); err != nil {
		violations = append(violations, fieldViolation("email", err))
	}

	return violations
}

func validateUpdateUserRequest(req *pb.UpdateUserReq) (violations []*errdetails.BadRequest_FieldViolation) {

	if req.GetPassword() != "" {
		if err := val.ValidatePassword(req.GetPassword()); err != nil {
			violations = append(violations, fieldViolation("password", err))
		}
	}
	if req.GetFullname() != "" {
		if err := val.ValidateFullName(req.GetFullname()); err != nil {
			violations = append(violations, fieldViolation("fullname", err))
		}
	}
	if req.GetEmail() != "" {
		if err := val.ValidateEmail(req.GetEmail()); err != nil {
			violations = append(violations, fieldViolation("email", err))
		}
	}
	return violations
}
func validateLoginUserRequest(req *pb.LoginUserReq) (violations []*errdetails.BadRequest_FieldViolation) {

	if err := val.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}
	if err := val.ValidateEmail(req.GetEmail()); err != nil {
		violations = append(violations, fieldViolation("email", err))
	}

	return violations
}
func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserReq) (*pb.CreateUserRes, error) {
	violations := validateCreateUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}
	password, err := util.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "password hash failed")
	}

	params := db.CreateUserParams{
		Email:          req.GetEmail(),
		FullName:       req.GetFullname(),
		HashedPassword: password,
		Role:           db.UserRoleUser,
	}
	resultTx, err := server.Store.CreateUserTx(ctx, db.CreateUserTxParams{
		CreateUserParams: params,
		AfterCreate: func(user db.User) error {
			opts := []asynq.Option{
				asynq.MaxRetry(10),
				asynq.ProcessIn(10 * time.Second),
			}
			err = server.TaskDistributor.DistributeTaskSendVerifyEmail(ctx, &worker.PayloadSendVerifyEmail{
				UserID: user.UserID,
			}, opts...)
			if err != nil {
				return err
			}
			return nil
		},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create user failed %s", err)
	}

	res := &pb.CreateUserRes{
		Status: "Create user successfully",
		User:   ConvertUser(resultTx.User),
	}
	return res, nil
}
func (server *Server) LoginUser(ctx context.Context, req *pb.LoginUserReq) (*pb.LoginUserRes, error) {
	violations := validateLoginUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}
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
	accessToken, _, err := server.TokenMaker.CreateToken(strconv.FormatInt(user.UserID, 10), user.Email, user.Role, server.Config.TokenDuration)
	if err != nil {

		return nil, status.Errorf(codes.Internal, "error when creating access token %v", err)
	}
	refreshToken, refreshPayload, err := server.TokenMaker.CreateToken(strconv.FormatInt(user.UserID, 10), user.Email, user.Role, server.Config.RefreshTokenDuration)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "error when creating refresh token %v", err)
	}

	mtdt := util.ExtractMetadata(ctx)
	_, err = server.Store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		UserID:       int64(refreshPayload.UserID),
		RefreshToken: refreshToken,
		UserAgent:    mtdt.UserAgent,
		ClientIp:     mtdt.ClientIp,
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

func (server *Server) UpdateMe(ctx context.Context, req *pb.UpdateUserReq) (*pb.UpdateUserRes, error) {
	violations := validateUpdateUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}
	payload, err := GetAuthPayload(ctx)
	if err != nil {
		return nil, err
	}
	params := db.UpdateUserParams{
		Email:    sql.NullString{String: req.GetEmail(), Valid: req.GetEmail() != ""},
		FullName: sql.NullString{String: req.GetFullname(), Valid: req.GetFullname() != ""},
		UserID:   int64(payload.UserID),
	}
	if req.GetPassword() != "" {
		hashedPassword, err := util.HashPassword(req.GetPassword())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "password hash failed")
		}
		params.HashedPassword = sql.NullString{String: hashedPassword, Valid: hashedPassword != ""}
		params.PasswordChangedAt = sql.NullTime{Time: time.Now(), Valid: hashedPassword != ""}
	}

	user, err := server.Store.UpdateUser(ctx, params)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "update user failed %v", err)
	}

	res := &pb.UpdateUserRes{
		Status: "Update user successfully",
		Data:   ConvertUser(user),
	}
	return res, nil
}
