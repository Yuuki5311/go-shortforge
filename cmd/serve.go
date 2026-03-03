package cmd

import (
	"os"

	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/server/egin"
	"github.com/spf13/cobra"

	"shorturl/internal/app"
)

var (
	cfgFile string
)

// serve 子命令：
// - 初始化并启动 Ego 应用
// - 构建 HTTP 服务器组件，注册依赖初始化 Invoker，最终运行
// - 支持通过 --config 指定配置文件路径（默认 config.toml）
func init() {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "启动API服务",
		RunE: func(cmd *cobra.Command, args []string) error {
			commercialization()
			e := ego.New()
			server := app.NewHTTPServer()
			if err := e.
				Serve(server).
				Invoker(app.InitComponents).
				Run(); err != nil {
				elog.Panic("startup", elog.FieldErr(err))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&cfgFile, "config", "config.toml", "配置文件路径")
	rootCmd.AddCommand(cmd)
}

// commercialization：占位函数，确保编译期保留 egin 组件引用（避免仅在装配中引用被 go mod tidy 移除）
func commercialization() {
	_ = &egin.Component{}
	_ = os.ErrInvalid
}
