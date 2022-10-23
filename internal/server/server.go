package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"FXinnovation/alertmanager-webhook-rocketchat/internal/config"
	"FXinnovation/alertmanager-webhook-rocketchat/pkg/services/alertprocessor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

type IServer interface {
	Run(ctx context.Context, shutDownCh chan struct{}) error
}

var (
	_ IServer = (*Server)(nil)
)

type Server struct {
	cfg config.AppConfig
	svc alertprocessor.IAlertProcessor

	srv *http.Server
}

func New(cfg config.AppConfig, svc alertprocessor.IAlertProcessor) (*Server, error) {
	s := &Server{
		cfg: cfg,
		svc: svc,

		srv: &http.Server{
			Addr: cfg.ListenAddr,
		},
	}

	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/webhook", s.webhook)

	return s, nil
}

func (s *Server) Run(ctx context.Context, shutDownCh chan struct{}) error {
	idleConnsClosed := make(chan struct{})

	go func(ctx context.Context) {
		<-ctx.Done()

		log.Debug("http server shutdown initiated")

		shutdownCtx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		if err := s.srv.Shutdown(shutdownCtx); err != nil {
			log.Warnf("cannot gracefully shutdown http server listener: %v", err)
		}
		log.Infof("http server shutdown completed")

		close(idleConnsClosed)

		shutDownCh <- struct{}{}
	}(ctx)

	log.Infof("listening on: %v", s.cfg.ListenAddr)

	if err := s.srv.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server listen error: %w", err)
		}
	}

	<-idleConnsClosed

	return nil
}

func (s *Server) webhook(w http.ResponseWriter, r *http.Request) {
	data, err := readRequestBody(r)
	if err != nil {
		err := sendJSONResponse(w, http.StatusBadRequest, err.Error())
		if err != nil {
			log.Warnf("cannot send error: %v", err)
		}
		return
	}

	if err := s.svc.SendNotification(data); err != nil {
		log.Errorf("cannot send notification: %v", err)

		if err := sendJSONResponse(w, http.StatusInternalServerError, fmt.Sprintf("cannot send notification: %v", err)); err != nil {
			log.Warnf("cannot send error: %v", err)
		}

		return
	}

	if err := sendJSONResponse(w, http.StatusOK, "Success"); err != nil {
		log.Warnf("cannot send response: %v", err)

		return
	}
}
