package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/joshwi/go-pkg/logger"
	"github.com/joshwi/go-pkg/utils"
	"github.com/joshwi/go-svc/db"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

var (
	// Pull in env variables: USERNAME, PASSWORD, uri
	DIRECTORY = os.Getenv("TARGET")
	LOGFILE   = os.Getenv("LOGFILE")
	USERNAME  = os.Getenv("NEO4J_USERNAME")
	PASSWORD  = os.Getenv("NEO4J_PASSWORD")
	HOST      = os.Getenv("NEO4J_SERVICE_HOST")
	PORT      = os.Getenv("NEO4J_SERVICE_PORT")

	// Init flag values
	query string
)

func init() {

	// Define flag arguments for the application
	flag.StringVar(&query, `q`, `n.filepath=~'.*/tmp.*'`, `Run query to DB for input parameters. Default: <empty>`)
	flag.Parse()

	// Initialize LOGFILE at user given path. Default: ./collection.log
	logger.InitLog(LOGFILE)

	logger.Logger.Info().Str("query", query).Str("status", "start").Msg("FILE PHOTOS")
}

var a2 = regexp.MustCompile(`\/\w+\/[a-zA-Z0-9]{3}\/\d{4}\/\d{4}\-\d{1,2}\-\d{1,2}(\/\w+|)\/\w+\.[a-zA-Z0-9]{3,4}`)

func main() {

	// var a0 = regexp.MustCompile(`[^a-zA-Z\d\/]+`)
	// var a1 = regexp.MustCompile(`\_{2,}`)

	// Create application session with Neo4j
	uri := "bolt://" + HOST + ":" + PORT
	driver := db.Connect(uri, USERNAME, PASSWORD)
	sessionConfig := neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}
	session := driver.NewSession(sessionConfig)

	songs, _ := db.GetNode(session, "photo", query, 0, []string{"label", "filepath", "filename", "filetype", "create_date", "category"})

	for _, entry := range songs {

		log.Println(entry)

		if entry["label"] != "" && entry["filepath"] != "" && entry["filename"] != "" && entry["filetype"] != "" && entry["create_date"] != "" && entry["category"] != "" {

			date, _ := time.Parse("2006:01:02 15:04:05", entry["create_date"])
			year := date.Year()
			format_date := date.Format("2006-01-02")

			expected_filepath := fmt.Sprintf("/%v/%v/%v/%v/", entry["category"], strings.ToUpper(entry["filetype"]), year, format_date)
			expected_filepath = utils.FormatPath(expected_filepath)

			pano_flag := strings.Contains(entry["filepath"], "/PANORAMA/")

			if pano_flag == true {
				dirs := strings.Split(entry["filepath"], "/")
				if len(dirs) > 1 {
					expected_filepath = fmt.Sprintf("%v%v/%v", expected_filepath, dirs[len(dirs)-2], dirs[len(dirs)-1])
				}
			} else {
				expected_filepath += entry["filename"]
			}

			match := a2.FindAllString(expected_filepath, -1)

			if entry["filepath"] != expected_filepath && len(match) > 0 {

				_, err := os.Stat(DIRECTORY + expected_filepath)
				if os.IsNotExist(err) {
					// log.Println(DIRECTORY+entry["filepath"], DIRECTORY+expected_filepath)
					err := utils.Move(DIRECTORY+entry["filepath"], DIRECTORY+expected_filepath)
					if err != nil {
						log.Fatal(err)
						os.Exit(1)
					} else {
						db.PutNode(session, "photo", entry["label"], []utils.Tag{{Name: "filepath", Value: expected_filepath}})
					}
				}
			}
		}
	}

	logger.Logger.Info().Str("query", query).Str("status", "end").Msg("FILE PHOTOS")

}
