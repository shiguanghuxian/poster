package service

import (
	"bytes"
	"errors"
	"fmt"
	"image"
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
)

// Service 具体生成海报业务代码
type Service struct {
	Param *PosterParam // 绘图参数
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
	s = &Service{
		Param: param,
	}

	return
}

// DrawPoster 生成海报
func (s *Service) DrawPoster() (img []byte, err error) {
	startTime := time.Now()
	defer func() {
		logger.Log.Infow("生成图片耗时", "time", fmt.Sprintf("%d ms", time.Now().Sub(startTime).Nanoseconds()/100000))
	}()

	/* 生成画布 */
	rgba := image.NewRGBA(image.Rect(0, 0, s.Param.Width, s.Param.Height))
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
	draw.Draw(rgba, image.Rect(0, 0, s.Param.Width, s.Param.Height), picResized,
		image.Point{int((picResized.Bounds().Dx() - s.Param.Width) / 2), int((picResized.Bounds().Dy() - s.Param.Height) / 2)},
		draw.Src)

	/* 插入子图片 */
	for subKey, subImg := range s.Param.SubImages {
		sub, err := s.getBackgroundImg(subImg.Image, subImg.ImageURL)
		if err != nil {
			logger.Log.Errorw("解析子图错误1", "err", err, "subKey", subKey)
			return nil, err
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
			return nil, err
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
					return nil, err
				}
				for x := 0; x < dst.Bounds().Dx(); x++ {
					for y := 0; y < dst.Bounds().Dy(); y++ {
						dst.Set(x, y, subBColor)
					}
				}
			}
			err = graphics.Rotate(dst, subImage, &graphics.RotateOptions{gg.Radians(subImg.Angle)})
			if err != nil {
				logger.Log.Errorw("图片旋转错误", "err", err, "subKey", subKey)
				return nil, err
			}
			subImage = dst
		}
		qrcodeResized := transform.Resize(subImage, subImg.Width, subImg.Height, transform.Linear)
		draw.Draw(rgba,
			image.Rectangle{
				image.Point{subImg.Left + subImg.Width/2, subImg.Top + subImg.Height/2},
				rgba.Bounds().Max,
			},
			qrcodeResized,
			image.Point{0, 0},
			draw.Src)
	}

	/* 添加文本 */
	for _, txt := range s.Param.Texts {
		// 换行后的文本
		texts := txt.GetTest()
		// log.Println(texts)
		// 字体
		font, err := s.getFont(txt.FontName)
		if err != nil {
			logger.Log.Errorw("获取字体错误", "err", err)
			return nil, err
		}
		c := freetype.NewContext()
		c.SetFont(font)
		c.SetFontSize(txt.FontSize)
		c.SetClip(image.Rect(txt.Left, txt.Top, txt.Left+txt.Width, txt.Top+txt.Height)) //文字区域
		c.SetDst(rgba)
		// 字体颜色
		fontColor, err := common.HexToColor(txt.FontColor)
		if err != nil {
			logger.Log.Errorw("解析字体颜色错误", "err", err, "FontColor", txt.FontColor)
			return nil, err
		}
		c.SetSrc(image.NewUniform(fontColor))

		pt := freetype.Pt(txt.Left, txt.Top+int(c.PointToFixed(txt.FontSize)>>6))
		for _, t := range texts {
			c.DrawString(t, pt)
			pt.Y += c.PointToFixed(txt.FontSize * txt.LineHeight)
		}
	}

	// 输出图片到字节
	outImg := rgba.SubImage(rgba.Bounds())
	f := bytes.NewBuffer(make([]byte, 0))
	err = jpeg.Encode(f, outImg, nil)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(f)
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
