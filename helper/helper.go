package helper

import (
	"context"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"sxp-server/model"
)

var SECRETKEY = []byte("sxp-server") //私钥

// ParseToken 解析JWT
func ParseToken(tokenString string) (*model.MyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &model.MyClaims{}, func(token *jwt.Token) (i interface{}, err error) {
		return SECRETKEY, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*model.MyClaims); ok && token.Valid { // 校验token
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

// CheckToken
//
//	@Description: grpc服务端的token校验
//	@param md
//	@return err
//	@return flag
func CheckToken(md metadata.MD) (err error, flag bool) {
	t, exist := md["token"]
	if !exist || t[0] == "" {
		err = errors.New("metadata里面没有token")
		return
	}
	mc, err := ParseToken(t[0])
	if mc == nil || err != nil {
		err = status.Errorf(codes.DataLoss, "grpc服务端解析token失败")
		return
	}
	flag = true
	return
}

// HeadResponse
//
//	@Description: header返回数据
//	@param ctx
//	@param checkRes
//	@return err
func HeadResponse(ctx context.Context, checkRes string) (err error) {
	// 创建和发送header.
	header := metadata.New(map[string]string{"check_token": checkRes})
	err = grpc.SendHeader(ctx, header)
	return
}

func TrailerResponse(ctx context.Context) (err error) {
	trailer := metadata.Pairs("sign", "sxp-alan")
	err = grpc.SetTrailer(ctx, trailer)
	if err != nil {
		return
	}
	return
}
