package wtd

import (
	// "github.com/reisub1/go/tbBridge/wtd"
	"github.com/stretchr/testify/assert"
	"testing"
	// "time"
	"errors"
)

type testpair struct {
	input       string
	expected    WTD
	expectedErr error
}

var tests = []testpair{
	//Empty message
	{"*ZJ#",
		WTD{},
		errors.New("Empty message")},
	//No commas
	{"*ZJ2030295119V1134310A3106.3677N7710.9352E2.380.0012051800000000#",
		WTD{},
		errors.New("Not a CSV message: *ZJ2030295119V1134310A3106.3677N7710.9352E2.380.0012051800000000#")},
	//Not 13 fields
	{"*ZJ,2030295119V1134310A3106.3677N7710.9352E2.380.0012051800000000#",
		WTD{},
		errors.New("Incorrect number of fields in CSV: 2")},
	//Non float latitude
	{"*ZJ,2030295119,V1,134310,A,3106.3677.12,S,7710.9352,W,2.38,0.00,120518,00000000#",
		WTD{},
		errors.New("Parsing error for latitude 3106.3677.12")},
	//Non float longitude
	{"*ZJ,2030295119,V1,134310,A,3106.3677,S,7710.9352.12,W,2.38,0.00,120518,00000000#",
		WTD{},
		errors.New("Parsing error for longitude 7710.9352.12")},
	//Valid message
	{"*ZJ,2030295119,V1,134310,A,3106.3677,N,7710.9352,E,2.38,0.00,120518,00000000#",
		WTD{[]string{}, "2030295119", 1526132590000, 31.063677, 77.109352},
		nil},
	//Valid message with SW
	{"*ZJ,2030295119,V1,134310,A,3106.3677,S,7710.9352,W,2.38,0.00,120518,00000000#",
		WTD{[]string{}, "2030295119", 1526132590000, -31.063677, -77.109352},
		nil},
}

func TestParse(t *testing.T) {
	assert := assert.New(t)
	for _, testCase := range tests {
		output, outputErr := Parse(testCase.input)
		assert.Equal(testCase.expectedErr, outputErr)
		if testCase.expectedErr == nil {
			assert.Equal(testCase.expected.Uniqid, output.Uniqid)

			assert.Equal(testCase.expected.TS_Millis, output.TS_Millis)

			assert.Equal(testCase.expected.ActualLat, output.ActualLat)
			assert.Equal(testCase.expected.ActualLng, output.ActualLng)
		}
	}
}

var benchoutput *WTD
var bencherr error

func BenchmarkParse(b *testing.B) {
	var wtdo *WTD
	var eo error
	for n := 0; n < b.N; n++ {
		for _, testCase := range tests {
			output, outputErr := Parse(testCase.input)
			wtdo, eo = output, outputErr
		}
	}
	benchoutput, bencherr = wtdo, eo
}
