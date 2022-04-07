package main

import (
	"context"
	"fmt"

	"net/http"
	"os"
	"os/signal"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mect/eid-rest/pkg/eidenv"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(NewServeCmd())
}

type serveCmdOptions struct {
	BindAddr  string
	Port      int
	AuthToken string
}

// NewServeCmd generates the `serve` command
func NewServeCmd() *cobra.Command {
	s := serveCmdOptions{}
	c := &cobra.Command{
		Use:     "serve",
		Short:   "Serves the HTTP REST endpoint",
		Long:    `Serves the HTTP REST endpoint on the given bind address and port`,
		PreRunE: s.Validate,
		RunE:    s.RunE,
	}
	c.Flags().StringVarP(&s.BindAddr, "bind-address", "b", "0.0.0.0", "address to bind port to")
	c.Flags().IntVarP(&s.Port, "port", "p", 8080, "Port to listen on")
	c.Flags().StringVarP(&s.AuthToken, "auth-token", "t", "", "Authentication token")

	c.MarkFlagRequired("auth-token")

	return c
}

func (s *serveCmdOptions) Validate(cmd *cobra.Command, args []string) error {

	return nil
}

func (s *serveCmdOptions) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// handlers
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "EID API endpoint")
	})

	e.GET("/read", func(c echo.Context) error {
		if c.QueryParam("token") != s.AuthToken {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token"})
		}
		e, err := eidenv.New()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		}

		info, err := e.ReadCard()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, info)
	})

	go func() {
		e.Start(fmt.Sprintf("%s:%d", s.BindAddr, s.Port))
		cancel() // server ended, stop the world
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
			return nil
		}
	}
}
