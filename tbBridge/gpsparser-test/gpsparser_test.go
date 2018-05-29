package gpsparser

import (
	"github.com/reisub1/go/tbBridge/gpsparser"
	"github.com/stretchr/testify/assert"
	"testing"
	// "time"
)

type GPSParsed = gpsparser.GPSParsed

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
	//No commas
	// {"*ZJ2030295119V1134310A3106.3677N7710.9352E2.380.0012051800000000#",
	// 	WTD{},
	// 	errors.New("Not a CSV message: *ZJ2030295119V1134310A3106.3677N7710.9352E2.380.0012051800000000#")},
	//Not 13 fields
	// {"*ZJ,2030295119V1134310A3106.3677N7710.9352E2.380.0012051800000000#",
	// 	WTD{},
	// 	errors.New("Incorrect number of fields in CSV: 2")},
	//Non float latitude
	// {"*ZJ,2030295119,V1,134310,A,3106.3677.12,S,7710.9352,W,2.38,0.00,120518,00000000#",
	// 	WTD{},
	// 	errors.New("Parsing error for latitude 3106.3677.12")},
	//Non float longitude
	// {"*ZJ,2030295119,V1,134310,A,3106.3677,S,7710.9352.12,W,2.38,0.00,120518,00000000#",
	// 	WTD{},
	// 	errors.New("Parsing error for longitude 7710.9352.12")},
	//Valid message
	//{"*ZJ,2030295119,V1,134310,A,3106.3677,N,7710.9352,E,2.38,0.00,120518,00000000#",
	//	WTD{[]string{}, "2030295119", 1526132590000, 31.063677, 77.109352},
	//	nil},
	////Valid message with SW
	//{"*ZJ,2030295119,V1,134310,A,3106.3677,S,7710.9352,W,2.38,0.00,120518,00000000#",
	//	WTD{[]string{}, "2030295119", 1526132590000, -31.063677, -77.109352},
	//	nil},
}

func TestParse(t *testing.T) {
	assert := assert.New(t)
	for _, testCase := range tests {
		c := make(chan *GPSParsed, 10)
		gpsparser.Parse(&testCase.input, c)
		for _, expOutput := range testCase.expected {
			output := <-c

			assert.Equal(expOutput.Uniqid, output.Uniqid)

			assert.Equal(expOutput.TS_Millis, output.TS_Millis)

			assert.Equal(expOutput.ActualLat, output.ActualLat)
			assert.Equal(expOutput.ActualLng, output.ActualLng)
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
