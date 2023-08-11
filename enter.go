// Package goimgp 对govips简单封装，提供一些好的方法实现图片压缩和转换
package goimgp

import (
	"errors"
	"github.com/davidbyttow/govips/v2/vips"
	"log"
	"runtime"
)

const (
	ImgTypeJPEG ImgType = 1
	ImgTypePng  ImgType = 2
	ImgTypeGIF  ImgType = 3
	ImgTypeWEBP ImgType = 4
)

type ImgType int

type Encoder struct {
	img    *vips.ImageRef
	Format ImgType
	Data   []byte // 图片的byte流数据
}

var (
	boolFalse   vips.BoolParameter
	intMinusOne vips.IntParameter
)

var (
	ErrorSupportImage = errors.New("不支持的图片格式")
)

// custom logger. Only log error, critical, and warning
func myLogger(messageDomain string, verbosity vips.LogLevel, message string) {
	var messageLevelDescription string
	switch verbosity {
	case vips.LogLevelError:
		messageLevelDescription = "error"
	case vips.LogLevelCritical:
		messageLevelDescription = "critical"
	case vips.LogLevelWarning:
		messageLevelDescription = "warning"
	//case vips.LogLevelMessage:
	//	messageLevelDescription = "message"
	//case vips.LogLevelInfo:
	//	messageLevelDescription = "info"
	//case vips.LogLevelDebug:
	//	messageLevelDescription = "debug"
	default:
		return
	}

	log.Printf("[%v.%v] %v", messageDomain, messageLevelDescription, message)
}

func init() {
	vips.LoggingSettings(myLogger, vips.LogLevelInfo)
	vips.Startup(&vips.Config{
		ConcurrencyLevel: runtime.NumCPU(),
	})
	boolFalse.Set(false)
	intMinusOne.Set(-1)
}

// ShutDown 关闭libvips，调用govips的Shutdown方法，不清楚具体作用
func ShutDown() {
	vips.Shutdown()
}

// LoadImgFromBufferOnePage 从比特流读取图片，只读取第一帧
func LoadImgFromBufferOnePage(buffer []byte) (*Encoder, error) {
	var intNumPages vips.IntParameter
	intNumPages.Set(1)
	encode := new(Encoder)
	img, err := vips.LoadImageFromBuffer(buffer, &vips.ImportParams{
		FailOnError: boolFalse,
		NumPages:    intNumPages,
	})

	if err != nil {
		return nil, err
	}
	switch img.Format() {
	case vips.ImageTypeWEBP:
		encode.Format = ImgTypeWEBP
	case vips.ImageTypeGIF:
		encode.Format = ImgTypeGIF
	case vips.ImageTypeJPEG:
		encode.Format = ImgTypeJPEG
	case vips.ImageTypePNG:
		encode.Format = ImgTypePng
	default:
		return nil, ErrorSupportImage
	}

	encode.img = img
	encode.Data = buffer

	return encode, nil
}

// LoadImgFromBuffer 从比特流读取图片
func LoadImgFromBuffer(buffer []byte) (*Encoder, error) {
	encode := new(Encoder)
	img, err := vips.LoadImageFromBuffer(buffer, &vips.ImportParams{
		FailOnError: boolFalse,
		NumPages:    intMinusOne,
	})

	if err != nil {
		return nil, err
	}
	switch img.Format() {
	case vips.ImageTypeWEBP:
		encode.Format = ImgTypeWEBP
	case vips.ImageTypeGIF:
		encode.Format = ImgTypeGIF
	case vips.ImageTypeJPEG:
		encode.Format = ImgTypeJPEG
	case vips.ImageTypePNG:
		encode.Format = ImgTypePng
	default:
		return nil, ErrorSupportImage
	}

	encode.img = img
	encode.Data = buffer

	return encode, nil
}

// ToPng 转为png格式
// 动图只会截取第一帧
// 该方法不会改变引用对象
func (that *Encoder) ToPng() ([]byte, error) {
	var (
		err       error
		resultBuf []byte
	)

	img := that.img

	if img.Pages() > 1 {
		newThat, err := LoadImgFromBufferOnePage(that.Data)
		if err != nil {
			return nil, err
		}

		img = newThat.img
	}

	resultBuf, _, err = img.ExportPng(&vips.PngExportParams{
		// StripMetadata: false,
		Compression: 9, // 压缩等级 0 - 9
		Filter:      vips.PngFilterNone,
		Interlace:   false, // 交错, 会增大体积，但是浏览器体验好
		Quality:     75,    // 优化程度，仅在palette开启时有效
		Palette:     true,  // 调色板模式, 有效减小体积
		// Dither:      0,
		Bitdepth: 8, // 色深
		// Profile:       "",
	})

	if err != nil {
		return nil, err
	}

	return resultBuf, nil
}

// ToJpeg 转为jpeg格式
// 动图只会截取第一帧
// 该方法不会改变引用对象
func (that *Encoder) ToJpeg() ([]byte, error) {
	var (
		err       error
		resultBuf []byte
	)

	img := that.img

	if img.Pages() > 1 {
		newThat, err := LoadImgFromBufferOnePage(that.Data)
		if err != nil {
			return nil, err
		}

		img = newThat.img
	}

	resultBuf, _, err = img.ExportJpeg(&vips.JpegExportParams{
		// https://www.libvips.org/API/current/VipsForeignSave.html#vips-jpegsave
		StripMetadata:  true, // 从图像中删除所有元数据
		Quality:        75,
		Interlace:      true, // 交错（渐进式）, 浏览器体验好，体积也会变小
		OptimizeCoding: true, // 优化编码，会减小一点体积
		// SubsampleMode:      vips.VipsForeignSubsampleOn, // 色度子采样模式
		TrellisQuant:       true, // 对每个8x8块应用trellis量化。这可以减小文件大小，但会增加压缩时间
		OvershootDeringing: true, // 对具有极端值的样本应用过冲, 减少压缩引起的震荡伪影
		OptimizeScans:      true, // 将DCT系数的频谱拆分为单独扫描。这可以减小文件大小，但会增加压缩时间
		QuantTable:         3,    // 量化表 0 - 8
	})

	if err != nil {
		return nil, err
	}

	return resultBuf, nil
}

// ToGif 转为Gif
// 该方法不会改变引用对象
func (that *Encoder) ToGif() ([]byte, error) {
	var (
		err       error
		resultBuf []byte
	)

	resultBuf, _, err = that.img.ExportGIF(&vips.GifExportParams{
		StripMetadata: true,
		Quality:       75,
		// Dither:        0,
		Effort:   7,
		Bitdepth: 8,
	})

	if err != nil {
		return nil, err
	}

	return resultBuf, nil
}

// ToWebp 转为webp格式
// 该方法不会改变引用对象
func (that *Encoder) ToWebp() ([]byte, error) {
	var (
		err       error
		resultBuf []byte
	)

	resultBuf, _, err = that.img.ExportWebp(&vips.WebpExportParams{
		Quality:         75,
		Lossless:        false,
		StripMetadata:   true,
		ReductionEffort: 4,
	})

	if err != nil {
		return nil, err
	}

	return resultBuf, nil
}

// LossLess 压缩图片尽可能保证质量
// 该方法不会改变引用对象
func (that *Encoder) LossLess() ([]byte, error) {
	switch that.Format {
	case ImgTypeJPEG:
		return that.ToJpeg()
	case ImgTypePng:
		return that.ToPng()
	case ImgTypeGIF:
		return that.ToGif()
	case ImgTypeWEBP:
		return that.ToWebp()
	default:
		return nil, ErrorSupportImage
	}
}

// Compress 压缩并限制长宽，默认质量为65，你也可以自行设置
// 该方法不会改变引用对象
func (that *Encoder) Compress(maxW, maxH int, quality ...int) (buf []byte, err error) {
	q := 65
	img, err := that.img.Copy()
	if err != nil {
		return nil, err
	}
	if len(quality) > 0 {
		q = quality[0]
		if q <= 0 {
			q = 35
		}

		if q > 99 {
			q = 100
		}
	}
	_, err = thumbNail(img, maxW, maxH)
	if err != nil {
		return nil, err
	}
	switch that.Format {
	case ImgTypeJPEG:
		buf, _, err = img.ExportJpeg(&vips.JpegExportParams{
			StripMetadata:  true,
			Quality:        q,
			Interlace:      true,
			OptimizeCoding: true,
			// SubsampleMode:      0,
			TrellisQuant:       true,
			OvershootDeringing: true,
			OptimizeScans:      true,
			QuantTable:         3,
		})
	case ImgTypePng:
		buf, _, err = img.ExportPng(&vips.PngExportParams{
			StripMetadata: true,
			Compression:   9, // 压缩等级 0 - 9
			Filter:        vips.PngFilterNone,
			Interlace:     false, // 交错, 会增大体积，但是浏览器体验好
			Quality:       q,     // 优化程度，仅在palette开启时有效
			Palette:       true,  // 调色板模式, 有效减小体积
			// Dither:      0,
			Bitdepth: 8, // 色深
			// Profile:       "",
		})
	case ImgTypeGIF:
		buf, _, err = img.ExportGIF(&vips.GifExportParams{
			StripMetadata: true,
			Quality:       q,
			// Dither:        0,
			Effort:   7,
			Bitdepth: 8,
		})
	case ImgTypeWEBP:
		if q == 100 {
			buf, _, err = img.ExportWebp(&vips.WebpExportParams{
				Lossless:      true,
				StripMetadata: true,
			})
		} else {
			buf, _, err = img.ExportWebp(&vips.WebpExportParams{
				Quality:         q,
				Lossless:        false,
				StripMetadata:   true,
				ReductionEffort: 4,
			})
		}

	default:
		return nil, errors.New("不支持的图片格式")
	}

	return
}

// Tiny 压缩图片为万能的webp格式，尽可能的减小体积
// 该方法不会改变引用对象
func (that *Encoder) Tiny(maxW, maxH int) (buf []byte, err error) {
	img, err := that.img.Copy()
	if err != nil {
		return nil, err
	}
	_, err = thumbNail(img, maxW, maxH)

	if err != nil {
		return
	}

	buf, _, err = img.ExportWebp(&vips.WebpExportParams{
		Quality:         35,
		Lossless:        false,
		StripMetadata:   true,
		ReductionEffort: 4,
	})

	return
}

// ResizeSelf 对图片进行缩放
// 注意：该方法会改变引用对象
func (that *Encoder) ResizeSelf(width int, height ...int) error {
	buf, err := resizeImg(that.img, width, height...)
	if err != nil {
		return err
	}

	that.Data = buf
	return nil
}

// Resize 对图片进行缩放
// 该方法不会改变引用对象
func (that *Encoder) Resize(width int, height ...int) ([]byte, error) {
	var (
		buf []byte
		err error
	)

	img, err := that.img.Copy()

	if err != nil {
		return nil, err
	}

	buf, err = resizeImg(img, width, height...)

	return buf, err
}

// Height 返回图片的高度，如果是动图，那么返回的是每一帧的高度
func (that *Encoder) Height() int {
	return that.img.PageHeight()
}

// Width 返回图片的宽度
func (that *Encoder) Width() int {
	return that.img.Width()
}

// Pages 返回图片的帧数，非动图返回1
func (that *Encoder) Pages() int {
	return that.img.Pages()
}

// Suffix 根据图片格式返回后缀名
// eg ImgTypeWEBP 会返回 webp
// 注意是不带 . 的
func (that *Encoder) Suffix() string {
	suffixMap := map[ImgType]string{
		ImgTypeWEBP: "webp",
		ImgTypeGIF:  "gif",
		ImgTypeJPEG: "jpg",
		ImgTypePng:  "png",
	}

	return suffixMap[that.Format]
}
