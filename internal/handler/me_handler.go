package handler

import (
	"encoding/json"
	"net/http"

	"github.com/djbook/backend/internal/auth"
	"github.com/djbook/backend/internal/model"
	"github.com/djbook/backend/internal/service"
)

// MeHandler serves authenticated user endpoints.
type MeHandler struct {
	userSvc    *service.UserService
	profileSvc *service.ProfileService
	eventSvc   *service.EventService
}

func NewMeHandler(
	userSvc *service.UserService,
	profileSvc *service.ProfileService,
	eventSvc *service.EventService,
) *MeHandler {
	return &MeHandler{
		userSvc:    userSvc,
		profileSvc: profileSvc,
		eventSvc:   eventSvc,
	}
}

type meEventResponse struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Venue         string   `json:"venue"`
	Date          string   `json:"date"`
	StartTime     string   `json:"startTime"`
	EndTime       *string  `json:"endTime,omitempty"`
	Notes         *string  `json:"notes,omitempty"`
	AmountEur     *float64 `json:"amountEur,omitempty"`
	EventStatus   string   `json:"eventStatus"`
	PaymentStatus string   `json:"paymentStatus"`
	CreatedAt     string   `json:"createdAt"`
}

type mePhotoResponse struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	SortOrder int    `json:"sortOrder"`
	CreatedAt string `json:"createdAt"`
}

type meSocialLinkResponse struct {
	ID       string `json:"id"`
	Platform string `json:"platform"`
	URL      string `json:"url"`
}

type meProfileResponse struct {
	ID          string                 `json:"id"`
	DjName      string                 `json:"djName"`
	Bio         *string                `json:"bio,omitempty"`
	Genres      []string               `json:"genres"`
	Photos      []mePhotoResponse      `json:"photos"`
	SocialLinks []meSocialLinkResponse `json:"socialLinks"`
	Events      []meEventResponse      `json:"events"`
	CreatedAt   string                 `json:"createdAt"`
}

type meResponse struct {
	ID        string              `json:"id"`
	Email     string              `json:"email"`
	Profiles  []meProfileResponse `json:"profiles"`
	CreatedAt string              `json:"createdAt"`
}

// GetMe handles GET /me — returns the authenticated user with all DJ profiles.
func (h *MeHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := auth.RequireUserID(r.Context())
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	u, err := h.userSvc.GetByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	profiles, err := h.profileSvc.ListByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, "profiles error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responseProfiles := make([]meProfileResponse, 0, len(profiles))
	for _, p := range profiles {
		rp, err := h.buildProfileResponse(r, p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		responseProfiles = append(responseProfiles, rp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(meResponse{
		ID:        u.ID,
		Email:     u.Email,
		Profiles:  responseProfiles,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *MeHandler) buildProfileResponse(r *http.Request, p *model.DjProfile) (meProfileResponse, error) {
	ctx := r.Context()

	genres, err := h.profileSvc.GetGenres(ctx, p.ID)
	if err != nil {
		return meProfileResponse{}, err
	}
	if genres == nil {
		genres = []string{}
	}

	dbPhotos, err := h.profileSvc.GetPhotos(ctx, p.ID)
	if err != nil {
		return meProfileResponse{}, err
	}
	photos := make([]mePhotoResponse, 0, len(dbPhotos))
	for _, ph := range dbPhotos {
		photos = append(photos, mePhotoResponse{
			ID:        ph.ID,
			URL:       ph.URL,
			SortOrder: ph.SortOrder,
			CreatedAt: ph.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	dbLinks, err := h.profileSvc.GetSocialLinks(ctx, p.ID)
	if err != nil {
		return meProfileResponse{}, err
	}
	links := make([]meSocialLinkResponse, 0, len(dbLinks))
	for _, l := range dbLinks {
		links = append(links, meSocialLinkResponse{
			ID:       l.ID,
			Platform: l.Platform,
			URL:      l.URL,
		})
	}

	dbEvents, err := h.eventSvc.List(ctx, p.ID, nil)
	if err != nil {
		return meProfileResponse{}, err
	}
	events := make([]meEventResponse, 0, len(dbEvents))
	for _, e := range dbEvents {
		events = append(events, meEventResponse{
			ID:            e.ID,
			Title:         e.Title,
			Venue:         e.Venue,
			Date:          e.Date.Format("2006-01-02"),
			StartTime:     e.StartTime,
			EndTime:       e.EndTime,
			Notes:         e.Notes,
			AmountEur:     e.AmountEur,
			EventStatus:   string(e.EventStatus),
			PaymentStatus: string(e.PaymentStatus),
			CreatedAt:     e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return meProfileResponse{
		ID:          p.ID,
		DjName:      p.DjName,
		Bio:         p.Bio,
		Genres:      genres,
		Photos:      photos,
		SocialLinks: links,
		Events:      events,
		CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}
