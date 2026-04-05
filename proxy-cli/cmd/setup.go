package cmd

import (
	"fmt"

	"github.com/idzamik/proxy-cli/xray"

	"github.com/spf13/cobra"
)

var xrayCmd = &cobra.Command{
	Use:   "xray",
	Short: "Управление локальным xray-core",
	Long: `Команды для подготовки локального xray-core.

proxy xray setup
  - скачивает latest release xray-core

proxy xray setup /path/to/file
  - устанавливает xray-core из локального ZIP-архива
    или из уже готового бинарного файла

proxy xray remove
  - удаляет локально установленный xray-core`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var xraySetupCmd = &cobra.Command{
	Use:   "setup [path-to-local-file]",
	Short: "Скачать или установить xray-core",
	Long: `Устанавливает xray-core в ./bin/xray.

Если путь не передан:
- скачивает latest release из GitHub Releases

Если путь передан:
- принимает локальный .zip архив
- либо локальный исполняемый файл xray`,
	Args: cobra.MaximumNArgs(1),
	Example: `  proxy-cli proxy xray setup
  proxy-cli proxy xray setup ./downloads/Xray-linux-64.zip
  proxy-cli proxy xray setup ./downloads/xray`,
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := xray.XrayManager("./bin")

		localPath := ""
		if len(args) == 1 {
			localPath = args[0]
		}

		msg, err := manager.Setup(localPath)
		if err != nil {
			return err
		}

		fmt.Println(msg)
		return nil
	},
}

var xrayRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Удалить локальный xray-core",
	Long:  "Удаляет локальный бинарник xray-core из ./bin/xray.",
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := xray.XrayManager("./bin")
		msg, err := manager.Remove()
		if err != nil {
			return err
		}

		fmt.Println(msg)
		return nil
	},
}

func init() {
	xrayCmd.AddCommand(xraySetupCmd, xrayRemoveCmd)
	proxyCmd.AddCommand(xrayCmd)
}
