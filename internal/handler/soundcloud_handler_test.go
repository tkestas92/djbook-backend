package handler

import "testing"

func TestParseMediaFromHTML(t *testing.T) {
	html := `{"kind":"playlist","id":1,"title":"First Set","permalink_url":"https://soundcloud.com/u/sets/a","artwork_url":"https://i1.sndcdn.com/artworks-x-large.jpg","track_count":5}
{"kind":"playlist","id":2,"title":"Second","permalink_url":"https://soundcloud.com/u/sets/b","track_count":3}
{"kind":"track","id":9,"title":"My Track","permalink_url":"https://soundcloud.com/u/my-track","artwork_url":"https://i1.sndcdn.com/a-large.jpg"}`
	playlists := parseMediaFromHTML(html, "playlist", 3)
	if len(playlists) != 2 {
		t.Fatalf("expected 2 playlists, got %d", len(playlists))
	}
	if playlists[0].Title != "First Set" || playlists[0].MediaKind != "playlist" {
		t.Fatalf("unexpected first playlist: %+v", playlists[0])
	}
	if playlists[0].ArtworkURL != "https://i1.sndcdn.com/artworks-x-t200x200.jpg" {
		t.Fatalf("artwork normalize failed: %s", playlists[0].ArtworkURL)
	}

	tracks := parseMediaFromHTML(html, "track", 3)
	if len(tracks) != 1 || tracks[0].Title != "My Track" || tracks[0].MediaKind != "track" {
		t.Fatalf("unexpected track parse: %+v", tracks)
	}
}

func TestParseTracksFromKantrybesHTML(t *testing.T) {
	html := `prefix {"kind":"track","id":1620722019,"title":"Gresia 2023","permalink_url":"https://soundcloud.com/kantrybes/a","artwork_url":"https://i1.sndcdn.com/x-large.jpg"}
{"kind":"track","id":1036010077,"title":"Gresia 2021","permalink_url":"https://soundcloud.com/kantrybes/b","artwork_url":""}
{"kind":"track","id":825929902,"title":"Gresia 2020","permalink_url":"https://soundcloud.com/kantrybes/c","artwork_url":""} suffix`

	tracks := parseMediaFromHTML(html, "track", 3)
	if len(tracks) != 3 {
		t.Fatalf("expected 3 tracks, got %d", len(tracks))
	}
	if tracks[0].MediaKind != "track" || tracks[0].Title != "Gresia 2023" {
		t.Fatalf("unexpected first track: %+v", tracks[0])
	}
}

func TestParseSoundCloudUsername(t *testing.T) {
	if got := parseSoundCloudUsername("https://soundcloud.com/forss"); got != "forss" {
		t.Fatalf("profile username: %q", got)
	}
	if got := parseSoundCloudUsername("https://soundcloud.com/forss/sets/live"); got != "" {
		t.Fatalf("set url should not be profile: %q", got)
	}
}
