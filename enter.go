package goimgp

import (
	"errors"
	"github.com/davidbyttow/govips/v2/vips"
	"runtime"
)

const (
	ImgTypeJPEG = 1
	ImgTypePng  = 2
	ImgTypeGIF  = 3
	ImgTypeWEBP = 4
)

type Encoder struct {
	img     *vips.ImageRef
	ImgType int
}

var (
	boolFalse   vips.BoolParameter
	intMinusOne vips.IntParameter
)

var (
	ErrorSupportImage = errors.New("不支持的图片格式")
)

func init() {
	vips.Startup(&vips.Config{
		ConcurrencyLevel: runtime.NumCPU(),
	})

	boolFalse.Set(false)
	intMinusOne.Set(-1)
}

// ShowDown 关闭libvips
func ShowDown() {
	vips.Shutdown()
}

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
		encode.ImgType = ImgTypeWEBP
	case vips.ImageTypeGIF:
		encode.ImgType = ImgTypeGIF
	case vips.ImageTypeJPEG:
		encode.ImgType = ImgTypeJPEG
	case vips.ImageTypePNG:
		encode.ImgType = ImgTypePng
	default:
		return nil, ErrorSupportImage
	}

	encode.img = img

	return encode, nil
}

// ToPng 转为png格式并压缩
func (that Encoder) ToPng() ([]byte, error) {
	var (
		err       error
		resultBuf []byte
	)

	resultBuf, _, err = that.img.ExportPng(&vips.PngExportParams{
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

// ToJpeg 转为jpeg格式并压缩
func (that Encoder) ToJpeg() ([]byte, error) {
	var (
		err       error
		resultBuf []byte
	)

	resultBuf, _, err = that.img.ExportJpeg(&vips.JpegExportParams{
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

// ToGif 转为Gif并压缩
func (that Encoder) ToGif() ([]byte, error) {
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

// ToWebp 转为webp格式并压缩
func (that Encoder) ToWebp() ([]byte, error) {
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

// Compress 压缩图片尽可能保证质量
func (that Encoder) Compress() ([]byte, error) {
	switch that.ImgType {
	case ImgTypeJPEG:
		return that.ToJpeg()
	case ImgTypePng:
		return that.ToPng()
	case ImgTypeGIF:
		return that.ToGif()
	case ImgTypeWEBP:
		return that.ToWebp()
	default:
		return nil, errors.New("不支持的图片格式")
	}
}

// Tiny 压缩并限制长宽，默认质量为35，你也可以自行设置
func (that Encoder) Tiny(maxW, maxH int, quality ...int) (buf []byte, err error) {
	q := 35
	if len(quality) > 0 {
		q = quality[0]
		if q <= 0 {
			q = 35
		}

		if q > 99 {
			q = 100
		}
	}
	err = thumbNail(that.img, maxW, maxH)
	if err != nil {
		return nil, err
	}
	switch that.ImgType {
	case ImgTypeJPEG:
		buf, _, err = that.img.ExportJpeg(&vips.JpegExportParams{
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
		buf, _, err = that.img.ExportPng(&vips.PngExportParams{
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
		buf, _, err = that.img.ExportGIF(&vips.GifExportParams{
			StripMetadata: true,
			Quality:       q,
			// Dither:        0,
			Effort:   7,
			Bitdepth: 8,
		})
	case ImgTypeWEBP:
		if q == 100 {
			buf, _, err = that.img.ExportWebp(&vips.WebpExportParams{
				Lossless:      true,
				StripMetadata: true,
			})
		} else {
			buf, _, err = that.img.ExportWebp(&vips.WebpExportParams{
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
