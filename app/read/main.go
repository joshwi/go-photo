package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/joshwi/go-pkg/logger"
	"github.com/joshwi/go-pkg/utils"
	"github.com/joshwi/go-svc/db"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

var (
	DIRECTORY = os.Getenv("TARGET")
	USERNAME  = os.Getenv("NEO4J_USERNAME")
	PASSWORD  = os.Getenv("NEO4J_PASSWORD")
	HOST      = os.Getenv("NEO4J_SERVICE_HOST")
	PORT      = os.Getenv("NEO4J_SERVICE_PORT")
	LOGFILE   = os.Getenv("LOGFILE")
	types     = map[string]string{
		"title":    "TIT2",
		"album":    "TALB",
		"artist":   "TPE1",
		"genre":    "TCON",
		"producer": "TCOM",
		"track":    "TRCK",
		"year":     "TYER",
		"comments": "COMM",
		"lyrics":   "USLT",
	}
	// Init flag values
	query     string
	filetypes string
	search    string
)

func init() {

	// Define flag arguments for the application
	flag.StringVar(&query, `q`, ``, `Run query to DB for input parameters. Default: <empty>`)
	flag.StringVar(&search, `d`, `tmp`, `Directory to search for. Default: tmp`)
	flag.StringVar(&filetypes, `f`, `cr2|jpg|png|mp4`, `Filetypes to consider. Default: cr2|jpg|png|mp4`)
	flag.Parse()

	// Initialize logfile at user given path. Default: ./collection.log
	logger.InitLog(LOGFILE)

	logger.Logger.Info().Str("status", "start").Msg("READ PHOTOS")
}

func ReadMetadata(directory string, filename string) ([]utils.Tag, string, error) {

	output := map[string]string{}

	exif, err := exiftool.NewExiftool()
	if err != nil {
		fmt.Printf("Error when intializing: %v\n", err)
		return nil, ``, err
	}
	defer exif.Close()

	data := exif.ExtractMetadata(directory + filename)

	label := ``

	for _, info := range data {

		if info.Err != nil {
			return nil, ``, info.Err

		}

		for k, v := range info.Fields {
			if k == "FileName" {
				label = fmt.Sprintf("%v", v)
			}
			// log.Println(fmt.Sprintf("[%v] %v", k, v))
			output[k] = fmt.Sprintf("%v", v)
		}
	}

	props := []utils.Tag{
		{Name: "model", Value: output["Model"]},
		{Name: "lens", Value: output["Lens"]},
		{Name: "lens_model", Value: output["LensModel"]},
		{Name: "focal_length", Value: output["FocalLength"]},
		{Name: "file_size", Value: output["FileSize"]},
		{Name: "file_number", Value: output["FileNumber"]},
		{Name: "filename", Value: output["FileName"]},
		{Name: "filetype", Value: output["FileTypeExtension"]},
		{Name: "iso", Value: output["ISO"]},
		{Name: "shutter_speed", Value: output["ShutterSpeed"]},
		{Name: "aperture", Value: output["Aperture"]},
		{Name: "megapixels", Value: output["Megapixels"]},
		{Name: "create_date", Value: output["CreateDate"]},
		{Name: "modify_date", Value: output["ModifyDate"]},
		{Name: "filepath", Value: filename},
	}

	return props, label, nil
}

func main() {

	// Create application session with Neo4j
	uri := "bolt://" + HOST + ":" + PORT
	driver := db.Connect(uri, USERNAME, PASSWORD)

	filetree, err := utils.Scan(DIRECTORY)
	if err != nil {
		log.Fatal(err)
	}

	var REG_FILE_MATCH = regexp.MustCompile(fmt.Sprintf(`(?i)\.(%v)$`, filetypes))
	var REG_FOLDER_MATCH = regexp.MustCompile(fmt.Sprintf(`%v`, search))

	files := []string{}

	for _, item := range filetree {
		info, _ := os.Stat(DIRECTORY + item)
		file_match := REG_FILE_MATCH.FindAllString(item, -1)
		folder_match := REG_FOLDER_MATCH.FindAllString(item, -1)
		if !info.IsDir() && len(file_match) > 0 && len(folder_match) > 0 {
			files = append(files, item)
		}
	}

	start := time.Now()

	queue := make(chan string, 100)
	results := make(chan error)

	for i := 0; i < cap(queue); i++ {
		go worker(driver, queue, results)
	}

	go func() {
		for _, entry := range files {
			queue <- entry
		}
	}()

	pass := 0

	for range files {
		success := <-results
		if success == nil {
			pass++
		}
	}

	close(queue)
	close(results)

	end := time.Now()
	elapsed := end.Sub(start)
	duration := fmt.Sprintf("%v", elapsed.Round(time.Second/1000))

	avg := "0 ms/file"
	percent := 100.0

	if pass > 0 {
		avg = fmt.Sprintf("%v ms/file", int(elapsed.Milliseconds())/len(files))
		percent = (float64(pass) / float64(len(files))) * 100.0
	}

	success := fmt.Sprintf("%v%%", math.Round(percent*100)/100)

	logger.Logger.Info().Str("duration", duration).Str("success", success).Str("speed", avg).Int("completed", pass).Int("total", len(files)).Msg("Read Photos")

	logger.Logger.Info().Str("status", "end").Msg("READ PHOTOS")
}

func worker(driver neo4j.Driver, queue chan string, results chan error) {
	sessionConfig := neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}
	session := driver.NewSession(sessionConfig)
	for entry := range queue {
		tags, label, err := ReadMetadata(DIRECTORY, entry)
		if err != nil {
			results <- err
		}
		err = db.PutNode(session, "photo", label, tags)
		if err != nil {
			results <- err
		}
		results <- nil
	}
	session.Close()
}
