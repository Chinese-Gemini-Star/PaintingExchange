package service

import (
	"fmt"
	"gocv.io/x/gocv"
	"image"
	"io"
	"mime/multipart"
)

type SizeType int

const (
	BigSize SizeType = 3000
	MidSize SizeType = 300
)

// FileToMat 将multipart.File转换为gocv.Mat
func FileToMat(file multipart.File) (gocv.Mat, error) {
	// 读取文件内容
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return gocv.NewMat(), fmt.Errorf("failed to read file: %w", err)
	}

	// 将字节数据解码为 gocv.Mat
	imgMat, err := gocv.IMDecode(fileBytes, gocv.IMReadColor)
	if err != nil {
		return gocv.NewMat(), fmt.Errorf("failed to decode image: %w", err)
	}

	return imgMat, nil
}

// ResizeImage 根据目标高度进行等比缩放并使用插值方法
func ResizeImage(img gocv.Mat, originalWidth, originalHeight int, sizeType SizeType) gocv.Mat {
	var targetHeight int = int(sizeType)
	// 如果图片高度已经符合要求，则直接返回原图(只会出现在大尺寸上)
	if originalHeight >= targetHeight && sizeType == BigSize {
		return img.Clone()
	}

	// 计算缩放比例和目标宽度
	scale := float64(targetHeight) / float64(originalHeight)
	targetWidth := int(float64(originalWidth) * scale)

	// 选择插值方法
	//interpolation := gocv.InterpolationCubic
	interpolation := gocv.InterpolationArea
	if originalHeight > targetHeight {
		interpolation = gocv.InterpolationArea
	}

	// 调整图片尺寸
	resizedImg := gocv.NewMat()
	gocv.Resize(img, &resizedImg, image.Point{X: targetWidth, Y: targetHeight}, 0, 0, interpolation)

	return resizedImg
}
