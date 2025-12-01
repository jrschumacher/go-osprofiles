package profiles

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jrschumacher/go-osprofiles/internal/global"
	"github.com/stretchr/testify/suite"
	"github.com/zalando/go-keyring"
)

type ProfilesSuite struct {
	suite.Suite

	keyringProfiler *Profiler

	fileSystemProfiler *Profiler
	testTempDir        string
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

// TODO: integration test profile lifecycle for in-memory

func (s *ProfilesSuite) SetupSuite() {
	keyringProfiler, err := New(testConsumerServiceProfiler, WithKeyringStore())
	s.Require().NoError(err)
	s.Require().NotNil(keyringProfiler)
	s.keyringProfiler = keyringProfiler

	// Set up a temporary profiles directory
	tempDir := os.TempDir()
	dirPath := filepath.Join(tempDir, "profiles")
	err = os.MkdirAll(dirPath, os.ModePerm)
	s.Require().NoError(err)
	s.testTempDir = dirPath

	fileSystemProfiler, err := New(testConsumerServiceProfiler, WithFileStore(s.testTempDir))
	s.Require().NoError(err)
	s.Require().NotNil(fileSystemProfiler)
	s.fileSystemProfiler = fileSystemProfiler
}

func (s *ProfilesSuite) TearDownSuite() {
	// Remove all keyring entries added by the test suite
	//nolint:errcheck // teardown error not relevant
	keyring.DeleteAll(testConsumerServiceProfiler)

	// Remove all stored profiles set to the temp directory
	//nolint:errcheck // teardown error not relevant
	os.RemoveAll(s.testTempDir)
}

func (s *ProfilesSuite) assertDirFileCount(dir string, expected int) {
	files, err := os.ReadDir(dir)
	s.Require().NoError(err)
	s.Require().Len(files, expected)
}

func (s *ProfilesSuite) assertKeyringProfiles(shouldBeDeleted bool, names ...string) {
	for _, name := range names {
		key := name
		if name != global.STORE_KEY_GLOBAL {
			key = getStoreKey(name)
		}
		_, err := keyring.Get(testConsumerServiceProfiler, key)
		if shouldBeDeleted {
			s.Require().Error(err)
			s.Require().ErrorIs(err, keyring.ErrNotFound)
		} else {
			s.Require().Nil(err)
			s.Require().NoError(err)
		}
	}
}

func (s *ProfilesSuite) TestHasGlobalStore_FileStore() {
	configName := "test-has-global-store-fs"

	exists, err := HasGlobalStore(configName, WithFileStore(s.testTempDir))
	s.Require().NoError(err)
	s.Require().False(exists)

	profiler, err := New(configName, WithFileStore(s.testTempDir))
	s.Require().NoError(err)
	s.Require().NotNil(profiler)

	exists, err = HasGlobalStore(configName, WithFileStore(s.testTempDir))
	s.Require().NoError(err)
	s.Require().True(exists)
	s.Require().NoError(profiler.Cleanup(true))
}

func (s *ProfilesSuite) TestHasGlobalStore_Keyring() {
	configName := "test-has-global-store-keyring"

	exists, err := HasGlobalStore(configName, WithKeyringStore())
	s.Require().NoError(err)
	s.Require().False(exists)

	profiler, err := New(configName, WithKeyringStore())
	s.Require().NoError(err)
	s.Require().NotNil(profiler)

	exists, err = HasGlobalStore(configName, WithKeyringStore())
	s.Require().NoError(err)
	s.Require().True(exists)
	s.Require().NoError(profiler.Cleanup(true))
}

func (s *ProfilesSuite) TestLifecycleProfile_FileStore() {
	profile := &mockProfile{
		Name:      "test-profile-fs",
		TestValue: "test-value-fs",
		Nested: struct {
			SubValue int `json:"sub_value"`
		}{1},
	}
	fileSystemProfiler := s.fileSystemProfiler

	// no profiles
	list := ListProfiles(fileSystemProfiler)
	s.Require().Len(list, 0)

	// add a test profile
	s.Require().NoError(fileSystemProfiler.AddProfile(profile, true))

	// ensure new profile created
	list = ListProfiles(fileSystemProfiler)
	s.Require().Len(list, 1)
	s.Require().Equal(list[0], profile.Name)

	// ensure profile exists
	p, err := GetProfile[*mockProfile](fileSystemProfiler, profile.Name)
	s.Require().NoError(err)
	s.Require().NotNil(p)
	s.Require().NotNil(p.Profile)
	s.Require().Equal(profile.Name, p.Profile.GetName())
	s.Require().Equal(profile.TestValue, p.Profile.(*mockProfile).TestValue)
	s.Require().Equal(profile.Nested.SubValue, p.Profile.(*mockProfile).Nested.SubValue)

	// check the file system
	s.assertDirFileCount(s.testTempDir, 4)

	// test conflict if creating same profile name twice
	err = fileSystemProfiler.AddProfile(profile, true)
	s.Require().ErrorIs(err, ErrProfileNameConflict)

	// update it
	profile.TestValue = "test-value-updated-123"
	s.Require().NoError(UpdateCurrentProfile(fileSystemProfiler, profile))

	// get it again
	p, err = GetProfile[*mockProfile](fileSystemProfiler, profile.Name)
	s.Require().NoError(err)
	s.Require().NotNil(p)
	s.Require().NotNil(p.Profile)
	s.Require().Equal(profile.Name, p.Profile.GetName())
	// updated successfully
	s.Require().Equal(profile.TestValue, p.Profile.(*mockProfile).TestValue)

	// delete fails due to being default
	err = DeleteProfile[*mockProfile](fileSystemProfiler, profile.Name)
	s.Require().ErrorIs(err, ErrCannotDeleteDefaultProfile)

	// add a second profile
	profile2 := &mockProfile{
		Name: "test-profile-2-abc",
	}
	s.Require().NoError(fileSystemProfiler.AddProfile(profile2, false))

	// find both in the list
	list = ListProfiles(fileSystemProfiler)
	s.Require().Len(list, 2)
	s.Require().Equal(list[0], profile.Name)
	s.Require().Equal(list[1], profile2.Name)

	// check the file system
	s.assertDirFileCount(s.testTempDir, 6)

	// set the second profile as default
	s.Require().NoError(SetDefaultProfile(fileSystemProfiler, profile2.Name))
	// delete the first profile
	s.Require().NoError(DeleteProfile[*mockProfile](fileSystemProfiler, profile.Name))

	// ensure the first profile is gone
	list = ListProfiles(fileSystemProfiler)
	s.Require().Len(list, 1)
	s.Require().Equal(list[0], profile2.Name)

	s.assertDirFileCount(s.testTempDir, 4)

	s.Require().NoError(fileSystemProfiler.DeleteAllProfiles())
	list = ListProfiles(fileSystemProfiler)
	s.Require().Len(list, 0)
	s.assertDirFileCount(s.testTempDir, 2)

	// delete all remaining profiles
	s.Require().NoError(fileSystemProfiler.Cleanup(true))
	s.Require().Nil(fileSystemProfiler.globalStore)
	s.Require().Nil(fileSystemProfiler.currentProfileStore)

	s.assertDirFileCount(s.testTempDir, 0)
}

func (s *ProfilesSuite) TestLifecycleProfile_Keyring() {
	profile := &mockProfile{
		Name:      "test-profile",
		TestValue: "test-value",
		Nested: struct {
			SubValue int `json:"sub_value"`
		}{1},
	}
	keyringProfiler := s.keyringProfiler

	// no profiles
	list := ListProfiles(keyringProfiler)
	s.Require().Len(list, 0)

	// add a test profile
	s.Require().NoError(keyringProfiler.AddProfile(profile, true))

	// ensure new profile created
	list = ListProfiles(keyringProfiler)
	s.Require().Len(list, 1)
	s.Require().Equal(list[0], profile.Name)

	// ensure profile exists
	p, err := GetProfile[*mockProfile](keyringProfiler, profile.Name)
	s.Require().NoError(err)
	s.Require().NotNil(p)
	s.Require().NotNil(p.Profile)
	s.Require().Equal(profile.Name, p.Profile.GetName())
	s.Require().Equal(profile.TestValue, p.Profile.(*mockProfile).TestValue)
	s.Require().Equal(profile.Nested.SubValue, p.Profile.(*mockProfile).Nested.SubValue)

	// test conflict if creating same profile name twice
	err = keyringProfiler.AddProfile(profile, true)
	s.Require().ErrorIs(err, ErrProfileNameConflict)

	// update it
	profile.TestValue = "test-value-updated"
	s.Require().NoError(UpdateCurrentProfile(keyringProfiler, profile))

	// get it again
	p, err = GetProfile[*mockProfile](keyringProfiler, profile.Name)
	s.Require().NoError(err)
	s.Require().NotNil(p)
	s.Require().NotNil(p.Profile)
	s.Require().Equal(profile.Name, p.Profile.GetName())
	// updated successfully
	s.Require().Equal(profile.TestValue, p.Profile.(*mockProfile).TestValue)

	// delete fails due to being default
	err = DeleteProfile[*mockProfile](keyringProfiler, profile.Name)
	s.Require().ErrorIs(err, ErrCannotDeleteDefaultProfile)

	// add a second profile
	profile2 := &mockProfile{
		Name: "test-profile-2",
	}
	s.Require().NoError(keyringProfiler.AddProfile(profile2, false))

	// find both in the list
	list = ListProfiles(keyringProfiler)
	s.Require().Len(list, 2)
	s.Require().Equal(list[0], profile.Name)
	s.Require().Equal(list[1], profile2.Name)

	// set the second profile as default
	s.Require().NoError(SetDefaultProfile(keyringProfiler, profile2.Name))
	// delete the first profile
	s.Require().NoError(DeleteProfile[*mockProfile](keyringProfiler, profile.Name))

	// ensure the first profile is gone
	list = ListProfiles(keyringProfiler)
	s.Require().Len(list, 1)
	s.Require().Equal(list[0], profile2.Name)

	s.Require().NoError(keyringProfiler.DeleteAllProfiles())
	list = ListProfiles(keyringProfiler)
	s.Require().Len(list, 0)
	s.assertKeyringProfiles(false, global.STORE_KEY_GLOBAL)

	// delete all remaining profiles
	s.Require().NoError(keyringProfiler.Cleanup(true))
	s.Require().Nil(keyringProfiler.globalStore)
	s.Require().Nil(keyringProfiler.currentProfileStore)
	s.assertKeyringProfiles(true, global.STORE_KEY_GLOBAL, profile.Name, profile2.Name)
}

func TestAttributesSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping profiles test suite")
	}
	suite.Run(t, new(ProfilesSuite))
}
