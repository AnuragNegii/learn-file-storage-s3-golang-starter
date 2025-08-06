package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}


	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here
	const maxMemory = 10 << 20
	err = r.ParseMultipartForm(maxMemory)	
	if err != nil{
		respondWithError(w, http.StatusBadRequest, "couldn't parse maxMemory", err)
		return
	}
	file, header, err := r.FormFile("thumbnail")
	if err != nil{
		respondWithError(w, http.StatusBadRequest, "couldn't get thumbnail file", err)
		return
	}
	defer file.Close()	
	video, err := cfg.db.GetVideo(videoID)
	if err!=nil{
		respondWithError(w, http.StatusBadRequest, "couldn't get thumbnail file", err)
		return
	}
	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "you are not authorized for this", nil)
		return
	}
	fileString, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil{
		respondWithError(w, http.StatusBadRequest, "couldn't get the file type from the file", nil)
		return
	}
	var fileType string
	switch fileString{
	case "image/jpeg":
		fileType = ".jpg"	
	case "image/png":
		fileType = ".png"
	default:
		respondWithError(w, http.StatusBadRequest, "wrong file type submitted", nil)
		return
	}
	videoFile := fmt.Sprintf("%s%s",video.ID.String(),fileType)
	newPath := filepath.Join(cfg.assetsRoot,videoFile)
	dst, err := os.Create(newPath)	
	if err != nil{
		respondWithError(w, http.StatusBadRequest, "couldn't create file", err)
		return
	}
	defer dst.Close()
	_, err = io.Copy(dst, file)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't read the file", err)
		return
	}
	publicUrl := "http://localhost:8091/assets/" + videoFile
	video.ThumbnailURL = &publicUrl
	err = cfg.db.UpdateVideo(video)
	if err != nil{
		respondWithError(w, http.StatusInternalServerError, "couldn't update the video", err)
		return
	}
	respondWithJSON(w, http.StatusOK, video)
}
