package ingest

import (
	"fmt"
	"strconv"
	"strings"

	"diagonal.works/b6"
	"github.com/golang/geo/s2"
)

func ParseBoundsFlags(box string, ll string, radius string) (s2.Region, error) {
	rect, err := ParseBoundingBox(box)
	if err != nil {
		return nil, err
	}
	if !rect.IsFull() {
		return rect, nil
	}
	return ParseCap(ll, radius)
}

func ParseBoundingBox(box string) (s2.Rect, error) {
	if box == "" {
		return s2.FullRect(), nil
	}
	rect := s2.EmptyRect()
	fs := [4]float64{}
	if err := parseFloats(box, fs[0:]); err != nil {
		return rect, fmt.Errorf("lat,lng,lat,lng: %s", err)
	}
	rect = rect.AddPoint(s2.LatLngFromDegrees(fs[0], fs[1]))
	return rect.AddPoint(s2.LatLngFromDegrees(fs[2], fs[3])), nil
}

func ParseCap(ll string, radius string) (s2.Cap, error) {
	if ll == "" || radius == "" {
		return s2.FullCap(), nil
	}
	fs := [2]float64{}
	cap := s2.EmptyCap()
	if err := parseFloats(ll, fs[0:]); err != nil {
		return cap, fmt.Errorf("cap lat,lng: %s", err)
	}
	r := 0.0
	var err error
	if r, err = strconv.ParseFloat(radius, 64); err != nil {
		return cap, fmt.Errorf("cap radius: %s", err)
	}
	return s2.CapFromCenterAngle(s2.PointFromLatLng(s2.LatLngFromDegrees(fs[0], fs[1])), b6.MetersToAngle(r)), nil
}

func parseFloats(s string, fs []float64) error {
	ss := strings.Split(s, ",")
	ok := false
	var err error
	if len(ss) == len(fs) {
		for i, float := range ss {
			fs[i], err = strconv.ParseFloat(float, 64)
			if err != nil {
				break
			}
			ok = i == len(fs)-1
		}
	}
	if !ok {
		err = fmt.Errorf("expected %d floats, found %d", len(fs), len(ss))
	}
	return err
}
