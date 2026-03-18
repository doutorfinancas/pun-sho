package api

import (
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var templates *template.Template

func LoadTemplates(log *zap.Logger) *template.Template {
	funcMap := template.FuncMap{
		"safeURL": func(s string) template.URL {
			return template.URL(s)
		},
		"seq": func(n int) []int {
			s := make([]int, n)
			for i := range s {
				s[i] = i + 1
			}
			return s
		},
		"subtract": func(a, b int) int { return a - b },
		"add":      func(a, b int) int { return a + b },
		"multiply": func(a, b int) int { return a * b },
		"deref": func(v *int) int {
			if v == nil {
				return 0
			}
			return *v
		},
		"isExpired": func(t *time.Time) bool {
			if t == nil {
				return false
			}
			if t.Year() < 2 {
				return false
			}
			return t.Before(time.Now())
		},
		"isExpiredYear": func(t *time.Time) bool {
			if t == nil {
				return true
			}
			return t.Year() < 2
		},
		"isLimitReached": func(limit *int, count int) bool {
			if limit == nil {
				return false
			}
			return *limit > 0 && count >= *limit
		},
		"formatDate": func(t *time.Time) string {
			if t == nil {
				return ""
			}
			return t.Format("Jan 02, 2006")
		},
		"formatDateTime": func(t *time.Time) string {
			if t == nil {
				return ""
			}
			return t.Format("Jan 02, 2006 15:04")
		},
	}

	patterns := []string{
		"templates/*.html",
		"templates/**/*.html",
		"templates/**/**/*.html",
	}

	tmpl := template.New("").Funcs(funcMap)

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			log.Warn("Failed to glob templates", zap.String("pattern", pattern), zap.Error(err))
			continue
		}
		for _, match := range matches {
			_, err := tmpl.ParseFiles(match)
			if err != nil {
				log.Warn("Failed to parse template", zap.String("file", match), zap.Error(err))
			}
		}
	}

	templates = tmpl
	return tmpl
}

func renderTemplate(c *gin.Context, name string, data interface{}) {
	if templates == nil {
		c.String(http.StatusInternalServerError, "Templates not loaded")
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := templates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.String(http.StatusInternalServerError, "Template error: %s", err.Error())
	}
}
