package cmd

import (
	"fmt"

	"github.com/idzamik/proxy-cli/xray"
	"github.com/spf13/cobra"
)

var (
	confInstall bool
	confUpdate  bool
	confDelete  bool
)

var xrayConfCmd = &cobra.Command{
	Use:   "conf [source]",
	Short: "Установить, обновить или удалить список серверов",
	Long: `Работа со списком серверов для Xray.

source опционален:
- если не указан, используется захардкоженный URL
- если указан http(s) URL, список скачивается по сети
- если указан локальный путь, список читается из файла`,
	Args: cobra.MaximumNArgs(1),
	Example: `  proxy-cli proxy xray conf -i
  proxy-cli proxy xray conf -u
  proxy-cli proxy xray conf -i https://example.com/list.txt
  proxy-cli proxy xray conf -u /home/user/Downloads/list.txt
  proxy-cli proxy xray conf -d`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		count := 0
		if confInstall {
			count++
		}
		if confUpdate {
			count++
		}
		if confDelete {
			count++
		}

		if count != 1 {
			return fmt.Errorf("use exactly one flag: -i, -u or -d")
		}

		if confDelete && len(args) > 0 {
			return fmt.Errorf("source is not used with -d")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		core := xray.NewLists(
			"./configs/servers.txt",
			"https://raw.githubusercontent.com/igareck/vpn-configs-for-russia/refs/heads/main/Vless-Reality-White-Lists-Rus-Mobile.txt",
		)

		source := ""
		if len(args) == 1 {
			source = args[0]
		}

		var (
			msg string
			err error
		)

		switch {
		case confInstall:
			msg, err = core.Install(source)
		case confUpdate:
			msg, err = core.Update(source)
		case confDelete:
			msg, err = core.Delete()
		}

		if err != nil {
			return err
		}

		fmt.Println(msg)
		return nil
	},
}

func init() {
	xrayConfCmd.Flags().BoolVarP(&confInstall, "install", "i", false, "install config list")
	xrayConfCmd.Flags().BoolVarP(&confUpdate, "update", "u", false, "update config list")
	xrayConfCmd.Flags().BoolVarP(&confDelete, "delete", "d", false, "delete config list")

	xrayCmd.AddCommand(xrayConfCmd)
}
