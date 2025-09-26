package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil{
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil{
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil{
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}
	fmt.Println("uploading vido file for video", videoID, "by user", userID)

	const maxMemory = 1 << 30
	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, maxMemory)	

	file, header, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't get the video file", err)
		return
	}
	defer file.Close()

	fileString, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil{
		respondWithError(w, http.StatusBadRequest, "Couldn't get the file type from the file", err)
		return
	}
	if fileString != "video/mp4"{
		respondWithError(w, http.StatusBadRequest, "wrong file type submitted", nil)
		return
	}
	tmpfile, err := os.CreateTemp("","tubely-upload.mp4")
	if err != nil{
		respondWithError(w, http.StatusBadRequest, "Couldn't create the file", nil)
		return
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()
	io.Copy(tmpfile, file)
	_, err = tmpfile.Seek(0, io.SeekStart)
	if err != nil{
		respondWithError(w, http.StatusBadRequest, "Couldn't reset the temp file pointer", nil)
		return
	}
	aspectratio, err := getVideoAspectRation(tmpfile.Name())
	if err != nil{
		respondWithError(w, http.StatusInternalServerError, "couldn't determine the aspect ratio", nil)
		return 
	}
	key := fmt.Sprintf("%s/%s.mp4", aspectratio, "12k3jkljlksdajlfgghalkhg123123as")
	_, err = cfg.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket: &cfg.s3Bucket,
		Key: aws.String(key),
		Body: tmpfile,
		ContentType: aws.String("video/mp4"),
	})
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't upload file to s3", err)
		return
	}
	fmt.Printf("%v", key)
	urlString := "https://" + cfg.s3Bucket + ".s3." + cfg.s3Region + ".amazonaws.com/" + key
	fmt.Printf("%v", urlString)
	video, err := cfg.db.GetVideo(videoID)	
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't get thumbnail file", err)
		return
	}
	video.VideoURL = &urlString
	err = cfg.db.UpdateVideo(video)
	if err != nil {
    respondWithError(w, http.StatusInternalServerError, "Couldn't update video URL in database", err)
    return
	}

	respondWithJSON(w, 200, response{
		CreatedAt: video.CreatedAt,
		Description: video.Description,
		ID: video.ID,
		Thumbnail_url: *video.ThumbnailURL,
		Title: video.Title,
		UpdatedAT: video.UpdatedAt,
		UserID: video.UserID,
		VideoURL: *video.VideoURL,
	}) 

}

type response struct{
	CreatedAt time.Time`json:"created_at"`
	Description string `json:"description"`
	ID uuid.UUID `json:"id"`
	Thumbnail_url string `json:"thumbnail_url"`
	Title string `json:"title"`
	UpdatedAT time.Time `json:"updated_at"`
	UserID uuid.UUID `json:"user_id"`
	VideoURL string `json:"video_url"`
}
