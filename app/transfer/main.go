package main

import (
	"flag"
	"log"
	"os"

	"github.com/joshwi/go-pkg/logger"
	"github.com/joshwi/go-pkg/utils"
)

var (
	LOGFILE   = os.Getenv("LOGFILE")
	subfolder = "/tmp"

	extensions string
	source     string
	target     string
)

func init() {

	// Define flag arguments for the application
	flag.StringVar(&extensions, `files`, ``, `File extensions to transfer. Default: empty`)
	flag.StringVar(&source, `source`, ``, `Source directory. Default: empty`)
	flag.StringVar(&target, `target`, ``, `Target directory. Default: empty`)
	flag.Parse()

	// Initialize logfile at user given path. Default: ./collection.log
	logger.InitLog(LOGFILE)

	logger.Logger.Info().Str("status", "start").Msg("TRANSFER")
}

func main() {

	pass, tot := utils.Backup(source, target, extensions, subfolder)
	if pass != tot {
		diff := tot - pass
		log.Fatalf("%v files failed to copy!", diff)
		os.Exit(1)
	}

	logger.Logger.Info().Str("status", "end").Msg("TRANSFER")

}
