package app

import (
	"compress/gzip"
	"flag"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/GevorkovG/go-shortener-tlp/config"
	logg "github.com/GevorkovG/go-shortener-tlp/internal/log"
	"github.com/GevorkovG/go-shortener-tlp/internal/storage"
	"github.com/go-chi/chi"
)

type App struct {
	cfg     *config.AppConfig
	storage *storage.Storage
}

func NewApp(cfg *config.AppConfig) *App {
	return &App{
		cfg:     cfg,
		storage: storage.NewStorage(),
	}
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	// w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
	return w.Writer.Write(b)
}

func defaultHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	io.WriteString(w, "<html><body>"+strings.Repeat("Hello, world<br>", 20)+"</body></html>")
}

func gzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// проверяем, что клиент поддерживает gzip-сжатие
		// это упрощённый пример. В реальном приложении следует проверять все
		// значения r.Header.Values("Accept-Encoding") и разбирать строку
		// на составные части, чтобы избежать неожиданных результатов
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// если gzip не поддерживается, передаём управление
			// дальше без изменений
			next.ServeHTTP(w, r)
			return
		}

		// создаём gzip.Writer поверх текущего w
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		// передаём обработчику страницы переменную типа gzipWriter для вывода данных
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

func Run() {
	conf := config.NewCfg()
	newApp := NewApp(conf)
	r := chi.NewRouter()

	r.Use(logg.WithLogging)
	r.Use(gzipHandle)

	r.Post("/api/shorten", newApp.JSONGetShortURL)
	r.Get("/{id}", newApp.GetOriginURL)
	r.Post("/", newApp.GetShortURL)

	flag.Parse()
	log.Fatal(http.ListenAndServe(conf.Host, r))

}
