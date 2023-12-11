package saver

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type saverSuite struct {
	suite.Suite
}

func TestSaverSuite(t *testing.T) {
	suite.Run(t, new(saverSuite))
}
func (s *saverSuite) TestNewSaver() {

	sep := runtimeops.GetSep()

	// test no folder no file
	_, err := os.Stat("logs")
	if err != os.ErrNotExist {
		os.RemoveAll("logs")
	}

	sv := NewSaver("logs", sep, int64(100))

	s.Equal("logs", sv.PathMap["all"])
	s.Equal("logs", sv.PathMap["err"])
	s.Equal("logs", sv.PathMap["sig"])

	s.Equal(int64(100), sv.Limit)

	_, err = os.Stat("logs")
	s.NoError(err)

	// test folder but no file
	os.RemoveAll("logs")

	folderName := "logs"
	os.Mkdir(folderName, 0777)
	fileName := "2023-11-23_all.log"
	filePath := filepath.Join(folderName, fileName)
	f, err := os.Create(filePath)
	if err != nil {
		log.Printf("in saver.TestNewSaver unable to create file %s:%v", fileName, err)
	}
	f.Close()
	sv = NewSaver(folderName, sep, int64(100))

	s.Equal("logs"+sep+"2023-11-23_all.log", sv.PathMap["all"])
	s.Equal("logs", sv.PathMap["err"])
	s.Equal("logs", sv.PathMap["sig"])

	// test folder and different files

	remFolderRecursively(folderName)
	os.Mkdir(folderName, 0777)
	initFiles := map[string]int{
		"23.11.01T15.15.15.000_all.txt": 222,
	}
	s.NoError(createAndFill("logs", sep, initFiles))

	sv = NewSaver(folderName, sep, int64(100))
	s.Equal("logs", sv.PathMap["all"])
	s.Equal("logs", sv.PathMap["err"])
	s.Equal("logs", sv.PathMap["sig"])

	remFolderRecursively(folderName)
}

func (s *saverSuite) TestGetFile() {

	time := time.Now()
	ts := time.Format("2006-01-02")
	u := uuid.New()
	sep := runtimeops.GetSep()

	tt := []struct {
		name        string
		folderPath  string
		filePath    []string
		limit       int64
		currentSize int
		initFiles   map[string]int
		wl          model.WrappedLog
		wantPathUpd map[string]string
		wantError   error
	}{

		{
			name:       "New first log file",
			folderPath: "logs",
			initFiles: map[string]int{
				"azaza.log": 44,
			},
			limit: int64(100),
			wl:    model.WrappedLog{T: time, UW: model.UUIDWrapper{Str: "", UUID: u}, L: "logloglog"},
			wantPathUpd: map[string]string{
				"all": "logs" + sep + ts + "_all" + ".log",
				"err": "logs",
				"sig": "logs",
			},
			wantError: nil,
		},

		{
			name:        "Current log file",
			folderPath:  "logs",
			filePath:    []string{"logs" + sep + ts + "_all" + ".log"},
			limit:       int64(100),
			wl:          model.WrappedLog{T: time, UW: model.UUIDWrapper{Str: "", UUID: u}, L: "logloglog"},
			currentSize: 20,
			wantPathUpd: map[string]string{
				"all": "logs" + sep + ts + "_all" + ".log",
				"err": "logs",
				"sig": "logs",
			},
			wantError: nil,
		},

		{
			name:        "New second log file, size == limit",
			folderPath:  "logs",
			filePath:    []string{"logs" + sep + ts + "_all" + ".log"},
			limit:       int64(100),
			wl:          model.WrappedLog{T: time, UW: model.UUIDWrapper{Str: "", UUID: u}, L: "logloglog"},
			currentSize: 100,
			wantPathUpd: map[string]string{
				"all": "logs" + sep + ts + "_all" + ".log",
				"err": "logs",
				"sig": "logs",
			},
			wantError: nil,
		},

		{
			name:        "New second log file, size + logString > linit",
			folderPath:  "logs" + sep + "06.10.2023-12_30_00.log",
			limit:       int64(100),
			currentSize: 90,
			wl:          model.WrappedLog{T: time, UW: model.UUIDWrapper{Str: "", UUID: u}, L: "loglogloglogloglog"},
			wantPathUpd: map[string]string{
				"all": "logs" + sep + ts + "_all" + ".log",
				"err": "logs",
				"sig": "logs",
			},
			wantError: nil,
		},
		{
			name:        "Any other file",
			folderPath:  "logs" + sep + "06.10.2023T12_30_00.log",
			limit:       int64(100),
			currentSize: 90,
			wl:          model.WrappedLog{T: time, UW: model.UUIDWrapper{Str: "", UUID: u}, L: "loglogloglogloglog"},
			wantPathUpd: map[string]string{
				"all": "logs" + sep + ts + "_all" + ".log",
				"err": "logs",
				"sig": "logs",
			},
			wantError: nil,
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			remFolderRecursively("logs")

			sv := NewSaver(v.folderPath, sep, v.limit)

			if strings.Contains(v.folderPath, ".") && len(v.folderPath) == 0 {
				v.filePath = append(v.filePath, v.folderPath)
			}

			// Creating and filling initial file
			if len(v.filePath) > 0 {
				for _, w := range v.filePath {
					f0, err := os.Create(w)
					if err != nil {
						log.Printf("in saver.TestGetFile error %v\n", err)
					}
					f0.WriteString(strings.Repeat("x", v.currentSize))
					f0.Close()
					sv = NewSaver(w, sep, v.limit)
				}
			}

			files, pathUpd, err := sv.getFile(v.wl, v.limit)
			s.NoError(err)

			for j, w := range v.wantPathUpd {
				s.Equal(w, pathUpd[j])
			}

			// Deleting logs recursively
			for _, v := range files {
				v.Close()
			}

			remFolderRecursively("logs")
		})
	}
}

func (s *saverSuite) TestSave() {
	time := time.Now()
	//ts := time.Format("2006-01-02 15:04:05")
	u := uuid.New()
	sep := runtimeops.GetSep()
	folderName := "logs"
	err := os.Mkdir(folderName, 0777)
	if err != nil && !errors.Is(err, os.ErrExist) {
		log.Fatalf("in saver.TestSave unable to create logs directory: %v\n", err)
	}

	tt := []struct {
		name        string
		initFiles   map[string]int
		log         model.WrappedLog
		limit       int64
		wantContent map[string]string
	}{
		{
			name: "To all only, size < limit",
			initFiles: map[string]int{
				"23.11.01T12.30.00.000_all.log": 30,
				"23.11.01T12.30.00.000_err.log": 30,
			},
			log: model.WrappedLog{
				T: time,
				UW: model.UUIDWrapper{
					UUID: u,
					Str:  "",
				},
				L: "logloglog",
			},
			limit: int64(100),
			wantContent: map[string]string{
				//"23.11.01T12.30.00.000_all.log": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\r\n" + fmt.Sprintf("%s: %s %s\r\n", ts, u.String(), "logloglog"),
				"23.11.01T12.30.00.000_all.log": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\r\n",
				"23.11.01T12.30.00.000_err.log": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\r\n",
			},
		},

		{
			name: "To all and err, size < limit",
			initFiles: map[string]int{
				"23.11.01T12.30.00.000_all.log": 30,
				"23.11.01T12.30.00.000_err.log": 30,
			},
			log: model.WrappedLog{
				T: time,
				UW: model.UUIDWrapper{
					UUID: u,
					Str:  "ERROR",
				},
				L: "logloglog",
			},
			limit: int64(100),
			wantContent: map[string]string{
				//"23.11.01T12.30.00.000_all.log": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\r\n" + fmt.Sprintf("%s: %s %s %s\r\n", ts, "ERROR", u.String(), "logloglog"),
				"23.11.01T12.30.00.000_all.log": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\r\n",
				//"23.11.01T12.30.00.000_err.log": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\r\n" + fmt.Sprintf("%s: %s %s %s\r\n", ts, "ERROR", u.String(), "logloglog"),
				"23.11.01T12.30.00.000_err.log": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\r\n",
			},
		},

		{
			name: "To all and sig, size < limit",
			initFiles: map[string]int{
				"23.11.01T12.30.00.000_all.log": 30,
				"23.11.01T12.30.00.000_err.log": 30,
				"23.11.01T12.30.00.000_sig.log": 30,
			},
			log: model.WrappedLog{
				T: time,
				UW: model.UUIDWrapper{
					UUID: u,
					Str:  "SIGNAL",
				},
				L: "logloglog",
			},
			limit: int64(100),
			wantContent: map[string]string{
				//"23.11.01T12.30.00.000_all.log": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\r\n" + fmt.Sprintf("%s: %s %s\r\n", ts, "SIGNAL", "logloglog"),
				"23.11.01T12.30.00.000_all.log": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\r\n",
				"23.11.01T12.30.00.000_err.log": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\r\n",
				//"23.11.01T12.30.00.000_sig.log": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\r\n" + fmt.Sprintf("%s: %s %s\r\n", ts, "SIGNAL", "logloglog"),
				"23.11.01T12.30.00.000_sig.log": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\r\n",
			},
		},
	}

	for _, v := range tt {
		s.Run(v.name, func() {
			// creating files
			s.NoError(createAndFill(folderName, sep, v.initFiles))

			sv := NewSaver(folderName, sep, v.limit)

			sv.Save(v.log)

			// check files content
			for j, w := range v.wantContent {
				content, err := readFile(folderName, j, sep)
				s.NoError(err)
				s.Equal(w, content)
			}

			// deleting files
			cleanFolder(folderName, sep)
		})
	}
	remFolderRecursively(folderName)
}

func remFolderRecursively(path string) {
	_, err := os.Stat(path)
	if err == nil || err != os.ErrNotExist {
		err = os.RemoveAll(path)
		if err != nil {
			log.Println(err)
		}
	}
}
func createAndFill(folder, sep string, files map[string]int) error {

	for i, v := range files {
		f, err := os.Create(filepath.Join(folder + sep + i))
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.WriteString(strings.Repeat("x", v) + "\r\n")
		if err != nil {
			return err
		}
	}
	return nil
}

func cleanFolder(path, sep string) error {
	_, err := os.Stat(path)
	if err != nil {
		return err
	}
	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, v := range files {
		name := v.Name()
		err = os.Remove(filepath.Join(path + sep + name))
		if err != nil {
			return err
		}
	}
	return nil
}
func readFile(folder, file, sep string) (string, error) {

	f, err := os.Open(folder + sep + file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	body, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
