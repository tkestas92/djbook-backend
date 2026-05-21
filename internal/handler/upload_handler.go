package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/djbook/backend/internal/auth"
	"github.com/djbook/backend/internal/service"
	"github.com/google/uuid"
)

const maxUploadSize = 10 << 20 // 10 MB

// UploadHandler serves photo upload endpoints.
type UploadHandler struct {
	profileSvc *service.ProfileService
	photoDir   string
}

func NewUploadHandler(profileSvc *service.ProfileService, photoDir string) *UploadHandler {
	return &UploadHandler{
		profileSvc: profileSvc,
		photoDir:   photoDir,
	}
}

type uploadPhotoResponse struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// UploadPhoto handles POST /upload/photo.
func (h *UploadHandler) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := auth.RequireUserID(r.Context())
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "invalid multipart form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("photo")
	if err != nil {
		http.Error(w, "photo field required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	profileID := r.FormValue("profileId")
	if profileID == "" {
		profiles, err := h.profileSvc.ListByUserID(r.Context(), userID)
		if err != nil {
			http.Error(w, "profiles error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if len(profiles) == 0 {
			http.Error(w, "profileId required", http.StatusBadRequest)
			return
		}
		profileID = profiles[0].ID
	}

	if err := h.profileSvc.AssertOwner(r.Context(), profileID, userID); err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	fileUUID := uuid.New().String()
	filename := fileUUID + ".jpg"
	destPath := filepath.Join(h.photoDir, filename)

	out, err := os.Create(destPath)
	if err != nil {
		http.Error(w, "failed to save photo", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		os.Remove(destPath)
		http.Error(w, "failed to save photo", http.StatusInternalServerError)
		return
	}

	url := fmt.Sprintf("/uploads/photos/%s", filename)
	photo, err := h.profileSvc.AddPhoto(r.Context(), profileID, url)
	if err != nil {
		os.Remove(destPath)
		http.Error(w, "failed to save photo record: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(uploadPhotoResponse{
		ID:  photo.ID,
		URL: photo.URL,
	})
}
