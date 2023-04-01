package main

import (
	"archive/zip"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/ingest/compact"
	"github.com/golang/geo/s2"
	"github.com/lukeroth/gdal"
)

const (
	PostcodeColumn  = "Postcode"
	EastingsColumn  = "Eastings"
	NorthingsColumn = "Northings"

	EPSGCodeBritishNationalGrid = 27700
	EPSGCodeWGS84               = 4326
)

func readColumnHeaders(z *zip.Reader) (map[string]int, error) {
	r, err := z.Open("Doc/Code-Point_Open_Column_Headers.csv")
	if err != nil {
		return nil, err
	}
	defer r.Close()
	c := csv.NewReader(r)
	columns := make(map[string]int)
	for {
		row, err := c.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		for i, column := range row {
			columns[column] = i
		}
	}
	return columns, nil
}

type Postcodes struct {
	Postcode []string
	Lat      []float64
	Lng      []float64
}

func (p *Postcodes) Read(options ingest.ReadOptions, emit ingest.Emit, ctx context.Context) error {
	if options.SkipPoints {
		return nil
	}
	point := ingest.PointFeature{
		Tags: []b6.Tag{{Key: "#place", Value: "postal_code"}},
	}
	for i := range p.Postcode {
		point.PointID = b6.PointIDFromGBPostcode(p.Postcode[i])
		if point.PointID == b6.PointIDInvalid {
			return fmt.Errorf("invalid postcode: %q", p.Postcode[i])
		}
		point.Location = s2.LatLngFromDegrees(p.Lat[i], p.Lng[i])
		if err := emit(&point, 0); err != nil {
			return err
		}
	}
	return nil
}

func readPostcodeCoordinates(z *zip.Reader, columns map[string]int, postcodes *Postcodes) error {
	skipped := 0
	for _, ff := range z.File {
		if strings.HasPrefix(ff.Name, "Data/CSV/") && strings.HasSuffix(ff.Name, ".csv") {
			fr, err := ff.Open()
			if err != nil {
				return err
			}
			cr := csv.NewReader(fr)
			cr.Comment = '#'
			cr.FieldsPerRecord = -1
			for {
				row, err := cr.Read()
				if err == io.EOF {
					break
				} else if err != nil {
					return err
				}
				postcodes.Postcode = append(postcodes.Postcode, row[columns[PostcodeColumn]])
				e, err := strconv.Atoi(row[columns[EastingsColumn]])
				if err != nil {
					skipped++
					continue
				}
				n, err := strconv.Atoi(row[columns[NorthingsColumn]])
				if err != nil {
					skipped++
					continue
				}
				postcodes.Lat = append(postcodes.Lat, float64(e))
				postcodes.Lng = append(postcodes.Lng, float64(n))
			}
		}
	}
	if skipped > 0 {
		log.Printf("skipped %d postcodes due to bad locations", skipped)
	}
	return nil
}

func transformPostcodeCoordinates(postcodes *Postcodes) error {
	from := gdal.CreateSpatialReference("")
	from.FromEPSG(int(EPSGCodeBritishNationalGrid))
	to := gdal.CreateSpatialReference("")
	to.FromEPSG(int(EPSGCodeWGS84))
	zs := make([]float64, len(postcodes.Lat))
	transform := gdal.CreateCoordinateTransform(from, to)
	transform.Transform(len(postcodes.Lat), postcodes.Lat, postcodes.Lng, zs)
	return nil
}

func readCodepointOpen(filename string) (Postcodes, error) {
	postcodes := Postcodes{}
	s, err := os.Stat(filename)
	if err != nil {
		return postcodes, fmt.Errorf("can't stat codepoint data: %s", err)
	}
	f, err := os.Open(filename)
	if err != nil {
		return postcodes, fmt.Errorf("can't open codepoint data: %s", err)
	}
	defer f.Close()
	z, err := zip.NewReader(f, s.Size())
	if err != nil {
		return postcodes, fmt.Errorf("can't read codepoint zip %s: %s", filename, err)
	}
	columns, err := readColumnHeaders(z)
	if err != nil {
		return postcodes, fmt.Errorf("can't read column headers from %s: %s", filename, err)
	}
	err = readPostcodeCoordinates(z, columns, &postcodes)
	if err != nil {
		return postcodes, fmt.Errorf("can't read codepoint data from: %s: %s", filename, err)
	}
	err = transformPostcodeCoordinates(&postcodes)
	if err != nil {
		return postcodes, err
	}
	return postcodes, nil
}

func main() {
	inputFlag := flag.String("input", "", "Input OS Code-Point Open data, zipped")
	outputFlag := flag.String("output", "", "Output index")
	coresFlag := flag.Int("cores", runtime.NumCPU(), "CPU cores to use")
	flag.Parse()

	var err error
	if *inputFlag == "" {
		err = fmt.Errorf("Must specifiy --input")
	} else if *outputFlag == "" {
		err = fmt.Errorf("Must specifiy --output")
	} else {
		var postcodes Postcodes
		if postcodes, err = readCodepointOpen(*inputFlag); err == nil {
			config := compact.Options{
				OutputFilename:       *outputFlag,
				Cores:                *coresFlag,
				WorkDirectory:        "",
				PointsWorkOutputType: compact.OutputTypeMemory,
			}
			err = compact.Build(&postcodes, &config)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
