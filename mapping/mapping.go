package mapping

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cast"
)

var mapping map[string]map[string]float64

func Init(dir string) {
	mapping = make(map[string]map[string]float64)

	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	for _, info := range infos {
		if info.IsDir() || filepath.Ext(info.Name()) != ".csv" {
			continue
		}

		feature := path.Base(strings.TrimRight(info.Name(), filepath.Ext(info.Name())))
		if _, ok := mapping[feature]; !ok {
			mapping[feature] = make(map[string]float64)
		}

		longFilename := filepath.Join(dir, info.Name())
		file, err := os.Open(longFilename)
		if err != nil {
			panic(file)
		}
		defer file.Close()

		csvr := csv.NewReader(file)
		_, _ = csvr.Read() // first line
		for {
			row, err := csvr.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				panic(err)
			}

			if len(row) != 2 {
				panic(fmt.Sprintf("not equal 2, values is %v", row))
			}
			k, v := row[0], row[1]
			mapping[feature][k] = cast.ToFloat64(v)
		}
	}
}

func GetFeatureMapping(featurename string, value string) float64 {
	if _, ok := mapping[featurename]; !ok {
		return cast.ToFloat64(value)
	}

	return cast.ToFloat64(mapping[featurename][value])
}

func GetMapping() map[string]map[string]float64 {
	return deepCopy(mapping)
}

func deepCopy(src map[string]map[string]float64) map[string]map[string]float64 {
	dst := make(map[string]map[string]float64)

	for k, m := range src {
		dst[k] = make(map[string]float64)
		for kk, vv := range m {
			dst[k][kk] = vv
		}
	}
	return dst
}
