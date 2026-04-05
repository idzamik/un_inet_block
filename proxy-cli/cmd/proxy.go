package cmd

import (
	"github.com/spf13/cobra"
)

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Manage proxy operations",
	Long: `proxy — группа команд для управления локальным прокси-процессом.

В этом разделе будут команды для:
- запуска локального SOCKS5/HTTP прокси
- остановки процесса
- проверки текущего статуса
- выбора серверного профиля

Пока подкоманды являются заготовками, но help уже можно сделать полноценным.`,
}

var proxyStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start proxy",
	Run:   func(cmd *cobra.Command, args []string) {},
}

var proxyStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop proxy",
	Run:   func(cmd *cobra.Command, args []string) {},
}

var proxyStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Proxy status",
	Run:   func(cmd *cobra.Command, args []string) {},
}

func init() {
	proxyStartCmd.Flags().Int("socks-port", 1080, "SOCKS5 listen port")
	proxyStartCmd.Flags().Int("http-port", 8080, "HTTP listen port")

	proxyCmd.AddCommand(proxyStartCmd, proxyStopCmd, proxyStatusCmd)
	rootCmd.AddCommand(proxyCmd)
}
