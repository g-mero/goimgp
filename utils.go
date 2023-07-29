package goimgp

import (
	"github.com/davidbyttow/govips/v2/vips"
	"math"
)

// 图片大小缩小
func thumbNail(img *vips.ImageRef, desW, desH int) error {
	var (
		srcW = float64(img.Width())
		srcH = float64(img.PageHeight())
	)

	// 不缩放
	if img.Width() <= desW && img.PageHeight() <= desH {
		return nil
	}

	// 宽高比
	ratio := srcW / srcH

	if desW > 0 && desH <= 0 {
		return img.Thumbnail(desW, int(srcH/ratio), 0)
	}

	if desW <= 0 && desH > 0 {
		return img.Thumbnail(int(float64(desH)*ratio), desH*img.Pages(), 0)
	}

	// 自适应计算，确保宽高满足条件
	if desW > 0 && desH > 0 {
		thumbRatio := math.Min(float64(desW)/srcW, float64(desH)/srcH)

		h := int(math.Ceil(srcH * thumbRatio))
		w := int(float64(h) * ratio)

		return img.Thumbnail(w, h*img.Pages(), 0)
	}

	return nil
}
