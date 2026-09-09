package main

import (
	"context"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/Ursa-Minor-Beta/baas-client/internal/build"
	"github.com/Ursa-Minor-Beta/baas-client/pkg/client"
	"github.com/Ursa-Minor-Beta/baas-client/pkg/client/dto"
	"github.com/Ursa-Minor-Beta/baas-client/pkg/util"
)

func main() {
	var cfg client.Config
	cfg.Url = "http://localhost:8090"
	if os.Getenv("BAAS_URL") != "" {
		cfg.Url = os.Getenv("BAAS_URL")
	}
	cfg.ApiKey = "test"
	if os.Getenv("BAAS_API_KEY") != "" {
		cfg.ApiKey = os.Getenv("BAAS_API_KEY")
	}
	cfg.LocalDebug = false
	cfg.UseProxy = true

	var cookiesSlice []string
	var cookieDomain string
	rootCmd := &cobra.Command{
		Use:     "baas",
		Version: build.Version,
		Short:   "BaaS is a Browser as a Service",
		Long:    "Easy way to control chrome browser within AWS Lambda",
		Run: func(cmd *cobra.Command, args []string) {
			for k, v := range util.SliceToMap(cookiesSlice) {
				cfg.Cookies = append(cfg.Cookies, dto.BrowserCookie{
					Name:   k,
					Value:  v,
					Domain: cookieDomain,
					Path:   "/",
				})
			}
			startBaasClient(cfg)
		},
	}
	rootCmd.PersistentFlags().StringVarP(&cfg.Url, "url", "u", cfg.Url, "BaaS backend URL")
	rootCmd.PersistentFlags().StringVarP(&cfg.ApiKey, "key", "k", cfg.ApiKey, "BaaS API Key")
	rootCmd.PersistentFlags().BoolVarP(&cfg.LocalDebug, "debug", "d", cfg.LocalDebug, "Local debug")
	rootCmd.PersistentFlags().BoolVarP(&cfg.UseProxy, "proxy", "p", cfg.UseProxy, "Use proxy")
	rootCmd.PersistentFlags().StringVarP(&cfg.Timeout, "timeout", "t", "10m", "Max session length (duration, e.g. 10m), default: 800s")
	rootCmd.PersistentFlags().StringVarP(&cfg.MessageTimeout, "message-timeout", "M", "30s", "Max time to wait for each message, default: 30s")
	rootCmd.PersistentFlags().StringSliceVarP(&cfg.Secrets, "secret", "S", []string{}, "Secrets to send to backend with each async request")
	rootCmd.PersistentFlags().StringSliceVarP(&cfg.Values, "value", "V", []string{}, "Values to send to backend with each async request")
	rootCmd.PersistentFlags().StringSliceVarP(&cookiesSlice, "cookie", "C", []string{}, "Cookies to send to backend with each async request")
	rootCmd.PersistentFlags().StringVarP(&cookieDomain, "cookie-domain", "D", "", "Cookies domain to set with cookies backend with each async request")

	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}

func startBaasClient(cfg client.Config) {
	client, err := client.BubbleClient(context.Background(), cfg)
	if err != nil {
		panic(err)
	}
	p := tea.NewProgram(client)

	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
