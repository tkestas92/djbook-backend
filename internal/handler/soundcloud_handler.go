package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// SoundCloudSet is a public playlist/set or track from a SoundCloud profile.
type SoundCloudSet struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	URL        string `json:"url"`
	ArtworkURL string `json:"artworkUrl,omitempty"`
	TrackCount int    `json:"trackCount,omitempty"`
	MediaKind  string `json:"mediaKind,omitempty"` // "playlist" or "track"
}

type soundcloudMediaJSON struct {
	Kind         string `json:"kind"`
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	PermalinkURL string `json:"permalink_url"`
	ArtworkURL   string `json:"artwork_url"`
	TrackCount   int    `json:"track_count"`
}

// SoundCloudHandler exposes helpers for SoundCloud public data.
type SoundCloudHandler struct {
	client *http.Client
}

func NewSoundCloudHandler() *SoundCloudHandler {
	return &SoundCloudHandler{
		client: &http.Client{},
	}
}

// GetSets handles GET /soundcloud/sets?url=<profile-or-set-url>
func (h *SoundCloudHandler) GetSets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rawURL := strings.TrimSpace(r.URL.Query().Get("url"))
	if rawURL == "" {
		http.Error(w, "url query parameter is required", http.StatusBadRequest)
		return
	}

	username := parseSoundCloudUsername(rawURL)
	if username == "" {
		set, err := h.fetchSetOEmbed(rawURL)
		if err != nil {
			http.Error(w, "could not resolve soundcloud url", http.StatusBadGateway)
			return
		}
		writeJSON(w, map[string][]SoundCloudSet{"sets": {set}})
		return
	}

	sets, err := h.fetchProfileMedia(username, 3)
	if err != nil {
		http.Error(w, "could not fetch soundcloud sets", http.StatusBadGateway)
		return
	}

	if sets == nil {
		sets = []SoundCloudSet{}
	}
	writeJSON(w, map[string][]SoundCloudSet{"sets": sets})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

var soundCloudReserved = map[string]bool{
	"discover": true, "stream": true, "you": true, "pages": true,
	"signin": true, "upload": true, "tracks": true, "sets": true,
	"likes": true, "followers": true, "following": true,
}

func parseSoundCloudUsername(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	host := strings.ToLower(u.Host)
	if !strings.Contains(host, "soundcloud.com") {
		return ""
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) != 1 {
		return ""
	}
	user := parts[0]
	if soundCloudReserved[user] {
		return ""
	}
	return user
}

func (h *SoundCloudHandler) fetchProfileMedia(username string, limit int) ([]SoundCloudSet, error) {
	req, err := http.NewRequest(http.MethodGet, "https://m.soundcloud.com/"+url.PathEscape(username), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	html := string(body)
	sets := parseMediaFromHTML(html, "playlist", limit)
	if len(sets) == 0 {
		sets = parseMediaFromHTML(html, "track", limit)
	}
	return sets, nil
}

func parseMediaFromHTML(html string, kind string, limit int) []SoundCloudSet {
	needle := `"kind":"` + kind + `"`
	var items []SoundCloudSet
	seen := make(map[string]bool)
	idx := 0

	for len(items) < limit {
		rel := strings.Index(html[idx:], needle)
		if rel == -1 {
			break
		}
		pos := idx + rel
		start := strings.LastIndex(html[:pos], "{")
		if start == -1 {
			idx = pos + 1
			continue
		}

		end := findJSONObjectEnd(html, start)
		if end <= start {
			idx = pos + 1
			continue
		}

		var media soundcloudMediaJSON
		if err := json.Unmarshal([]byte(html[start:end]), &media); err != nil || media.Kind != kind {
			idx = pos + 1
			continue
		}

		itemURL := media.PermalinkURL
		if itemURL == "" || seen[itemURL] {
			idx = pos + 1
			continue
		}
		seen[itemURL] = true

		items = append(items, SoundCloudSet{
			ID:         strconv.FormatInt(media.ID, 10),
			Title:      media.Title,
			URL:        itemURL,
			ArtworkURL: normalizeArtwork(media.ArtworkURL),
			TrackCount: media.TrackCount,
			MediaKind:  kind,
		})
		idx = end
	}

	return items
}

func findJSONObjectEnd(html string, start int) int {
	depth := 0
	max := len(html)
	if start+8000 < max {
		max = start + 8000
	}
	for i := start; i < max; i++ {
		switch html[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i + 1
			}
		}
	}
	return -1
}

func normalizeArtwork(raw string) string {
	if raw == "" {
		return ""
	}
	return strings.Replace(raw, "-large", "-t200x200", 1)
}

func (h *SoundCloudHandler) fetchSetOEmbed(setURL string) (SoundCloudSet, error) {
	endpoint := "https://soundcloud.com/oembed?format=json&url=" + url.QueryEscape(setURL)
	resp, err := h.client.Get(endpoint)
	if err != nil {
		return SoundCloudSet{}, err
	}
	defer resp.Body.Close()

	var payload struct {
		Title        string `json:"title"`
		ThumbnailURL string `json:"thumbnail_url"`
		AuthorName   string `json:"author_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return SoundCloudSet{}, err
	}

	title := payload.Title
	if title == "" {
		title = payload.AuthorName
	}

	return SoundCloudSet{
		ID:         setURL,
		Title:      title,
		URL:        setURL,
		ArtworkURL: payload.ThumbnailURL,
	}, nil
}
