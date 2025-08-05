package main

import (
	"fmt"
	"io"
	"net/http"

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
	mediaType := header.Header.Get("Content-Type")
	b, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't read the file", err)
		return
	}
	video, err := cfg.db.GetVideo(videoID)
	if err!=nil{
		respondWithError(w, http.StatusBadRequest, "couldn't get thumbnail file", err)
		return
	}
	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "you are not authorized for this", nil)
		return
	}
	videoId := video.ID.String()
	s := fmt.Sprintf("/api/thumbnails/%s", videoId)
	video.ThumbnailURL = &s
	videoThumbnails[video.ID] = thumbnail{
		mediaType: mediaType,
		data: b,	
	}
	err = cfg.db.UpdateVideo(video)
	if err != nil{
		respondWithError(w, http.StatusInternalServerError, "couldn't update the video", err)
		return
	}
	respondWithJSON(w, http.StatusOK, videoThumbnails[videoID])
}
