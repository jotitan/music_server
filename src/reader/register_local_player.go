package reader

import (
	"fmt"
	"github.com/jotitan/music_server/logger"
	"io"
	"net/http"
	"strings"
)

func RegisterService(urlMusicService, device, localUrl string) (string, string) {
	resp, err := http.Get(fmt.Sprintf("%s/shareService?device=%s&url=%s", urlMusicService, device, localUrl))
	if err == nil && resp.StatusCode == 200 {
		data, _ := io.ReadAll(resp.Body)
		shareID := string(data)
		sessionID := extractIdFromHeader(resp)
		logger.GetLogger().Info("Connected to music server")
		return sessionID, shareID
		// Extract cookie with sessionid and id device
	} else {
		panic("Impossible to connect to music server")
	}
}

func extractIdFromHeader(resp *http.Response) string {
	cookie := resp.Header.Get("Set-Cookie")
	// Extract jsessionid
	return strings.Split(cookie, "=")[1]
}
