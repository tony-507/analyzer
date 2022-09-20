package avContainer

import (
	"errors"

	"github.com/tony-507/analyzers/test/testUtils"
)

func readPATTest() testUtils.Testcase {
	tc := testUtils.GetTestCase("readPATTest")

	tc.Describe("Initialization", func(input interface{}) (interface{}, error) {
		dummyPAT := []byte{0x00, 0x00, 0xB0, 0x0D, 0x11, 0x11, 0xC1,
			0x00, 0x00, 0x00, 0x0A, 0xE1, 0x02, 0xAA, 0x4A, 0xE2, 0xD2}
		return dummyPAT, nil
	})

	tc.Describe("Parse PAT", func(input interface{}) (interface{}, error) {
		dummyPAT, isBts := input.([]byte)
		if !isBts {
			err := errors.New("Input not passed to next step")
			return nil, err
		}
		PAT, parsingErr := ParsePAT(dummyPAT, 0)
		if parsingErr != nil {
			return nil, parsingErr
		}

		var err error
		err = testUtils.Assert_equal(PAT.tableId, 0)
		if err != nil {
			return nil, err
		}

		err = testUtils.Assert_equal(PAT.tableIdExt, 4369)
		if err != nil {
			return nil, err
		}

		err = testUtils.Assert_equal(PAT.Version, 0)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	return tc
}

func AddUnitTestSuite(t *testUtils.Tester) {
	tmg := testUtils.GetTestCaseMgr()

	// We may add custom test filter here later

	// common
	tmg.AddTest(readPATTest, []string{"avContainer"})

	t.AddSuite("unitTest", tmg)
}
