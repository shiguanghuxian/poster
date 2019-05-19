package transport

import (
	"context"
	"fmt"
	"net"

	"github.com/shiguanghuxian/poster/program/config"
	"github.com/shiguanghuxian/poster/program/service"
	"github.com/shiguanghuxian/poster/proto"
	grpc "google.golang.org/grpc"
)

// GRPCTransport 提供grpc服务生成海报
type GRPCTransport struct {
	cfg *config.GRPCConfig
}

// NewGRPCTransport 创建grpc接口对象
func NewGRPCTransport(cfg *config.GRPCConfig) *GRPCTransport {
	if cfg.Port < 0 {
		cfg.Port = 10280
	}
	return &GRPCTransport{
		cfg: cfg,
	}
}

// ListenAndServe 监听启动服务
func (s *GRPCTransport) ListenAndServe(debug bool) (err error) {
	address := fmt.Sprintf("%s:%d", s.cfg.Address, s.cfg.Port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return
	}
	srv := grpc.NewServer()
	proto.RegisterPosterServer(srv, &PosterServer{})
	srv.Serve(lis)

	return
}

// PosterServer 实现grpc服务接口
type PosterServer struct {
}

// CreatePoster 创建海报
func (ps *PosterServer) CreatePoster(ctx context.Context, req *proto.CreatePosterRequest) (rsp *proto.CreatePosterReply, err error) {
	rsp = new(proto.CreatePosterReply)
	// 海报生成对象
	param := &service.PosterParam{
		Width:  int(req.Width),
		Height: int(req.Height),
	}
	// 背景
	if req.Background != nil {
		param.Background = &service.Background{
			Image:    req.Background.Image,
			ImageURL: req.Background.ImageUrl,
		}
	}
	// 文本
	if len(req.Texts) > 0 {
		param.Texts = make([]*service.Text, 0)
		for _, v := range req.Texts {
			param.Texts = append(param.Texts, &service.Text{
				SubObject: service.SubObject{
					Top:    int(v.Top),
					Left:   int(v.Left),
					Width:  int(v.Width),
					Height: int(v.Height),
				},
				LineCount:  int(v.LineCount),
				Content:    v.Content,
				FontName:   v.FontName,
				FontSize:   v.FontSize,
				LineHeight: v.LineHeight,
				FontColor:  v.FontColor,
			})
		}
	}
	// 子图片
	if len(req.SubImages) > 0 {
		param.SubImages = make([]*service.Image, 0)
		for _, v := range req.SubImages {
			param.SubImages = append(param.SubImages, &service.Image{
				SubObject: service.SubObject{
					Top:    int(v.Top),
					Left:   int(v.Left),
					Width:  int(v.Width),
					Height: int(v.Height),
				},
				Padding:   int(v.Padding),
				Angle:     v.Angle,
				Color:     v.Color,
				ImageType: v.ImageType,
				Image:     v.Image,
				ImageURL:  v.ImageUrl,
			})
		}
	}
	// 二维码
	if len(req.SubQrCode) > 0 {
		for _, v := range req.SubQrCode {
			param.SubQrCode = append(param.SubQrCode, &service.QrCode{
				SubObject: service.SubObject{
					Top:   int(v.Top),
					Left:  int(v.Left),
					Width: int(v.Width),
				},
				Angle:           v.Angle,
				BackgroundColor: v.BackgroundColor,
				ForegroundColor: v.ForegroundColor,
				Content:         v.Content,
			})
		}
	}
	// 小程序码
	if len(req.SubWxQrCode) > 0 {
		for _, v := range req.SubWxQrCode {
			param.SubWxQrCode = append(param.SubWxQrCode, &service.WxQrCode{
				SubObject: service.SubObject{
					Top:   int(v.Top),
					Left:  int(v.Left),
					Width: int(v.Width),
				},
				Angle:       v.Angle,
				AccessToken: v.AccessToken,
				Scene:       v.Scene,
				Page:        v.Page,
				AutoColor:   v.AutoColor,
				LineColor:   v.LineColor,
				IsHyaline:   v.IsHyaline,
			})
		}
	}

	// 生成图片
	srv, err := service.NewService(param)
	if err != nil {
		return
	}
	img, err := srv.DrawPoster()
	if err != nil {
		return
	}
	// 响应图片字节
	rsp.Image = img
	return
}
