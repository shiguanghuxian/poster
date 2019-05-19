package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"code.google.com/p/graphics-go/graphics"
	"github.com/anthonynsimon/bild/transform"
	"github.com/fogleman/gg"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/nfnt/resize"
	"github.com/shiguanghuxian/poster/program/common"
	"github.com/shiguanghuxian/poster/program/logger"
	qrcode "github.com/skip2/go-qrcode"
)

// Service 具体生成海报业务代码
type Service struct {
	Param *PosterParam // 绘图参数
	rgba  *image.RGBA  // 绘制图片对象
}

// NewService 创建绘图对象 - 检查参数
func NewService(param *PosterParam) (s *Service, err error) {
	if param == nil {
		err = errors.New("The parameter cannot be nil")
		return
	}
	// 主画布尺寸
	if param.Width == 0 {
		param.Width = DefaultWidth
	}
	if param.Height == 0 {
		param.Height = DefaultHeight
	}
	// 背景
	if param.Background == nil {
		err = errors.New("The background cannot be nil")
		return
	}
	if len(param.Background.Image) == 0 && param.Background.ImageURL == "" {
		err = errors.New("The background image url address and background image base64 value cannot be empty")
		return
	}
	if param.Background.ImageType == "" {
		param.Background.ImageType = "jpg"
	}
	// 文本
	if len(param.Texts) > 0 {
		for _, txt := range param.Texts {
			if txt.Content == "" {
				err = errors.New("An empty string exists for the text to be written")
				return
			}
			if txt.LineCount == 0 {
				txt.LineCount = 1
			}
			if txt.FontColor == "" {
				txt.FontColor = "#000000"
			}
			if txt.FontSize == 0 {
				txt.FontSize = 24.0
			}
			if txt.LineHeight == 0 {
				txt.LineHeight = 1.5
			}
			if txt.FontName == "" {
				txt.FontName = "default.ttc"
			}
		}
	}
	// 子图片
	if len(param.SubImages) > 0 {
		for _, subImage := range param.SubImages {
			if len(subImage.Image) == 0 && subImage.ImageURL == "" {
				err = errors.New("SubImage exists image url and image base64 are both empty")
				return
			}
		}
	}

	// 子二维码
	if len(param.SubQrCode) > 0 {
		for _, subQrCode := range param.SubQrCode {
			if subQrCode.Content == "" {
				err = errors.New("QRcode content cannot be empty")
				return
			}
			if subQrCode.BackgroundColor == "" {
				subQrCode.BackgroundColor = "#FFFFFF"
			}
			if subQrCode.ForegroundColor == "" {
				subQrCode.ForegroundColor = "#000000"
			}
			if subQrCode.Width == 0 {
				subQrCode.Width = 100
			}
		}
	}

	// 子小程序码
	if len(param.SubWxQrCode) > 0 {
		for _, subWxQrCode := range param.SubWxQrCode {
			if subWxQrCode.AccessToken == "" {
				return nil, errors.New("小程序码生成，AccessToken参数不能为空")
			}
			if subWxQrCode.Width == 0 {
				subWxQrCode.Width = 100
			}
			if subWxQrCode.LineColor == "" {
				subWxQrCode.LineColor = "#000000"
			}
		}
	}

	s = &Service{
		Param: param,
	}
	return
}

// DrawPoster 生成海报
func (s *Service) DrawPoster() (img []byte, err error) {
	startTime := time.Now()
	defer func() {
		logger.Log.Infow("生成图片耗时", "time", fmt.Sprintf("%dms", time.Now().Sub(startTime).Nanoseconds()/1000000))
	}()

	/* 生成画布 */
	s.rgba = image.NewRGBA(image.Rect(0, 0, s.Param.Width, s.Param.Height))
	// 背景图片拉伸至画布大小
	backgroundImgReader, err := s.getBackgroundImg(s.Param.Background.Image, s.Param.Background.ImageURL)
	if err != nil {
		logger.Log.Errorw("获取背景图错误1", "err", err)
		return
	}
	var backgroundImg image.Image
	if s.Param.Background.ImageType == "png" {
		backgroundImg, err = png.Decode(backgroundImgReader)
	} else if s.Param.Background.ImageType == "jpg" || s.Param.Background.ImageType == "jpeg" {
		backgroundImg, err = jpeg.Decode(backgroundImgReader)
	} else {
		logger.Log.Warnw("背景图片，不支持的图片格式类型，格式必须是png或jpg，格式不带点", "ImageType", s.Param.Background.ImageType)
		err = errors.New("Unsupported image types -- " + s.Param.Background.ImageType)
		return
	}
	if err != nil {
		logger.Log.Errorw("获取背景图错误2", "err", err)
		return
	}
	// 缩放到画布大小
	picResized := resize.Resize(uint(s.Param.Width), uint(s.Param.Height), backgroundImg, resize.Lanczos3)
	// 拉伸至中心完全显示
	draw.Draw(s.rgba, image.Rect(0, 0, s.Param.Width, s.Param.Height), picResized,
		image.Point{int((picResized.Bounds().Dx() - s.Param.Width) / 2), int((picResized.Bounds().Dy() - s.Param.Height) / 2)},
		draw.Src)

	/* 添加子图片 */
	err = s.drawSubImages()
	if err != nil {
		return nil, err
	}

	/* 添加二维码 */
	err = s.drawSubQrCodes()
	if err != nil {
		return nil, err
	}

	/* 添加小程序码 */
	err = s.drawSubWxQrCodes()
	if err != nil {
		return nil, err
	}

	/* 添加文本 */
	err = s.drawSubTexts()
	if err != nil {
		return nil, err
	}

	// 输出图片到字节
	outImg := s.rgba.SubImage(s.rgba.Bounds())
	f := bytes.NewBuffer(make([]byte, 0))
	err = jpeg.Encode(f, outImg, nil)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(f)
}

// 绘制子图片
func (s *Service) drawSubImages() (err error) {
	for subKey, subImg := range s.Param.SubImages {
		sub, err := s.getBackgroundImg(subImg.Image, subImg.ImageURL)
		if err != nil {
			logger.Log.Errorw("解析子图错误1", "err", err, "subKey", subKey)
			return err
		}
		var subImage image.Image
		imageType := strings.ToLower(subImg.ImageType)
		if imageType == "png" {
			subImage, err = png.Decode(sub)
		} else if imageType == "jpg" || imageType == "jpeg" {
			subImage, err = jpeg.Decode(sub)
		} else {
			logger.Log.Warnw("不支持的图片格式类型，格式必须是png或jpg，格式不带点", "err", err, "subKey", subKey)
			continue
		}
		if err != nil {
			logger.Log.Errorw("解析子图错误2", "err", err, "subKey", subKey)
			return err
		}
		// 图片缩放
		subImage = resize.Resize(uint(subImg.Width-subImg.Padding), uint(subImg.Height-subImg.Padding), subImage, resize.Lanczos3)

		// 旋转
		if subImg.Angle != 0 {
			dst := image.NewCMYK(image.Rect(0, 0, subImg.Width, subImg.Height))
			if subImg.Color != "" {
				subBColor, err := common.HexToColor(subImg.Color)
				if err != nil {
					logger.Log.Errorw("子图背景色解析错误", "err", err, "subKey", subKey, "subBColor", subBColor)
					return err
				}
				for x := 0; x < dst.Bounds().Dx(); x++ {
					for y := 0; y < dst.Bounds().Dy(); y++ {
						dst.Set(x, y, subBColor)
					}
				}
			}
			err = graphics.Rotate(dst, subImage, &graphics.RotateOptions{gg.Radians(subImg.Angle)})
			if err != nil {
				logger.Log.Errorw("图片旋转错误", "err", err, "subKey", subKey, "method", "drawSubQrCodes")
				return err
			}
			subImage = dst
		}
		imgResized := transform.Resize(subImage, subImg.Width, subImg.Height, transform.Linear)
		draw.Draw(s.rgba,
			image.Rectangle{
				image.Point{subImg.Left + subImg.Width/2, subImg.Top + subImg.Height/2},
				s.rgba.Bounds().Max,
			},
			imgResized,
			image.Point{0, 0},
			draw.Src)
	}
	return
}

// 绘制文本
func (s *Service) drawSubTexts() (err error) {
	for _, txt := range s.Param.Texts {
		// 换行后的文本
		texts := txt.GetTest()
		// log.Println(texts)
		// 字体
		font, err := s.getFont(txt.FontName)
		if err != nil {
			logger.Log.Errorw("获取字体错误", "err", err)
			return err
		}
		c := freetype.NewContext()
		c.SetFont(font)
		c.SetFontSize(txt.FontSize)
		c.SetClip(image.Rect(txt.Left, txt.Top, txt.Left+txt.Width, txt.Top+txt.Height)) //文字区域
		c.SetDst(s.rgba)
		// 字体颜色
		fontColor, err := common.HexToColor(txt.FontColor)
		if err != nil {
			logger.Log.Errorw("解析字体颜色错误", "err", err, "FontColor", txt.FontColor)
			return err
		}
		c.SetSrc(image.NewUniform(fontColor))

		pt := freetype.Pt(txt.Left, txt.Top+int(c.PointToFixed(txt.FontSize)>>6))
		for _, t := range texts {
			c.DrawString(t, pt)
			pt.Y += c.PointToFixed(txt.FontSize * txt.LineHeight)
		}
	}
	return
}

// 绘制二维码
func (s *Service) drawSubQrCodes() (err error) {
	for k, v := range s.Param.SubQrCode {
		// 生成二维码
		qr, err := qrcode.New(v.Content, qrcode.High)
		if err != nil {
			logger.Log.Errorw("生成二维码错误", "err", err)
			return err
		}
		backgroundColor, err := common.HexToColor(v.BackgroundColor)
		if err != nil {
			logger.Log.Errorw("解析二维码背景色错误", "err", err)
			return err
		}
		foregroundColor, err := common.HexToColor(v.ForegroundColor)
		if err != nil {
			logger.Log.Errorw("解析二维码前景色错误", "err", err)
			return err
		}
		qr.BackgroundColor = backgroundColor
		qr.ForegroundColor = foregroundColor
		qrImg := qr.Image(v.Width)
		// 旋转图片
		if v.Angle != 0 {
			dst := image.NewCMYK(image.Rect(0, 0, v.Width, v.Width))
			err = graphics.Rotate(dst, qrImg, &graphics.RotateOptions{gg.Radians(v.Angle)})
			if err != nil {
				logger.Log.Errorw("图片旋转错误", "err", err, "subKey", k, "method", "drawSubQrCodes")
				return err
			}
			qrImg = dst
		}
		// 绘入主图
		draw.Draw(s.rgba,
			image.Rectangle{
				image.Point{v.Left + v.Width/2, v.Top + v.Width/2},
				s.rgba.Bounds().Max,
			},
			qrImg,
			image.Point{0, 0},
			draw.Src)
	}
	return
}

// 绘制小程序码
func (s *Service) drawSubWxQrCodes() (err error) {
	for k, v := range s.Param.SubWxQrCode {
		lineColor, err := common.HexToColor(v.LineColor)
		if err != nil {
			logger.Log.Errorw("解析小程序码颜色错误", "err", err)
			return err
		}
		lineColorRGB := lineColor.(color.RGBA)
		req := map[string]interface{}{
			"scene":      v.Scene,
			"page":       v.Page,
			"width":      v.Width,
			"auto_color": v.AutoColor,
			"line_color": map[string]int{
				"r": int(lineColorRGB.R),
				"g": int(lineColorRGB.G),
				"b": int(lineColorRGB.B),
			},
			"is_hyaline": v.IsHyaline,
		}
		reqRed, _ := json.Marshal(req)
		wxQrCodeUrl := "https://api.weixin.qq.com/wxa/getwxacodeunlimit?access_token=" + v.AccessToken
		// 请求接口生成小程序码
		resp, err := http.Post(wxQrCodeUrl, "application/json", bytes.NewReader(reqRed))
		if err != nil {
			respBody, _ := ioutil.ReadAll(resp.Body)
			logger.Log.Errorw("请求获取小程序码错误", "err", err, "body", string(respBody))
			return err
		}
		qrImg, err := jpeg.Decode(resp.Body)
		if err != nil {
			logger.Log.Errorw("小程序码返回body解析错误", "err", err)
			return err
		}
		// 图片缩放
		qrImg = resize.Resize(uint(v.Width), uint(v.Width), qrImg, resize.Lanczos3)

		// 旋转图片
		if v.Angle != 0 {
			dst := image.NewCMYK(image.Rect(0, 0, v.Width, v.Width))
			err = graphics.Rotate(dst, qrImg, &graphics.RotateOptions{gg.Radians(v.Angle)})
			if err != nil {
				logger.Log.Errorw("图片旋转错误", "err", err, "subKey", k, "method", "drawSubWxQrCodes")
				return err
			}
			qrImg = dst
		}
		// 绘入主图
		draw.Draw(s.rgba,
			image.Rectangle{
				image.Point{v.Left + qrImg.Bounds().Dx()/2, v.Top + qrImg.Bounds().Dy()/2},
				s.rgba.Bounds().Max,
			},
			qrImg,
			image.Point{0, 0},
			draw.Src)
	}
	return
}

// getBackgroundImg 获取背景图
func (s *Service) getBackgroundImg(img []byte, imgUrl string) (r io.Reader, err error) {
	if len(img) != 0 {
		r = bytes.NewReader(img)
	} else {
		resp, err := http.Get(imgUrl)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		fileBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		r = bytes.NewReader(fileBytes)
	}
	return
}

var (
	allFonts = new(sync.Map)
)

// 获取字体
func (s *Service) getFont(name string) (font *truetype.Font, err error) {
	if val, ok := allFonts.Load(name); ok == true {
		font, ok = val.(*truetype.Font)
		if ok == true {
			return
		}
	}
	fontFile, err := os.Open(fmt.Sprintf("./resources/fonts/%s", name))
	if err != nil {
		return
	}
	fontBytes, err := ioutil.ReadAll(fontFile)
	if err != nil {
		return
	}
	font, err = freetype.ParseFont(fontBytes)
	if err != nil {
		return
	}
	// 存储字体
	allFonts.Store(name, font)
	return
}
