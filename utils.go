package goimgp

import (
	"errors"
	"github.com/davidbyttow/govips/v2/vips"
)

// 图片大小缩小
func thumbNail(img *vips.ImageRef, desW, desH int) (buf []byte, err error) {
	var (
		srcW = img.Width()
		srcH = img.PageHeight()
	)

	if desW > srcW && desH > srcH {
		buf, _, err = img.ExportNative()

		return
	}

	ratio := float64(srcH) / float64(srcW)

	if desW > 0 && desH <= 0 {
		err = img.ThumbnailWithSize(desW, int(float64(desW)*ratio), 0, vips.SizeDown)
	} else if desW <= 0 && desH > 0 {
		err = img.ThumbnailWithSize(int(float64(desH)/ratio), desH, 0, vips.SizeDown)
	} else if desW > 0 && desH > 0 {
		err = img.ThumbnailWithSize(desW, desH, 0, vips.SizeDown)
	}

	buf, _, err = img.ExportNative()

	return
}

// resize
func resizeImg(img *vips.ImageRef, desW int, desH ...int) ([]byte, error) {
	var (
		w   = desW
		h   = 0
		err error
		buf []byte
	)

	if w < 0 {
		w = 0
	}

	if len(desH) > 0 {
		h = desH[0]

		if h < 0 {
			h = 0
		}
	}

	if w == 0 && h == 0 {
		return nil, errors.New("宽高数据输入不合法")
	}

	ratio := float64(img.PageHeight()) / float64(img.Width())
	if w == 0 {
		// ratio := float64(h) / float64(img.PageHeight())
		// err = img.Resize(ratio, -1)

		err = img.ThumbnailWithSize(int(float64(h)/ratio), h, 0, vips.SizeForce)
		if err != nil {
			return nil, err
		}
	} else if h == 0 {
		err = img.ThumbnailWithSize(w, int(float64(w)*ratio), 0, vips.SizeForce)
		if err != nil {
			return nil, err
		}
	} else {
		err = img.ThumbnailWithSize(w, h, 0, vips.SizeForce)
		if err != nil {
			return nil, err
		}
	}

	buf, _, err = img.ExportNative()

	return buf, err
}
