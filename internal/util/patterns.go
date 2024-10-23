package util

import "regexp"

var NameWithShortSeasonAndEpisode = regexp.MustCompile(`^.*[sS](?P<series>\d+)(\s+)?[eE](?P<episode>\d+).*$`)
var NameWithLongSeasonAndEpisode = regexp.MustCompile(`^.*[sS](eason|eries) (?P<series>\d+) [eE]pisode (?P<episode>\d+).*$`)
var NameWithSplitSeasonAndEpisode = regexp.MustCompile(`^.*[sS](?P<series>\d+)\.[eE](?P<episode>\d+).*$`)
var ShortSeasonAndEpisode = regexp.MustCompile(`^\D*(?P<series>\d+)[xX](?P<episode>\d+)\D*$`)
