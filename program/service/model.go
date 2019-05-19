package service

import "unicode"

// 默认值
const (
	DefaultWidth       = 720  // 画布宽度
	DefaultHeight      = 1280 // 画布高度
	DefaultBorderWidth = 6    // 边框线条宽度
)

// PosterParam 生成海报参数
type PosterParam struct {
	Width       int         `json:"width,omitempty"`          // 画布宽度
	Height      int         `json:"height,omitempty"`         // 画布高度
	Background  *Background `json:"background,omitempty"`     // 背景图片
	Texts       []*Text     `json:"texts,omitempty"`          // 文本列表
	SubImages   []*Image    `json:"sub_images,omitempty"`     // 需要插入的子图片列表
	SubQrCode   []*QrCode   `json:"sub_qr_code,omitempty"`    // 需要每次都动态生成的二维码信息
	SubWxQrCode []*WxQrCode `json:"sub_wx_qr_code,omitempty"` // 微信小程序码
}

// Background 背景 - Image和ImageUrl至少传一个
type Background struct {
	Image     []byte `json:"image,omitempty"`      // 图片base64值
	ImageURL  string `json:"image_url,omitempty"`  // 背景图片地址
	ImageType string `json:"image_type,omitempty"` // 图片格式类型 jpg | png
}

// SubObject 子对象位置和大小
type SubObject struct {
	Top    int `json:"top,omitempty"`    // 距离顶部距离
	Left   int `json:"left,omitempty"`   // 距离左侧距离
	Width  int `json:"width,omitempty"`  // 文本区域宽度 - 当二维码和小程序码时只有宽度生效
	Height int `json:"height,omitempty"` // 文本区域高度
}

// Text 海报文字
type Text struct {
	SubObject
	LineCount  int     `json:"line_count,omitempty"`  // 每行字符数 - 长度 汉字=2 字母=1
	Content    string  `json:"content,omitempty"`     // 文字内容
	FontName   string  `json:"font_name,omitempty"`   // 字体名 - 需要先将字体问题放到资源目录
	FontSize   float64 `json:"font_size,omitempty"`   // 字体大小
	LineHeight float64 `json:"line_height,omitempty"` // 行间距
	FontColor  string  `json:"font_color,omitempty"`  // 字体颜色
}

// GetTest 获取换行后的文本 - 会根据每行字符数换行
func (txt *Text) GetTest() (texts []string) {
	line := 0
	count := 0
	for _, v := range txt.Content {
		// 汉字+2，字母+1
		if unicode.Is(unicode.Han, v) {
			count += 2
		} else {
			count++
		}
		if count > txt.LineCount {
			line++
			count = 0
		}
		if len(texts) <= line {
			texts = append(texts, string(v))
		} else {
			texts[line] = texts[line] + string(v)
		}
	}
	return
}

// Image 海报贴图 - Image和ImageUrl至少传一个
type Image struct {
	SubObject
	Padding   int     `json:"padding,omitempty"`    // 内边距 - 当图片旋转时有用
	Angle     float64 `json:"angle,omitempty"`      // 旋转角度 - 顺时针方向 - 弧度
	Color     string  `json:"color,omitempty"`      // 背景色
	ImageType string  `json:"image_type,omitempty"` // 图片格式类型 jpg | png
	Image     []byte  `json:"image,omitempty"`      // 图片base64值
	ImageURL  string  `json:"image_url,omitempty"`  // 背景图片地址
}

// QrCode 子二维码，根据内容生成 - 非图片
type QrCode struct {
	SubObject
	Angle           float64 `json:"angle,omitempty"`            // 旋转角度 - 顺时针方向 - 弧度
	BackgroundColor string  `json:"background_color,omitempty"` // 背景色 - 可为空 - 默认白色
	ForegroundColor string  `json:"foreground_color,omitempty"` // 前景色 - 可为空 - 默认黑色
	Content         string  `json:"content,omitempty"`          // 二维码内容
}

// WxQrCode 小程序码自动生成
type WxQrCode struct {
	SubObject
	Angle float64 `json:"angle,omitempty"` // 旋转角度 - 顺时针方向 - 弧度
	// 一下参数直接传给微信
	AccessToken string `json:"access_token,omitempty"` // 海报生成服务不负责保存access_token，请每次都传递可以access_token
	Scene       string `json:"scene,omitempty"`
	Page        string `json:"page,omitempty"`
	AutoColor   bool   `json:"auto_color,omitempty"`
	LineColor   string `json:"line_color,omitempty"`
	IsHyaline   bool   `json:"is_hyaline,omitempty"`
}
