package controller

import (
	"InstantWellnessKits/src/usecase"
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
)

type ImportHandler struct {
	uc *usecase.ImportOrdersUseCase
}

func NewImportHandler(uc *usecase.ImportOrdersUseCase) *ImportHandler {
	return &ImportHandler{
		uc: uc,
	}
}

func (h *ImportHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(rw, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		log.Println("Error retrieving file from form data:", err)
		http.Error(rw, "Failed to retrieve file from form data", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if header.Size == 0 {
		http.Error(rw, "Uploaded file is empty", http.StatusBadRequest)
		return
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(rw, "Failed to read file", http.StatusInternalServerError)
		return
	}

	go func() {
		bytesReader := bytes.NewReader(fileBytes)

		log.Println("Background import started...")
		_, err = h.uc.Execute(context.Background(), bytesReader)
		if err != nil {
			http.Error(rw, "Failed to import orders: "+err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("Background import finished!")
	}()

	rw.WriteHeader(http.StatusAccepted)
}
