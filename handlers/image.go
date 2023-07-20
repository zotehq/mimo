package handlers

import (
	"fmt"
	"image"
	"net/http"
	"strconv"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
)

func ResizeImage(c *gin.Context) {
	start := time.Now()

	imageURL := c.Query("url")

	if imageURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'url' query parameter"})
		return
	}

	resp, err := http.Get(imageURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch the image"})
		return
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid image format"})
		return
	}

	widthStr := c.DefaultQuery("width", "0")
	heightStr := c.DefaultQuery("height", "0")
	blurStr := c.DefaultQuery("blur", "0")
	sharpenStr := c.DefaultQuery("sharpen", "0")
	gammaStr := c.DefaultQuery("gamma", "1.0")
	contrastStr := c.DefaultQuery("contrast", "1.0")
	brightnessStr := c.DefaultQuery("brightness", "0")
	saturationStr := c.DefaultQuery("saturation", "1.0")

	width, err := strconv.Atoi(widthStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid width parameter"})
		return
	}

	height, err := strconv.Atoi(heightStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid height parameter"})
		return
	}

	blur, err := strconv.Atoi(blurStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid blur parameter"})
		return
	}

	sharpen, err := strconv.Atoi(sharpenStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sharpen parameter"})
		return
	}

	gamma, err := strconv.ParseFloat(gammaStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gamma parameter"})
		return
	}

	contrast, err := strconv.ParseFloat(contrastStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contrast parameter"})
		return
	}

	brightness, err := strconv.Atoi(brightnessStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brightness parameter"})
		return
	}

	saturation, err := strconv.ParseFloat(saturationStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid saturation parameter"})
		return
	}

	if width > 0 || height > 0 {
		img = imaging.Resize(img, width, height, imaging.Lanczos)
	}

	if blur > 0 {
		img = imaging.Blur(img, float64(blur))
	}

	if sharpen > 0 {
		img = imaging.Sharpen(img, float64(sharpen))
	}

	if gamma != 1.0 {
		img = imaging.AdjustGamma(img, gamma)
	}

	if contrast != 1.0 {
		img = imaging.AdjustContrast(img, contrast)
	}

	if brightness != 0 {
		img = imaging.AdjustBrightness(img, float64(brightness))
	}

	if saturation != 1.0 {
		img = imaging.AdjustSaturation(img, saturation)
	}

	c.Header("Content-Type", "image/jpeg")
	c.Header("Cache-Control", "max-age=3600")

	responseTime := time.Since(start).Milliseconds()
	c.Writer.Header().Set("X-Response-Time", fmt.Sprintf("%d ms", responseTime))

	if err := imaging.Encode(c.Writer, img, imaging.JPEG); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode the image"})
		return
	}
}
