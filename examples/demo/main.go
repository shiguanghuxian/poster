package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/shiguanghuxian/poster/proto"
	grpc "google.golang.org/grpc"
)

var (
	client proto.PosterClient
)

func init() {
	conn, err := grpc.Dial("127.0.0.1:20280", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	client = proto.NewPosterClient(conn)
}

func main() {
	// 系统日志显示文件和行号
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	http.HandleFunc("/love", show)
	log.Println("Starting server ...")
	log.Fatal(http.ListenAndServe(":1314", nil))
}

// 输出海报
func show(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name := r.FormValue("name")
	names := []rune(name)
	showName := "微 微"
	if len(names) == 2 {
		showName = string(names[0]) + " " + string(names[1])
	} else if len(names) > 2 {
		names1 := ""
		for _, v := range names {
			names1 += string(v) + " "
		}
		showName = strings.TrimSpace(names1)
	}

	/* 调用rpc生成图片 */
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// 读取背景图片
	backgroundImg, err := ioutil.ReadFile("./background.jpg")
	if err != nil {
		log.Println(err)
		return
	}
	// 文本
	texts := make([]*proto.Text, 0)
	texts = append(texts, &proto.Text{
		Top:        176,
		Left:       166,
		Width:      600,
		Height:     56,
		LineCount:  20,
		Content:    "我 爱 你 - " + showName,
		FontSize:   50,
		LineHeight: 1.5,
		FontColor:  "#FF3333",
	})
	// 生成二维码
	qrCodes := make([]*proto.QrCode, 0)
	qrCodes = append(qrCodes, &proto.QrCode{
		Top:     940,
		Left:    23,
		Width:   170,
		Content: fmt.Sprintf("http://140.143.234.132:1314/love?name=" + name),
	})
	// 调用
	rsp, err := client.CreatePoster(ctx, &proto.CreatePosterRequest{
		Width:  720,
		Height: 1280,
		Background: &proto.Background{
			Image: backgroundImg,
		},
		Texts:     texts,
		SubQrCode: qrCodes,
	})
	if err != nil {
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(rsp.GetImage())
}
