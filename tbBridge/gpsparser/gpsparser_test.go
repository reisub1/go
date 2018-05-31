package gpsparser

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type testpair struct {
	input    string
	expected []GPSParsed
}

var tests = []testpair{
	//Empty message
	{
		input:    "",
		expected: nil,
	},
	//Valid message with NE
	{
		input: "GTPL $1,867322035135813,A,290518,062804,18.709738,N,80.068397,E,0,406,309,11,0,14,1,0,26.4470#",
		expected: []GPSParsed{
			{
				Protocol:  "AIS140",
				Uniqid:    "867322035135813",
				TS_Millis: 1527575284000,
				ActualLat: 18.709738,
				ActualLng: 80.068397,
			},
		},
	},
	//Valid message with SW
	{
		input: "GTPL $1,867322035135813,A,290518,062804,18.709738,S,80.068397,W,0,406,309,11,0,14,1,0,26.4470#",
		expected: []GPSParsed{
			{
				Protocol:  "AIS140",
				Uniqid:    "867322035135813",
				TS_Millis: 1527575284000,
				ActualLat: -18.709738,
				ActualLng: -80.068397,
			},
		},
	},
	//No commas
	{
		input:    "GTPL $1867322035135813A29051806280418.709738N80.068397E0406309110141026.4470#",
		expected: nil,
	},
}

func TestParse(t *testing.T) {
	assert := assert.New(t)
	c := make(chan *GPSParsed, 10)
	for _, testCase := range tests {
		Parse(&testCase.input, c)
		// i := 0
		// for output := range c {
		// 	assert.Equal(testCase.expected[i].Uniqid, output.Uniqid)

		// 	assert.Equal(testCase.expected[i].TS_Millis, output.TS_Millis)

		// 	assert.Equal(testCase.expected[i].ActualLat, output.ActualLat)
		// 	assert.Equal(testCase.expected[i].ActualLng, output.ActualLng)
		// }
		for _, expOutput := range testCase.expected {
			select {
			case output := <-c:
				assert.Equal(expOutput.Uniqid, output.Uniqid)

				assert.Equal(expOutput.TS_Millis, output.TS_Millis)

				assert.Equal(expOutput.ActualLat, output.ActualLat)
				assert.Equal(expOutput.ActualLng, output.ActualLng)
			case <-time.After(10 * time.Millisecond):
				assert.Nil(expOutput)
			}
		}
	}
}

// var benchoutput *WTD
// var bencherr error

// func BenchmarkParse(b *testing.B) {
// 	var wtdo *WTD
// 	var eo error
// 	for n := 0; n < b.N; n++ {
// 		for _, testCase := range tests {
// 			output, outputErr := wtd.Parse(testCase.input)
// 			wtdo, eo = output, outputErr
// 		}
// 	}
// 	benchoutput, bencherr = wtdo, eo
// }

type Bench []string

var benchmarks = Bench{
	"GTPL $1,867322035135813,A,290518,062804,18.709738,N,80.068397,E,0,406,309,11,0,14,1,0,26.4470#",
	"GTPL $1,867322035135813,A,290518,062804,18.709738,S,80.068397,W,0,406,309,11,0,14,1,0,26.4470#",
	"GTPL $1867322035135813A29051806280418.709738N80.068397E0406309110141026.4470#",
}

func BenchmarkParse(b *testing.B) {
	c := make(chan *GPSParsed, 1000)
	go ChanSinker(c)
	for i := 0; i < b.N; i++ {
		for _, input := range benchmarks {
			Parse(&input, c)
		}
	}
}

func ChanSinker(c chan *GPSParsed) {
	for {
		_ = <-c
	}
}
