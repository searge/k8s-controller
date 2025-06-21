package server

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
)

// Start запускає HTTP сервер на вказаному порту
func Start(port int) error {
	addr := fmt.Sprintf(":%d", port)

	log.Info().Msgf("Starting HTTP server on %s", addr)

	handler := func(ctx *fasthttp.RequestCtx) {
		path := string(ctx.Path())

		log.Info().Msgf("Request: %s %s", ctx.Method(), path)

		switch path {
		case "/health":
			ctx.SetStatusCode(200)
			if _, err := fmt.Fprintf(ctx, `{"status":"ok"}`); err != nil {
				log.Error().Err(err).Msg("Failed to write health response")
			}
		default:
			if _, err := fmt.Fprintf(ctx, "Hello from k8s-controller!"); err != nil {
				log.Error().Err(err).Msg("Failed to write response")
			}
		}
	}

	return fasthttp.ListenAndServe(addr, handler)
}
