package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "proxy",
	Short: "CLI for proxy automation",
	Long: `proxy-cli — консольное приложение для управления прокси-подключениями.

Сейчас приложение находится на стадии подготовки архитектуры:
- хранение серверов в configs/servers.txt
- запуск proxy-команд через Cobra
- дальнейшая интеграция с Xray/VLESS

Пока реализован каркас CLI и читаемый help-вывод.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
