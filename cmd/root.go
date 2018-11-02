package cmd

import (
	"fmt"
	"os"

	"github.com/revan730/gamedev-backend/src"
	"github.com/spf13/cobra"
)

var (
	logVerbose bool
	serverPort int
	dbAddr     string
	db         string
	dbUser     string
	dbPass     string
	redisAddr  string
	redisPass  string
)

var RootCmd = &cobra.Command{
	Use:   "gamedev-backend",
	Short: "Backend for gamedev project",
}

var serveCmd = &cobra.Command{
	Use:   "start",
	Short: "Start server",
	Run: func(cmd *cobra.Command, args []string) {
		config := &src.Config{
			Port:          serverPort,
			DBAddr:        dbAddr,
			DB:            db,
			DBUser:        dbUser,
			DBPassword:    dbPass,
			RedisAddr:     redisAddr,
			RedisPassword: redisPass,
		}
		logger := src.NewLogger(logVerbose)
		server := src.NewServer(logger, config).Routes()
		server.Run()
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVarP(&serverPort, "port", "p", 8080,
		"Application TCP port")
	serveCmd.Flags().StringVarP(&dbAddr, "postgresAddr", "a",
		"postgres:5432", "Set PostsgreSQL address")
	serveCmd.Flags().StringVarP(&db, "db", "d",
		"fict", "Set PostgreSQL database to use")
	serveCmd.Flags().StringVarP(&dbUser, "user", "u",
		"fict", "Set PostgreSQL user to use")
	serveCmd.Flags().StringVarP(&dbPass, "pass", "c",
		"fict", "Set PostgreSQL password to use")
	serveCmd.Flags().StringVarP(&redisAddr, "redis", "r",
		"redis:6379", "Set redis address")
	serveCmd.Flags().StringVarP(&redisPass, "redispass", "b",
		"", "Set redis address")
}
