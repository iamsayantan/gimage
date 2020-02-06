package server

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/iamsayantan/gimage"
	uuid "github.com/satori/go.uuid"

	"github.com/go-chi/chi"
)

var maxFileUploadSize = int64(10 << 20) // 10<<20 is 10 MB

type uploadHandler struct {
	resizer  *gimage.Resizer
	uploader *gimage.S3Uploader
}

func (u *uploadHandler) Route() chi.Router {
	r := chi.NewRouter()
	r.Post("/image", u.uploadImage)

	return r
}

func (u *uploadHandler) uploadImage(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(maxFileUploadSize)

	file, fileHeader, err := r.FormFile("image")
	if err != nil || fileHeader.Size == 0 {
		sendError(w, http.StatusBadRequest, "Invalid image")
		return
	}

	contentType := fileHeader.Header.Get("Content-Type")
	fileExtension := strings.Split(fileHeader.Filename, ".")

	defer file.Close()
	err = u.resizer.ReadImage(file)

	// cropDimension := gimage.CropSize{
	// 	Height: 0,
	// 	Width:  250,
	// }

	resizers := u.resizer.ResizeMultiple(gimage.CropLarge, gimage.CropMedium, gimage.CropSmall)
	var uploadLocations []string

	for _, resizer := range resizers {
		reader, writer := io.Pipe()
		go func() {
			resizer.Write(writer)
			writer.Close()
		}()

		imgProps := resizer.GetResizedImageProps()

		folder := strconv.Itoa(imgProps.Height) + "X" + strconv.Itoa(imgProps.Width) + "/"
		uploadReq := gimage.UploadRequest{
			Bucket:        "tagfi-s3-dev1",
			ContentType:   contentType,
			Payload:       reader,
			UploadPath:    "bcc/" + folder,
			FileExtension: fileExtension[len(fileExtension)-1],
		}

		uploadLocation, err := u.uploader.Upload(uploadReq)
		if err != nil {
			sendError(w, http.StatusBadRequest, err.Error())
			return
		}
		uploadLocations = append(uploadLocations, uploadLocation)
	}
	// u.resizer.Resize(cropDimension)

	// reader, writer := io.Pipe()
	// go func() {
	// 	u.resizer.Write(writer)
	// 	writer.Close()
	// }()

	// uploadReq := gimage.UploadRequest{
	// 	Bucket:        "tagfi-s3-dev1",
	// 	ContentType:   contentType,
	// 	Payload:       reader,
	// 	UploadPath:    "bcc/",
	// 	FileExtension: fileExtension[len(fileExtension)-1],
	// }

	// uploadLocation, err := u.uploader.Upload(uploadReq)
	// if err != nil {
	// 	sendError(w, http.StatusBadRequest, err.Error())
	// 	return
	// }

	resp := struct {
		Location []string `json:"locations"`
	}{Location: uploadLocations}

	sendResponse(w, http.StatusCreated, "Image successfully uploaded", resp)
}

func resizePipeline(resizer *gimage.Resizer, uploader *gimage.S3Uploader) {

}

func generateUniqueFilename() string {
	uid := uuid.NewV4()
	idString := uid.String()

	idString = strings.Join(strings.Split(idString, "-"), "")
	return strings.ToUpper(idString)
}

// NewUploadHandler returns WebHandler implementation of uploading related request handlers.
func NewUploadHandler(resizer *gimage.Resizer, uploader *gimage.S3Uploader) WebHandler {
	return &uploadHandler{
		resizer:  resizer,
		uploader: uploader,
	}
}
