package main

type ZappedFile struct {
	Hash string
	Path string
	Size int64
}

//
//func GetAllZappedFilesImproved(zapPath string) iter.Seq2[*ZappedFile, error] {
//	return func(yield func(*ZappedFile, error) bool) {
//		for prefix := range GenerateZapHashDirectoryPrefixes {
//			basePath := path.Join(zapPath, prefix)
//			entries, err := os.ReadDir(basePath)
//
//			if err != nil {
//				yield(nil, err)
//				return
//			}
//
//			for _, e := range entries {
//				filePath := path.Join(basePath, e.Name())
//
//				if e.IsDir() {
//					yield(nil, errors.New(fmt.Sprintf("a subdirectory was unexpected \"%s\" in the Zap path", filePath)))
//					return
//				}
//
//				// todo: time this with and without call to .Info()!, see if it's the same time if so we keep this
//
//				fileInfo, fileInforErr := e.Info()
//
//				if fileInforErr != nil {
//					yield(nil, fileInforErr)
//					return
//				}
//
//				hash := strings.ReplaceAll(prefix, fmt.Sprintf("%c", os.PathSeparator), "") + e.Name()
//
//				yield(&ZappedFile{
//					Hash: hash,
//					Path: filePath,
//					Size: fileInfo.Size(),
//				}, nil)
//			}
//		}
//	}
//}
//
//func GenerateZapHashDirectoryPrefixes(yield func(string) bool) {
//	for level1 := 0; level1 < 0x100; level1++ {
//		for level2 := 0; level2 < 0x100; level2++ {
//			yield(fmt.Sprintf("%02x%c%02x", level1, os.PathSeparator, level2))
//		}
//	}
//}
