package program

import (
	"net/http"
	"os/exec"
	"runtime"

	"github.com/shiguanghuxian/poster/program/config"
	"github.com/shiguanghuxian/poster/program/logger"
	"github.com/shiguanghuxian/poster/program/transport"
)

// Program 主程序
type Program struct {
	cfg *config.Config
	s   *http.Server
}

// New 创建主程序
func New() (*Program, error) {
	// 配置文件
	cfg, err := config.LoadConfig("")
	if err != nil {
		return nil, err
	}

	// 日志对象
	_, err = logger.InitLogger(cfg.LogPath, cfg.Debug)
	if err != nil {
		return nil, err
	}

	// jj, _ := json.Marshal(cfg)
	// fmt.Println(string(jj))

	return &Program{
		cfg: cfg,
	}, nil
}

// Run 启动程序
func (p *Program) Run() error {
	// 启动http服务
	if p.cfg.HTTP.Enable == true {
		logger.Log.Infow("http服务启动", "address", p.cfg.HTTP.Address, "port", p.cfg.HTTP.Port)
		go func() {
			err := transport.NewHTTPTransport(p.cfg.HTTP).ListenAndServe(p.cfg.Debug)
			if err != nil {
				panic(err)
			}
		}()
	}

	// 启动grpc服务
	if p.cfg.GRPC.Enable == true {
		logger.Log.Infow("grpc服务启动", "address", p.cfg.GRPC.Address, "port", p.cfg.GRPC.Port)
		go func() {
			err := transport.NewGRPCTransport(p.cfg.GRPC).ListenAndServe(p.cfg.Debug)
			if err != nil {
				panic(err)
			}
		}()
	}

	return nil
}

// Stop 停止服务
func (p *Program) Stop() {
	if p.s != nil {
		p.s.Close()
	}
}

// 打开url
func openURL(urlAddr string) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", " /c start "+urlAddr)
	} else if runtime.GOOS == "darwin" {
		cmd = exec.Command("open", urlAddr)
	} else {
		return
	}
	err := cmd.Start()
	if err != nil {
		logger.Log.Errorw("Error opening browser", "err", err)
	}
}
