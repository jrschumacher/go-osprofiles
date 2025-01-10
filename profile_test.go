package profiles

import (
	"testing"

	"github.com/jrschumacher/go-osprofiles/internal/global"
	"github.com/stretchr/testify/suite"
	"github.com/zalando/go-keyring"
)

type ProfilesSuite struct {
	suite.Suite
	p *Profiler
}

type mockProfile struct {
	Name      string `json:"name"`
	TestValue string `json:"test_value"`
	Nested    struct {
		SubValue int `json:"sub_value"`
	}
}

func (p *mockProfile) GetName() string {
	return p.Name
}

const testConsumerServiceProfiler = "test-consumer-service-profiler"

func (s *ProfilesSuite) SetupSuite() {
	profiler, err := New(testConsumerServiceProfiler, WithKeyringStore())
	s.Require().NoError(err)
	s.Require().NotNil(profiler)

	s.p = profiler
}

func (s *ProfilesSuite) TearDownSuite() {
	// Remove all keyring entries added by the test suite
	//nolint:errcheck // teardown error not relevant
	keyring.DeleteAll(testConsumerServiceProfiler)
}

// TODO: integration test profile lifecycle for other store types

func (s *ProfilesSuite) TestLifecycleProfile_Keyring() {
	profile := &mockProfile{
		Name:      "test-profile",
		TestValue: "test-value",
		Nested: struct {
			SubValue int `json:"sub_value"`
		}{1},
	}

	// no profiles
	list := ListProfiles(s.p)
	s.Require().Len(list, 0)

	// add a test profile
	s.Require().NoError(s.p.AddProfile(profile, true))

	// ensure new profile created
	list = ListProfiles(s.p)
	s.Require().Len(list, 1)
	s.Require().Equal(list[0], profile.Name)

	// ensure profile exists
	p, err := GetProfile[*mockProfile](s.p, profile.Name)
	s.Require().NoError(err)
	s.Require().NotNil(p)
	s.Require().NotNil(p.Profile)
	s.Require().Equal(profile.Name, p.Profile.GetName())
	s.Require().Equal(profile.TestValue, p.Profile.(*mockProfile).TestValue)
	s.Require().Equal(profile.Nested.SubValue, p.Profile.(*mockProfile).Nested.SubValue)

	// update it
	profile.TestValue = "test-value-updated"
	s.Require().NoError(UpdateCurrentProfile(s.p, profile))

	// get it again
	p, err = GetProfile[*mockProfile](s.p, profile.Name)
	s.Require().NoError(err)
	s.Require().NotNil(p)
	s.Require().NotNil(p.Profile)
	s.Require().Equal(profile.Name, p.Profile.GetName())
	// updated successfully
	s.Require().Equal(profile.TestValue, p.Profile.(*mockProfile).TestValue)

	// delete fails due to being default
	err = DeleteProfile[*mockProfile](s.p, profile.Name)
	s.Require().ErrorIs(err, global.ErrDeletingDefaultProfile)

	// add a second profile
	profile2 := &mockProfile{
		Name: "test-profile-2",
	}
	s.Require().NoError(s.p.AddProfile(profile2, false))

	// find both in the list
	list = ListProfiles(s.p)
	s.Require().Len(list, 2)
	s.Require().Equal(list[0], profile.Name)
	s.Require().Equal(list[1], profile2.Name)

	// set the second profile as default
	s.Require().NoError(SetDefaultProfile(s.p, profile2.Name))
	// delete the first profile
	s.Require().NoError(DeleteProfile[*mockProfile](s.p, profile.Name))

	// ensure the first profile is gone
	list = ListProfiles(s.p)
	s.Require().Len(list, 1)
	s.Require().Equal(list[0], profile2.Name)
}

func TestAttributesSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping profiles test suite")
	}
	suite.Run(t, new(ProfilesSuite))
}
