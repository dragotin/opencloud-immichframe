// SPDX-FileCopyrightText: 2026 Klaas Freitag <opensource@freisturz.de>
//
// SPDX-License-Identifier: GPL-3.0-or-later

package frame

import "net/http"

// handleGetConfig returns the client settings.
func (s *Server) handleGetConfig(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, ClientSettingsDto{
		Interval:            s.cfg.Interval,
		TransitionDuration:  s.cfg.TransitionDuration,
		ShowClock:           s.cfg.ShowClock,
		ClockFormat:         s.cfg.ClockFormat,
		ClockDateFormat:     s.cfg.ClockDateFormat,
		ShowProgressBar:     s.cfg.ShowProgressBar,
		ShowPhotoDate:       s.cfg.ShowPhotoDate,
		PhotoDateFormat:     s.cfg.PhotoDateFormat,
		ShowImageDesc:       s.cfg.ShowImageDesc,
		ShowImageLocation:   s.cfg.ShowImageLocation,
		ImageLocationFormat: s.cfg.ImageLocationFormat,
		ShowAlbumName:       s.cfg.ShowAlbumName,
		Language:            s.cfg.Language,
	})
}

// handleGetVersion returns the service version. This route is unauthenticated,
// matching upstream ImmichFrame.
func (s *Server) handleGetVersion(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.version)
}

// handleWeather is a stub: there is no weather backend, so it returns null and
// the web UI hides the weather overlay.
func (s *Server) handleWeather(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, nil)
}
