package main

//
//func TestGetAllFilesRelativeToRootPath(t *testing.T) {
//	ctx := &Context{
//		Config: &config.Config{},
//		DB:     testDB(),
//	}
//
//	err := ctx.AddRootPath(testDataPath)
//	assert.NoError(t, err)
//
//	err = ctx.Crawl()
//	assert.NoError(t, err)
//}
//
//
//
//func TestAddRootGivenAFilePassedShouldReturnError(t *testing.T) {
//	ctx := &Context{
//		Config: &config.Config{},
//	}
//
//	err := ctx.AddRootPath(path.Join(testDataPath, "a/b/j.txt"))
//
//	assert.ErrorIs(t, err, ErrCouldNotResolvePath)
//}
//
//func TestAddRootGivenAnInvalidPathReturnError(t *testing.T) {
//	ctx := &Context{
//		Config: &config.Config{},
//	}
//
//	err := ctx.AddRootPath(path.Join(testDataPath, "fail/"))
//
//	assert.ErrorIs(t, err, ErrCouldNotResolvePath)
//}
