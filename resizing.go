package gimage

import (
	"image"
	"io"
	"sync"

	"github.com/disintegration/imaging"
)

var (
	// CropMedium is default medium crop size
	CropMedium = CropSize{
		Height: 0,
		Width:  250,
	}

	// CropLarge is default large crop size
	CropLarge = CropSize{
		Height: 100,
		Width:  700,
	}

	// CropSmall is default small crop size
	CropSmall = CropSize{
		Height: 0,
		Width:  100,
	}
)

// Resizer struct provides methods
type Resizer struct {
	image   image.Image
	resized image.Image
}

// CropSize represents a cropping dimension
type CropSize struct {
	Height int
	Width  int
}

// ReadImage image reads image from a Reader interface.
func (r *Resizer) ReadImage(source io.Reader) error {
	img, err := imaging.Decode(source)
	if err != nil {
		return err
	}

	// r.resized is initially set to the original image so that if Resize
	// is not called we can return the original image.
	r.image, r.resized = img, img
	return nil
}

// Resize resizes the image to
func (r *Resizer) Resize(size CropSize) {
	r.resized = imaging.Resize(r.image, size.Width, size.Height, imaging.Lanczos)
}

// ResizeMultiple returns resized copy of the image as per the crop sizes provided.
// This method copies the current struct and resizes the image in the copies.
func (r *Resizer) ResizeMultiple(sizes ...CropSize) []*Resizer {
	var resizers []*Resizer
	var wg sync.WaitGroup

	for _, cropSize := range sizes {
		resizerCopy := Resizer{}
		resizerCopy.image, resizerCopy.resized = r.image, r.resized

		resizers = append(resizers, &resizerCopy)

		wg.Add(1)
		go func(cropSize CropSize, wg *sync.WaitGroup) {
			resizerCopy.Resize(cropSize)
			wg.Done()
		}(cropSize, &wg)
	}

	wg.Wait()
	return resizers
}

func (r *Resizer) Write(w io.Writer) error {
	return imaging.Encode(w, r.resized, imaging.JPEG)
}

// GetResizedImageProps returns the height and width of the resized image.
func (r *Resizer) GetResizedImageProps() CropSize {
	return CropSize{
		Height: r.resized.Bounds().Dy(),
		Width:  r.resized.Bounds().Dx(),
	}
}

func NewResizer() *Resizer {
	return &Resizer{}
}
